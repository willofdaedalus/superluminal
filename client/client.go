package client

import (
	"context"
	"errors"
	"log"
	"net"
	"willofdaedalus/superluminal/utils"
)

const (
	maxConnTries = 3
)

type Client struct {
	Name     string
	PassUsed string
}

func CreateClient(name string) *Client {
	return &Client{
		Name: name,
	}
}

func (c *Client) ConnectToServer(host, port string) error {
	buf := make([]byte, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// to get the dialwithctx func
	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		buf, err = utils.TryRead(ctx, conn, maxConnTries)
		if err != nil {
			// we assume we couldn't read from the server then
			cancel()
			return err
		}

		hdrType, message, err := c.parseIncomingMsg(buf)
		if err != nil {
			// we connected to the wrong server
			if errors.Is(err, utils.ErrWrongServer) {
				log.Println(err)
				return err
			} else if errors.Is(err, utils.ErrUnknownHeader) {
				log.Println(err)
				return err
			}
		}

		c.doActionWithHeader(hdrType, message)
	}
}
