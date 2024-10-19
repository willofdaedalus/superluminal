package client

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
	client *Client
	reader *bufio.Reader = bufio.NewReader(os.Stdin)
)

type TransportClient struct {
	Name     string // name user sends to application
	PassUsed string
}

type Client struct {
	TransportClient
	Conn       net.Conn
	TimeJoined int64
}

func CreateClient(name string, conn net.Conn) *Client {
	c := &Client{}
	c.Name = name
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

	// need to find a way to exit if the read is blocked
	go handleSignals(conn)

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
			return nil
			// return config.ErrServerShutdown
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
	case
		config.RejectedPass,
		fmt.Sprintf("%s - passphrase has rotated. contact session owner", config.RejectedPass):

		fmt.Println("sprlmnl:", msg)
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
	client = CreateClient(fmt.Sprintf("%d", time.Now().UnixMicro()), conn)
	client.PassUsed = pass
	enc := gob.NewEncoder(&toSend)
	if err := enc.Encode(client.TransportClient); err != nil {
		return config.ErrEncodingClient
	}

	// not using the sendmessage function to preserver byte order
	if _, err := conn.Write(toSend.Bytes()); err != nil {
		return config.ErrSendingClient
	}

	return nil
}
