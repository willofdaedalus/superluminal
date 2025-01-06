package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/utils"
)

const (
	maxConnTries = 3
)

type Client struct {
	TermContent chan string
	name        string
	joined      time.Time
	serverConn  net.Conn
	signals     []os.Signal
	exitChan    chan struct{}
	sigChan     chan os.Signal
	// channel for bubbletea ui to send the password to the backend
	bbltPass   chan string
	SentPass   bool
	isApproved bool
	mu         sync.Mutex
	tracker    *utils.SyncTracker
}

func New(name string) *Client {
	return &Client{
		name:        name,
		joined:      time.Now(),
		TermContent: make(chan string, 1),
		exitChan:    make(chan struct{}, 1),
		sigChan:     make(chan os.Signal, 1),
		signals:     []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		SentPass:    false,
		isApproved:  false,
		bbltPass:    make(chan string, 1),
		serverConn:  nil,
		tracker:     utils.NewSyncTracker(),
	}
}

func (c *Client) ConnectToSession(host string) error {
	var dialer net.Dialer
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	c.serverConn, err = dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			log.Println("Server appears to have shutdown")
		case errors.Is(err, context.DeadlineExceeded):
			log.Println("Connection attempt timed out")
		}

		return err
	}

	// // Add more connection verification
	// log.Printf("Connection SUCCESSFUL to %s\n", host)
	// log.Printf("Local Address: %v\n", c.serverConn.LocalAddr())
	// log.Printf("Remote Address: %v\n", c.serverConn.RemoteAddr())

	return nil
}

func (c *Client) ListenForMessages(errChan chan<- error) {
	ctx, cancel := signal.NotifyContext(context.Background(), c.signals...)
	defer func() {
		// cleanup has its own context timeout
		c.startCleanup()
		close(errChan)
	}()

	readErr := make(chan error, 1)
	readData := make(chan []byte, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer func() {
			close(readErr)
			close(readData)
			wg.Done()
		}()

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			// check context cancellation first, before the select
			if ctx.Err() != nil {
				return
			}

			// use a select with a default case to prevent blocking
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// attempt to read data with a timeout
				read, err := utils.ReadFull(ctx, c.serverConn, c.tracker)
				if err != nil {
					select {
					case readErr <- err:
					default:
					}
					return
				}
				if read != nil {
					select {
					case readData <- read:
					default:
						fmt.Println("dropped message: readData is full")
					}
				}
			default:
				// allow immediate context cancellation check
				time.Sleep(10 * time.Millisecond)
			}
		}
	}(&wg)

	// main loop
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case <-c.exitChan:
			cancel()
			wg.Wait()
			return
		case err := <-readErr:
			if err != nil {
				log.Println("critical error: ", err)
				select {
				case errChan <- err:
				default:
					log.Println("error channel blocked, dropping error")
				}
				cancel()
				wg.Wait()
				return
			}
		case read := <-readData:
			go c.processPayload(ctx, read, errChan)
		}
	}
}

func (c *Client) processPayload(ctx context.Context, data []byte, errChan chan<- error) {
	procCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	payload, err := base.DecodePayload(data)
	if err != nil {
		fmt.Println(len(data))
		errChan <- err
		return
	}

	switch payload.GetHeader() {
	case common.Header_HEADER_ERROR:
		errPayload, ok := payload.GetContent().(*base.Payload_Error)
		if ok {
			errChan <- c.handleErrPayload(*errPayload)
			return
		}

	case common.Header_HEADER_AUTH:
		// we only need to check its type but we won't use the actual payload
		_, ok := payload.GetContent().(*base.Payload_Auth)
		if ok {
			errChan <- c.handleAuthPayload(procCtx)
			return
		}

	case common.Header_HEADER_INFO:
		infoPayload, ok := payload.GetContent().(*base.Payload_Info)
		if ok {
			errChan <- c.handleInfoPayload(procCtx, *infoPayload)
			return
		}

	case common.Header_HEADER_TERMINAL_DATA:
		termPayload, ok := payload.GetContent().(*base.Payload_TermContent)
		if ok {
			errChan <- c.handleTermPayload(*termPayload)
			return
		}

	case common.Header_HEADER_HEARTBEAT:
		_, ok := payload.GetContent().(*base.Payload_Heartbeat)
		if ok {
			errChan <- c.handleHeartbeatPayload()
			return
		}
	default:
		// temporary solution
		fmt.Print(string(data))
	}

	errChan <- utils.ErrUnspecifiedPayload
	return
}
