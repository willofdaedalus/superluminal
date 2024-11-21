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
	"syscall"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	u "willofdaedalus/superluminal/internal/utils"
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
}

func New(name string) *client {
	return &client{
		name:        name,
		joined:      time.Now(),
		TermContent: make(chan string, 1),
		exitChan:    make(chan struct{}, 1),
		sigChan:     make(chan os.Signal, 1),
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

func (c *client) cleanResources() {
	c.serverConn.Close()
	close(c.exitChan)
}

func (c *client) ListenForMessages(errChan chan<- error) {
	ctx := context.Background()
	defer c.cleanResources()
	defer close(errChan) // Ensure error channel gets closed

	messageHandler := make(chan error)
	go func() {
		for {
			data, err := u.TryReadCtx(ctx, c.serverConn)
			if err != nil {
				if errors.Is(err, u.ErrConnectionClosed) {
					log.Println("server has shutdown or not available")
				}
				if errors.Is(err, u.ErrCtxTimeOut) {
					log.Println("timed out waiting for server")
					continue
				}
				messageHandler <- err
				return
			}

			if data == nil {
				messageHandler <- errors.New("sprlmnl: received an empty payload")
				return
			}

			if err := c.processPayload(ctx, data); err != nil {
				messageHandler <- err
				return
			}
		}
	}()

	select {
	case err := <-messageHandler:
		if err != nil {
			errChan <- err
		}
		fmt.Println("brrrr")
		c.exitChan <- struct{}{}
	case <-c.exitChan:
		fmt.Println("brrrr (from exit)")
	}
}

func (c *client) processPayload(ctx context.Context, data []byte) error {
	procCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

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
			return c.handleAuthPayload(procCtx)
		}

	case common.Header_HEADER_INFO:
		infoPayload, ok := payload.GetContent().(*base.Payload_Info)
		if ok {
			return c.handleInfoPayload(*infoPayload)
		}
	}

	return u.ErrUnspecifiedPayload
}

type Client struct {
	Name             string
	PassUsed         string
	Conn, serverConn net.Conn
	shutdownChan     chan struct{}
	sigChan          chan os.Signal
	signals          []os.Signal
}

func CreateClient(name, pass string) *Client {
	return &Client{
		Name:         name,
		PassUsed:     pass,
		shutdownChan: make(chan struct{}, 1),
		sigChan:      make(chan os.Signal, 1),
	}
}

func (c *Client) ConnectToServer(host, port string) error {
	buf := make([]byte, 1024)
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// to get the dialwithctx func
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	c.serverConn = conn

	// Create a channel to listen for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the signal handling goroutine
	go func() {
		select {
		case <-sigChan:
			// received a signal, cancel the context and return from the function
			cancel()
			errChan <- fmt.Errorf("received signal, exiting ConnectToServer")
		case <-ctx.Done():
			// Context was canceled, likely due to an error in the read loop
			errChan <- ctx.Err()
		}
	}()

	go c.handleSignals()
	go func() {
		for {
			buf, err = u.TryRead(ctx, c.serverConn, maxConnTries)
			if err != nil {
				// we assume we couldn't read from the server after many retries
				cancel()
				errChan <- err
				return
			}
			hdrType, hdrMsg, message, err := u.ParseIncomingMsg(buf)
			if err != nil {
				log.Println(err)
				errChan <- err
				return
			}
			c.doActionWithMsg(ctx, hdrType, hdrMsg, message)
		}
	}()

	// wait for either a signal or an error from the read loop
	return <-errChan
}

func (c *Client) doActionWithMsg(ctx context.Context, hdrVal, hdrMsg int, msg []byte) error {
	switch hdrVal {
	case u.HdrInfoVal:
		return c.fulfillInfoReq(ctx, hdrMsg)
	case u.HdrErrVal:
		return c.fulfillErrReq(ctx, hdrMsg)
	case u.HdrAckVal:
		return c.fulfillAckReq(ctx, hdrMsg)
	}

	// FIXME
	// why is that FIXME there??
	return nil
}

func (c *Client) handleSignals() {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGKILL}
	signal.Notify(c.sigChan, signals...)

	for {
		select {
		case <-c.sigChan:
			c.startCleanup(context.Background())
			return
		}
	}

}

func (c *Client) startCleanup(ctx context.Context) {
	cleanCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	u.TryWrite(cleanCtx, &u.WriteStruct{
		Conn:     c.serverConn,
		MaxTries: 3,
		HdrVal:   u.HdrAckVal,
		HdrMsg:   u.AckClientShutdown,
	})

	// don't read anything from the server
	// serverResp, err := u.TryRead(cleanCtx, c.serverConn, maxConnTries)
	// if err != nil {
	// 	if error.Is(err, io.EOF) {
	// 		// server already shutdown
	// 		return
	// 	}

	// 	c.serverConn.Close()
	// }
	c.serverConn.Close()
}
