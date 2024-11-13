package client

import (
	"context"
	"fmt"
	"github.com/charmbracelet/huh"
	"log"
	u "willofdaedalus/superluminal/internal/utils"
)

func (c *Client) fulfillInfoReq(ctx context.Context, hdrMsg int) error {
	switch hdrMsg {
	case u.InfoServerShutdown:
		log.Println("server is shutting down")
		c.serverConn.Close()
		return nil
	default:
		return fmt.Errorf("unknown info msg")
	}
}

func (c *Client) fulfillErrReq(ctx context.Context, hdrMsg int) error {
	switch hdrMsg {
	case u.ErrWrongPassphrase:
		return c.sendPassphrase(ctx)
	case u.ErrServerKick:
		log.Println("sprlmnl: server kicked you for wrong passphrase")
		return nil
	default:
		return fmt.Errorf("sprlmnl: unknown err msg")
	}
}

// resends the passphrase to the server again
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

	if err != nil {
		log.Fatal(err)
	}

	return err
}
