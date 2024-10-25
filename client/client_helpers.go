package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
	"willofdaedalus/superluminal/utils"
)

// header types
const (
	HdrUnkown = iota + 1
	HdrInfo
	HdrAck
	HdrErr
)

// parse the server's header for more info and what to do
func parseServerHeader(header []byte) (int, error) {
	if len(header) != 11 {
		return HdrUnkown, utils.ErrInvalidHeader
	}

	split := bytes.Split(header, []byte("+"))
	if !bytes.Contains(split[0], []byte("sprlmnl")) || len(split[1]) != 3 {
		return HdrUnkown, utils.ErrInvalidHeader
	}

	switch string(split[1]) {
	case "inf":
		return HdrInfo, nil
	case "err":
		return HdrErr, nil
	case "ack":
		return HdrAck, nil
	default:
		return HdrUnkown, utils.ErrUnknownHeader
	}
}

func (c *Client) parseIncomingMsg(msg []byte) (int, []byte, error) {
	header, message, ok := bytes.Cut(msg, []byte(":"))
	if !ok {
		return HdrUnkown, nil, utils.ErrInvalidHeader
	}

	headerType, err := parseServerHeader(header)
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
	case HdrInfo:
		fmt.Println("info")
		fmt.Printf("%s", msg)
	case HdrErr:
		fmt.Println("err")
		fmt.Printf("%s", msg)
	case HdrAck:
		return c.fulfillAckReq(ctx, msg)
	}

	// FIXME
	return nil
}

// handles the various ack messages that come from the server
func (c *Client) fulfillAckReq(ctx context.Context, msg []byte) error {
	if bytes.Equal(msg, []byte(utils.AckSelfReport)) {
		return c.encodeAndSend(ctx)
	}

	return nil
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

	if err := utils.TryWrite(sendCtx, c.serverConn, maxConnTries, []byte(utils.HdrRes), client.Bytes()); err != nil {
		return err
	}

	return nil
}
