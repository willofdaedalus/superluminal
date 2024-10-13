package client

// send user's password to the server for authentication
// once the user is authenticated, encode and send the whole user client
// object to the server to decode and store

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"

	"willofdaedalus/superluminal/config"
)

var (
	readAttempts = 3

	client *Client
	reader *bufio.Reader = bufio.NewReader(os.Stdin)
)

type TransportClient struct {
	Name       string    // name user sends to application
	Alive      bool      // client still connected
	Master     bool      // owner of the current
	TimeJoined int64 // timestamp of when client connected
	PassUsed   string
}

type Client struct {
	TransportClient
	Conn net.Conn
}

func CreateClient(name string, owner bool, conn net.Conn) *Client {
	c := &Client{}
	c.Name = name
	c.TimeJoined = time.Now().Unix()
	c.Alive = true
	c.Master = owner
	c.Conn = conn

	return c
}

func ConnectToServer(ctx context.Context, pass string) error {
	var d net.Dialer
	buf := make([]byte, 1024)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(config.MaxServerHandshakeTime))
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", config.DefaultConnection)
	if err != nil {
		return fmt.Errorf("couldn't find server: %w", err)
	}
	defer conn.Close()

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return handleReadError(client, err)
		}

		contents := string(buf[:4])
		switch contents {
		case config.PreHeader:
			header := buf[4 : 21+4]
			if err := validateAndJoin(header, conn, pass); err != nil {
				return err
			}
			fmt.Println("sent client data")
		case config.PreShutdown:
			return config.ErrServerShutdown
		case config.PreInfo:
			fmt.Println(string(buf[4:n]))
		case config.PreErr:
			msg := string(buf[4:n])
			if err := parseServerErrMessage(msg); err != nil {
				return err
			}

			// Prompt user for a new password on error
			fmt.Printf("Re-enter the passphrase: ")
			newPass, err := reader.ReadString('\n') // Read until newline
			if err != nil {
				return fmt.Errorf("error reading input: %v", err)
			}

			fmt.Println("client sending pass")
			conn.Write([]byte(newPass))
		default:
			return config.ErrUnknownMessage
		}
	}
}

func parseServerErrMessage(msg string) error {
	switch msg {
	case config.ServerFull:
		return config.ErrServerFull
	case config.RejectedPass:
		fmt.Println("sprlmnl:", config.RejectedPass)
		return nil // no error, just retry
	case config.ClientAuthFailed:
		return config.ErrWrongServerPass
	default:
		return fmt.Errorf("sprlmnl: unknown server message: %s", msg)
	}
}

func validateAndJoin(header []byte, conn net.Conn, pass string) error {
	var toSend bytes.Buffer

	// Validate the header
	if err := validateHeader(header); err != nil {
		return err
	}

	// Create the client and send data once validated
	client = CreateClient(fmt.Sprintf("%d", time.Now().UnixMicro()), false, conn)
	client.PassUsed = pass
	enc := gob.NewEncoder(&toSend)
	if err := enc.Encode(client.TransportClient); err != nil {
		return config.ErrEncodingClient
	}

	if _, err := conn.Write(toSend.Bytes()); err != nil {
		return config.ErrSendingClient
	}

	return nil
}
