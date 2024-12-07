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

	dialCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	c.serverConn, err = dialer.DialContext(dialCtx, "tcp", host)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("server has already shutdown")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("timeout exceeded when dialling server")
		}
		return err
	}

	fmt.Printf("connecting to server at %s...\n", host)
	return nil
}

func (c *client) ListenForMessages(errChan chan<- error) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		fmt.Println("cleaning up...")
		c.startCleanup(ctx)
		cancel()
		fmt.Println("done cleaning")
		close(errChan)
	}()

	go c.handleSignals()
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		for {
			// Wrap read in a select to check exitChan
			readErrChan := make(chan error, 1)
			readDataChan := make(chan []byte, 1)

			go func() {
				read, err := c.readFromServer(ctx)
				if err != nil {
					readErrChan <- err
					return
				}
				readDataChan <- read
			}()

			select {
			case <-c.exitChan:
				return
			case err := <-readErrChan:
				if errors.Is(err, utils.ErrCtxTimeOut) {
					continue
				}
				errChan <- err
				return
			case read := <-readDataChan:
				if err := c.processPayload(ctx, read); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()
	<-done
}

func (c *client) processPayload(ctx context.Context, data []byte) error {
	payload, err := base.DecodePayload(data)
	if err != nil {
		return err
	}

	switch payload.GetHeader() {
	case common.Header_HEADER_ERROR:
		errPayload, ok := payload.GetContent().(*base.Payload_Error)
		if ok {
			return c.handleErrPayload(*errPayload)
		}

	case common.Header_HEADER_AUTH:
		// we only need to check its type but we won't use the actual payload
		_, ok := payload.GetContent().(*base.Payload_Auth)
		if ok {
			return c.handleAuthPayload(ctx)
		}

	case common.Header_HEADER_INFO:
		infoPayload, ok := payload.GetContent().(*base.Payload_Info)
		if ok {
			return c.handleInfoPayload(ctx, *infoPayload)
		}

	case common.Header_HEADER_TERMINAL_DATA:
		termPayload, ok := payload.GetContent().(*base.Payload_TermContent)
		if ok {
			return c.handleTermPayload(*termPayload)
		}

	case common.Header_HEADER_HEARTBEAT:
		_, ok := payload.GetContent().(*base.Payload_Heartbeat)
		if ok {
			return c.handleHeartbeatPayload()
		}
	default:
		fmt.Print(string(data))
	}

	return utils.ErrUnspecifiedPayload
}
