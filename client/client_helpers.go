package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
	u "willofdaedalus/superluminal/utils"
)

func (c *Client) doActionWithHeader(ctx context.Context, headerType int, msg []byte) error {
	switch headerType {
	case u.HdrInfoVal:
		fmt.Println("info")
		fmt.Printf("%s", msg)
	case u.HdrErrVal:
		fmt.Println("err")
		fmt.Printf("%s", msg)
	case u.HdrAckVal:
		return c.fulfillAckReq(ctx, msg)
	}

	// FIXME
	// why is that FIXME there??
	return nil
}

// handles the various ack messages that come from the server
func (c *Client) fulfillAckReq(ctx context.Context, msg []byte) error {
	if bytes.Equal(msg, []byte(u.AckSelfReport)) {
		return c.encodeAndSend(ctx)
	}

	return nil
}

// encode and send the client data to the server for safe keeping :)
func (c *Client) encodeAndSend(ctx context.Context) error {
	sendCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	c.PassUsed = "password"
	var client bytes.Buffer
	enc := gob.NewEncoder(&client)
	err := enc.Encode(c)
	if err != nil {
		return err
	}

	if err := u.TryWrite(sendCtx, c.serverConn, maxConnTries, []byte(u.HdrResMsg), client.Bytes()); err != nil {
		return err
	}

	return nil
}
