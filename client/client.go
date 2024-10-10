package client

// send user's password to the server for authentication
// once the user is authenticated, encode and send the whole user client
// object to the server to decode and store

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	// "io"
	"log"
	"net"
	"time"

	"willofdaedalus/superluminal/config"
)

type TransportClient struct {
	Name       string    // name user sends to application
	Alive      bool      // client still connected
	Master     bool      // owner of the current
	TimeJoined time.Time // timestamp of when client connected
	PassUsed   string
}

type Client struct {
	TransportClient
	Conn net.Conn
}

func CreateClient(name string, owner bool, conn net.Conn) *Client {
	c := &Client{}
	c.Name = name
	c.TimeJoined = time.Now()
	c.Alive = true
	c.Master = owner
	c.Conn = conn

	return c
}

func (c *Client) ListenForMessages(ctx context.Context) {
	buf := make([]byte, 1024)

	for {
		// case <-ctx.Done():
		// 	log.Println("context canceled, stopping read")
		// 	return
		n, err := c.Conn.Read(buf)
		if err != nil {
			opErr, _ := err.(*net.OpError)
			if err == io.EOF {
				// this could be a kick
				log.Println("server closed unexpectedly")
				c.Conn.Close()
				return
			}
			// if !ok {
			// 	fmt.Println("not network error:", err)
			// 	c.Conn.Close()
			// 	return
			// }
			// deadline for reading from the server time out
			if errors.Is(opErr.Err, os.ErrDeadlineExceeded) {
				log.Println("read timed out. reseting...")
				c.Conn.SetDeadline(time.Now().Add(config.MaxConnectionTime))
				continue
			}
			// if err != io.EOF {
			// 	log.Printf("connection closed for client %q: %v", c.Name, err)
			// 	c.Alive = false
			// 	return
			// }
		}

		message := string(buf[:n])
		if strings.Contains(message, config.RejectedPass) {
			fmt.Println(message)
			return
		}
		fmt.Println(message)
	}
}

func ConnectToServer(ctx context.Context, pass string) error {
	var d net.Dialer
	var toSend bytes.Buffer

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(config.MaxConnectionTime))
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", config.DefaultConnection)
	if err != nil {
		return fmt.Errorf("couldn't find server: %w", err)
	}
	defer conn.Close()

	// set a deadline for the read operation
	conn.SetDeadline(time.Now().Add(config.MaxConnectionTime))
	header := make([]byte, 21)
	_, err = conn.Read(header)
	if err != nil {
		return handleReadError(err)
	}

	if ok, err := validateHeader(header); !ok {
		return err
	}

	// create a new client and then transport only the necessary parts to the client
	client := CreateClient("john doe", false, conn)
	tc := client.TransportClient
	tc.PassUsed = pass
	enc := gob.NewEncoder(&toSend)
	err = enc.Encode(tc)
	if err != nil {
		return fmt.Errorf("struct encoding err: %v", err)
	}

	errChan := make(chan error)
	go func() {
		_, err = conn.Write(toSend.Bytes())
		errChan <- err
	}()

	select {
	case err = <-errChan:
		if err != nil {
			return fmt.Errorf("couldn't send the struct across the network: %v", err)
		}
		// case <-ctx.Done():
		// 	conn.Close()
		// 	return fmt.Errorf("couldn't find or connect to server: %v", err)
	}

	client.ListenForMessages(ctx)
	return nil
}
