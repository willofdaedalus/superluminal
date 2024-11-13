package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
	u "willofdaedalus/superluminal/internal/utils"
)

// handles the various ack messages that come from the server
func (c *Client) fulfillAckReq(ctx context.Context, hdrMsg int) error {
	switch hdrMsg {
	case u.AckSelfReport:
		return c.encodeAndSend(ctx)
	default:
		return fmt.Errorf("sprlmnl: unknown ack msg")
	}
}

// encode and send the client data to the server for safe keeping :)
func (c *Client) encodeAndSend(ctx context.Context) error {
	sendCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var client bytes.Buffer
	enc := gob.NewEncoder(&client)
	err := enc.Encode(c)
	if err != nil {
		return err
	}

	err = u.TryWrite(sendCtx, &u.WriteStruct{
		Conn:     c.serverConn,
		MaxTries: maxConnTries,
		HdrVal:   u.HdrResVal,
		HdrMsg:   u.RespSelfReport,
		Message:  client.Bytes(),
	})

	if err != nil {
		return err
	}

	return nil
}
