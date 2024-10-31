package client

import (
	"context"
	"log"
	"net"
	"willofdaedalus/superluminal/utils"
)

const (
	maxConnTries = 3
)

type Client struct {
	Name             string
	PassUsed         string
	Conn, serverConn net.Conn
}

func CreateClient(name, pass string) *Client {
	return &Client{
		Name:     name,
		PassUsed: pass,
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
	c.serverConn = conn
	defer c.serverConn.Close()

	for {
		buf, err = utils.TryRead(ctx, c.serverConn, maxConnTries)
		if err != nil {
			// we assume we couldn't read from the server after many retries
			cancel()
			return err
		}

		hdrType, hdrMsg, message, err := utils.ParseIncomingMsg(buf)
		if err != nil {
			// we connected to the wrong server
			// if errors.Is(err, utils.ErrWrongServer) {
			// 	log.Println(err)
			// 	return err
			// } else if errors.Is(err, utils.ErrUnknownHeader) {
			// 	log.Println(err)
			// 	return err
			// }
			log.Println(err)
			return err
		}

		c.doActionWithMsg(ctx, hdrType, hdrMsg, message)
	}
}
