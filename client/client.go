package client

// send user's password to the server for authentication
// once the user is authenticated, encode and send the whole user client
// object to the server to decode and store

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
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

func (c *Client) ListenForMessages() error {
	buf := make([]byte, 1024)

	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			// in this case it was a timeout error
			if handleServerReadErr(c, err) == nil {
				fmt.Println("timeout reset")
				continue
			}

			return handleServerReadErr(c, err)
		}

		message := string(buf[:n])
		if err := handleMessage(message); err != nil {
			return err
		}
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
	if err = readAndValidateHeader(conn); err != nil {
		return err
	}

	// create a new client and then transport only the necessary parts to the client
	client := CreateClient("john doe", false, conn)
	client.PassUsed = pass
	enc := gob.NewEncoder(&toSend)
	err = enc.Encode(client.TransportClient)
	if err != nil {
		return fmt.Errorf("struct encoding err: %v", err)
	}

	_, err = conn.Write(toSend.Bytes())
	if err != nil {
		return fmt.Errorf("couldn't send client across the network: %v", err)
	}

	for {
		return client.ListenForMessages()
	}
}
