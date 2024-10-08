package cli

import (
	"context"
	"fmt"
	"net"
	"time"

	c "willofdaedalus/superluminal/client"
)

var (
	DefaultConnection string
)

func ConnectToServer(ctx context.Context) (*c.Client, error) {
	var d net.Dialer
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second * 5))
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", DefaultConnection)
	if err != nil {
		return nil, fmt.Errorf("couldn't find the server")
	}

	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("couldn't find or connect to server due to timeout")
			conn.Close()
			return
		}
	}()

	client := c.CreateClient("default name", false, conn)
	client.ListenForMessages()
	return client, nil
}
