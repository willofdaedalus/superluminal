package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/utils"
)

const (
	maxConnTries = 3
)

type client struct {
	TermContent chan string
	name        string
	joined      time.Time
	serverConn  net.Conn
	exitChan    chan struct{}
	sigChan     chan os.Signal
	sentPass    bool
	isApproved  bool
	mu          sync.Mutex
	tracker     *utils.SyncTracker
}

func New(name string) *client {
	return &client{
		name:        name,
		joined:      time.Now(),
		TermContent: make(chan string, 1),
		exitChan:    make(chan struct{}, 1),
		sigChan:     make(chan os.Signal, 1),
		sentPass:    false,
		isApproved:  false,
		serverConn:  nil,
		tracker:     utils.NewSyncTracker(),
	}
}

func (c *client) ConnectToSession(ctx context.Context, host string) error {
	var dialer net.Dialer
	var err error

	log.Printf("Attempting to connect to host: %s", host)

	dialCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	c.serverConn, err = dialer.DialContext(dialCtx, "tcp", host)
	if err != nil {
		log.Printf("Connection FAILED - detailed error: %+v", err)

		switch {
		case errors.Is(err, io.EOF):
			log.Println("Server appears to have shutdown")
		case errors.Is(err, context.DeadlineExceeded):
			log.Println("Connection attempt timed out")
		}

		return err
	}

	// Add more connection verification
	log.Printf("Connection SUCCESSFUL to %s\n", host)
	log.Printf("Local Address: %v\n", c.serverConn.LocalAddr())
	log.Printf("Remote Address: %v\n", c.serverConn.RemoteAddr())

	return nil
}

func (c *client) ListenForMessages(errChan chan<- error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		c.startCleanup(ctx)
		cancel()
		close(errChan)
	}()

	go c.handleSignals(ctx)

	readErr := make(chan error, 1)
	readData := make(chan []byte, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer func() {
			close(readErr)
			close(readData)
			wg.Done()
			fmt.Println("exiting from read goroutine")
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
						fmt.Println("dropped message: readDataChan is full")
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
			log.Println("exiting from exitChan channel")
			cancel()
			fmt.Println("waiting for wg...")
			wg.Wait()
			fmt.Println("wg is done")
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

func (c *client) processPayload(ctx context.Context, data []byte, errChan chan<- error) {
	procCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	payload, err := base.DecodePayload(data)
	if err != nil {
		fmt.Println(len(data))
		fmt.Printf("what the fuck %v", err)
		errChan <- err
		return
	}

	switch payload.GetHeader() {
	case common.Header_HEADER_ERROR:
		errPayload, ok := payload.GetContent().(*base.Payload_Error)
		if ok {
			log.Println("got a header error")
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
