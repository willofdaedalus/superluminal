package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
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

// Example of how to listen for messages from the server
func (c *Client) ListenForMessages(ctx context.Context) {
	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			log.Println("context canceled, stopping read")
			return
		default:
			n, err := c.Conn.Read(buf)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Err != nil {
					log.Println("server shutting down. goodbye...")
					c.Conn.Close()
					return
				}
				// if err != io.EOF {
				// 	log.Printf("connection closed for client %q: %v", c.Name, err)
				// 	c.Alive = false
				// 	return
				// }
			}

			message := string(buf[:n])
			if message == "server rejected your passphrase. check and re-enter\n" {
				fmt.Println(message)
				return
			}
			fmt.Println(message)
		}
	}
}

func ConnectToServer(ctx context.Context, pass string) error {
	var d net.Dialer
	var toSend bytes.Buffer

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", config.DefaultConnection)
	if err != nil {
		return fmt.Errorf("couldn't find the server")
	}

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			log.Printf("failed to close the connection: %v", err)
		}
	}()

	// create a new client and then transport only the necessary parts to the client
	client := CreateClient("default name", false, conn)
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
	case <-ctx.Done():
		conn.Close()
		return fmt.Errorf("couldn't find or connect to server: %v", err)
	}

	client.ListenForMessages(ctx)
	return nil
}
