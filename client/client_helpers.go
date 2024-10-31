package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
	u "willofdaedalus/superluminal/utils"

	"github.com/charmbracelet/huh"
)

func (c *Client) doActionWithMsg(ctx context.Context, hdrVal, hdrMsg int, msg []byte) error {
	switch hdrVal {
	case u.HdrInfoVal:
		fmt.Println("info")
		fmt.Printf("%s", msg)
	case u.HdrErrVal:
		c.fulfillErrReq(ctx, hdrMsg)
	case u.HdrAckVal:
		return c.fulfillAckReq(ctx, hdrMsg)
	}

	// FIXME
	// why is that FIXME there??
	return nil
}

func (c *Client) fulfillErrReq(ctx context.Context, hdrMsg int) error {
	switch hdrMsg {
	case u.ErrWrongPassphrase:
		return c.sendPassphrase(ctx)
	default:
		return fmt.Errorf("sprlmnl: unknown err msg")
	}
}

func (c *Client) sendPassphrase(ctx context.Context) error {
	newPass := ""
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("incorrect passphrase. re-enter;").
				Prompt("> ").
				Value(&newPass),
		).WithTheme(huh.ThemeBase16()),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	err = u.TryWrite(ctx, &u.WriteStruct{
		Conn:     c.serverConn,
		MaxTries: maxConnTries,
		HdrVal:   u.HdrResVal,
		HdrMsg:   u.RespNewPass,
		Message:  []byte(newPass),
	})

	return err
}

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
