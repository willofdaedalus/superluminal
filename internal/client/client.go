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
	pass        string
	joined      time.Time
	serverConn  net.Conn
	exitChan    chan struct{}
}

func New(name, pass string) *client {
	return &client{
		name:        name,
		joined:      time.Now(),
		pass:        pass,
		TermContent: make(chan string, 1),
		exitChan:    make(chan struct{}, 1),
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

	go func() {
		defer func() {
			c.exitChan <- struct{}{}
		}()

		for {
			data, err := u.TryReadCtx(ctx, c.serverConn)
			if err != nil {
				if errors.Is(err, u.ErrServerClosed) {
					log.Println("server has shutdown or not available")
				}
				errChan <- err
				return
			}

			// data may never be nil but this a good check nonetheless
			if data == nil {
				errChan <- errors.New("sprlmnl: received an empty payload")
				return
			}

			if err := c.processPayload(data); err != nil {
				errChan <- err
				// maybe don't return too early??
				return
			}
		}
	}()

	for {
		select {
		case <-c.exitChan:
			return
		}
	}

}

func (c *client) processPayload(data []byte) error {
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
