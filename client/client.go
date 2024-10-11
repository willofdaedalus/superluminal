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

func ConnectToServer(ctx context.Context, pass string) error {
	var d net.Dialer
	var toSend bytes.Buffer
	var validSession bool
	var client *Client
	buf := make([]byte, 1024)
	// var errChan chan error

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(config.MaxConnectionTime))
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", config.DefaultConnection)
	if err != nil {
		return fmt.Errorf("couldn't find server: %w", err)
	}
	defer conn.Close()

	for {
		_, err = conn.Read(buf)
		if err != nil {
			return handleReadError(client, err)
		}

		// never confirmed from the server
		if !validSession {
			// we only need the first 21 bytes for validation
			if err = validateHeader(buf[:21]); err != nil {
				return err
			}

			// only send the struct once the header is validated
			client = CreateClient("john doe", false, conn)
			client.PassUsed = pass
			enc := gob.NewEncoder(&toSend)
			err = enc.Encode(client.TransportClient)
			if err != nil {
				return config.ErrEncodingClient
			}

			_, err = conn.Write(toSend.Bytes())
			if err != nil {
				return config.ErrSendingClient
			}

			validSession = true
		}

		fmt.Println(string(buf))
	}
}
