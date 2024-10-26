package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
	u "willofdaedalus/superluminal/utils"
)

// parse the server's header for more info and what to do
func parseHeader(header []byte) (int, error) {
	if len(header) != 11 {
		return u.HdrUnknownVal, u.ErrInvalidHeader
	}

	split := bytes.Split(header, []byte("+"))
	if !bytes.Contains(split[0], []byte("sprlmnl")) || len(split[1]) != 3 {
		return u.HdrUnknownVal, u.ErrInvalidHeader
	}

	headerType := u.GetHeaderType(split[1])
	if headerType == u.HdrUnknownVal {
		return u.HdrUnknownVal, u.ErrUnknownHeader
	}

	return headerType, nil
}

func (c *Client) parseIncomingMsg(msg []byte) (int, []byte, error) {
	header, message, ok := bytes.Cut(msg, []byte(":"))
	if !ok {
		return u.HdrUnknownVal, nil, u.ErrInvalidHeader
	}

	headerType, err := parseHeader(header)
	if err != nil {
		return headerType, nil, err
	}

	// switch headerType {
	// case ack:
	// 	return c.parseAckMessage(message)
	// case info:
	// 	log.Printf("%s", message)
	// 	return nil
	// }

	return headerType, message, nil
}

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
