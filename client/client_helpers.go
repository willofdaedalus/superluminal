package client

import (
	"bytes"
	"fmt"
	"willofdaedalus/superluminal/utils"
)

// header types
const (
	HDR_UKNOWN = iota + 1
	HDR_INFO
	HDR_ACK
	HDR_ERR
)

func (c *Client) parseIncomingMsg(msg []byte) (int, []byte, error) {
	header, message, ok := bytes.Cut(msg, []byte(":"))
	if !ok {
		return HDR_UKNOWN, nil, utils.ErrInvalidHeader
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

func (c *Client) doActionWithHeader(headerType int, msg []byte) {
	switch headerType {
	case HDR_INFO:
		fmt.Println("info")
		fmt.Printf("%s", msg)
	case HDR_ERR:
		fmt.Println("err")
		fmt.Printf("%s", msg)
	case HDR_ACK:
		fmt.Println("ack")
		fmt.Printf("%s", msg)
	}
}

func (c *Client) encodeAndSend() error {
	return nil
}

// parse the server's header for more info and what to do
func parseServerHeader(header []byte) (int, error) {
	if len(header) != 11 {
		return HDR_UKNOWN, utils.ErrInvalidHeader
	}

	split := bytes.Split(header, []byte("+"))
	if !bytes.Contains(split[0], []byte("sprlmnl")) || len(split[1]) != 3 {
		return HDR_UKNOWN, utils.ErrInvalidHeader
	}

	switch string(split[1]) {
	case "inf":
		return HDR_INFO, nil
	case "err":
		return HDR_ERR, nil
	case "ack":
		return HDR_ACK, nil
	default:
		return HDR_UKNOWN, utils.ErrUnknownHeader
	}
}
