package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	err1 "willofdaedalus/superluminal/internal/payload/error"
	"willofdaedalus/superluminal/internal/payload/info"
	"willofdaedalus/superluminal/internal/utils"
)

const (
	passEntryTimeout = time.Second * 5
)

func (c *client) handleErrPayload(payload base.Payload_Error) error {
	switch payload.Error.GetCode() {
	case err1.ErrorMessage_ERROR_SERVER_FULL:
		// TODO => add a notification system to send messages like this to the
		// user through bubble tea type  interface
		log.Println(string(payload.Error.GetDetail()))
		c.exitChan <- struct{}{}
		return utils.ErrServerFull
	case err1.ErrorMessage_ERROR_AUTH_FAILED:
		log.Println(string(payload.Error.GetDetail()))
		c.exitChan <- struct{}{}
		return utils.ErrFailedServerAuth
	}

	return utils.ErrUnspecifiedPayload
}

func (c *client) handleAuthPayload(ctx context.Context) error {
	var pass string

	authCtx, cancel := context.WithTimeout(ctx, passEntryTimeout)
	defer cancel()

	passChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		prompt := "enter passphrase: "
		if c.sentPass {
			prompt = "re-enter the passphrase: "
		}
		fmt.Print(prompt)

		// use a scanner to handle potential input issues
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			passChan <- scanner.Text()
		} else {
			errChan <- scanner.Err()
		}
	}()

	select {
	case <-ctx.Done():
		log.Print("you waited too long")
		return errors.New("passphrase entry timed out")
	case inputErr := <-errChan:
		return fmt.Errorf("input error: %w", inputErr)
	case pass = <-passChan:
		// Continue with authentication
	}

	authResp := base.GenerateAuthResp(c.name, pass)
	resp, err := base.EncodePayload(common.Header_HEADER_AUTH, authResp)
	if err != nil {
		return err
	}

	if err := utils.TryWriteCtx(authCtx, c.serverConn, resp); err != nil {
		return err
	}

	c.sentPass = true
	return nil
}

func (c *client) handleInfoPayload(payload base.Payload_Info) error {
	switch payload.Info.GetInfoType() {
	case info.Info_INFO_AUTH_SUCCESS:
		log.Println(payload.Info.GetMessage())
		return nil
	}

	return utils.ErrUnspecifiedPayload
}

func (c *client) handleSignals() {
	signals := []os.Signal{syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL}
	signal.Notify(c.sigChan, signals...)
}
