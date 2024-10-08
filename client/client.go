package client

import (
	"io"
	"log"
	"net"
	"time"
)

type Client struct {
	Name       string    // name user sends to application
	Alive      bool      // client still connected
	Master     bool      // owner of the current
	TimeJoined time.Time // timestamp of when client connected
	PassUsed   string
	Conn       net.Conn
}

func CreateClient(name string, owner bool, conn net.Conn) *Client {
	return &Client{
		Name:       name,
		TimeJoined: time.Now(),
		Alive:      true,
		Master:     owner,
		Conn:       conn,
	}
}

// Example of how to listen for messages from the server
func (c *Client) ListenForMessages() {
	buf := make([]byte, 1024)
	msg := make(chan struct{})

	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("connection closed for client %s: %v", c.Name, err)
				c.Alive = false
				msg<-struct{}{}
				return
			}
		}

		message := string(buf[:n])
		log.Printf("received message: %s", message)
		<-msg
	}
}
