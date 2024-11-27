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

func (c *client) ConnectToSession(ctx context.Context, host, port string) error {
	var dialer net.Dialer
	var err error

	dialCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	c.serverConn, err = dialer.DialContext(dialCtx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("server has already shutdown")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("timeout exceeded when dialling server")
		}
		return err
	}

	return nil
}

func (c *client) ListenForMessages(errChan chan<- error) {
	ctx, cancelMain := context.WithCancel(context.Background())

	defer func() {
		c.startCleanup(ctx)
		cancelMain()
	}()

	// go c.handleSignals()

	messageHandler := make(chan error)
	go func() {
		for {
			data, err := c.readFromServer(ctx)
			if err != nil {
				if errors.Is(err, utils.ErrConnectionClosed) {
					log.Println("server has shutdown or not available")
				}
				if errors.Is(err, utils.ErrCtxTimeOut) {
					log.Println("timed out waiting for server")
					continue
				}
				// messageHandler <- err
				// return
				// TODO; SEND THIS ERROR UPSTREAM FOR BUBBLETEA TO INTERCEPT
				log.Fatal(err)
			}

			if data == nil {
				messageHandler <- errors.New("sprlmnl: received an empty payload")
				return
			}

			if err := c.processPayload(ctx, data); err != nil {
				// TODO: in the future reconnect when necessary so to ensure a smooth UX
				// if errors.Is(err, u.ErrCtxTimeOut) {
				// 	messageHandler <- u.ErrLongWait
				// 	return
				// }

				messageHandler <- err
				return
			}
			fmt.Println("waiting...")
		}
	}()

	select {
	case err := <-messageHandler:
		// we'll always send a non-nil error so no need to check
		errChan <- err
		return
	}
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

	case common.Header_HEADER_HEARTBEAT:
		_, ok := payload.GetContent().(*base.Payload_Heartbeat)
		if ok {
			return c.handleHeartbeatPayload()
		}
	}

	return utils.ErrUnspecifiedPayload
}
