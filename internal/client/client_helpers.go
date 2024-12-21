package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
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
	passEntryTimeout = time.Minute * 5
	// passEntryTimeout   = time.Second * 5
	cleanupTime        = time.Second * 30
	serverShutdownTime = time.Second * 20
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
		// let the client handle closing
		log.Println(string(payload.Error.GetDetail()))
		c.exitChan <- struct{}{}
		return utils.ErrClientFailedAuth
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
	payload, err := base.EncodePayload(common.Header_HEADER_AUTH, authResp)
	if err != nil {
		return err
	}

	if err := utils.WriteFull(authCtx, c.serverConn, c.tracker, payload); err != nil {
		return err
	}

	c.sentPass = true
	return nil
}

func (c *client) handleHeartbeatPayload() error {
	return nil
}

func (c *client) handleInfoPayload(ctx context.Context, payload base.Payload_Info) error {
	switch payload.Info.GetInfoType() {
	case info.Info_INFO_AUTH_SUCCESS:
		log.Println(payload.Info.GetMessage())
		c.isApproved = true
		return nil

	case info.Info_INFO_SHUTDOWN:
		c.handleServerShutdown(ctx)
		// c.exitChan <- struct{}{}
		log.Println(payload.Info.GetMessage())
		os.Exit(0)
		return nil

		// case info.Info_INFO_REQ_ACK:
		// 	if utils.GoodbyeMsg == payload.Info.GetMessage() {
		// 		c.serverConn.Close()
		// 	}
	}

	return utils.ErrUnspecifiedPayload
}

func (c *client) handleServerShutdown(ctx context.Context) error {
	shutCtx, cancel := context.WithTimeout(ctx, serverShutdownTime)
	defer func() {
		c.serverConn.Close()
		cancel()
	}()

	// wait for a short duration for actions to complete
	for i := 0; i < maxConnTries; i++ {
		if !c.tracker.AnyActionInProgress() {
			break
		}
		select {
		case <-time.After(time.Second * 5):
			continue
		case <-shutCtx.Done(): // this might not be necessary
			return shutCtx.Err()
			// case <-ctx.Done():
			// 	return ctx.Err()
		}
	}

	// If actions still in progress, force close
	if c.tracker.AnyActionInProgress() {
		log.Println("Forcing connection close due to ongoing actions")
	}

	return nil
}

func (c *client) handleTermPayload(payload base.Payload_TermContent) error {
	termContent := payload.TermContent

	sameLen := len(termContent.GetData()) == int(termContent.GetMessageLength())
	crcMatch := crc32.ChecksumIEEE(termContent.GetData()) == termContent.GetCrc32()

	if !sameLen {
		return fmt.Errorf("message data length differ")
	}
	if !crcMatch {
		return fmt.Errorf("crc doesn't match")
	}

	fmt.Print(string(termContent.GetData()))

	return nil
}

func (c *client) startCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTime)

	defer func() {
		utils.SafeClose(c.sigChan)
		utils.SafeClose(c.exitChan)
		c.serverConn.Close()
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// only send shutdown message if approved
			if c.isApproved {
				infoPayload := base.GenerateInfo(info.Info_INFO_SHUTDOWN, "client_shutdown")
				payload, err := base.EncodePayload(common.Header_HEADER_INFO, infoPayload)
				if err != nil {
					return
				}

				fmt.Println("beginning write")
				// use the cleanup context for writing
				err = utils.WriteFull(ctx, c.serverConn, c.tracker, payload)
				if err != nil {
					log.Println("Error during shutdown write:", err)
				}

				fmt.Println("wrote to server")
				c.isApproved = false
				cancel()

			}
		}
	}

}

func (c *client) handleSignals(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	signal.Notify(c.sigChan, signals...)
	defer signal.Stop(c.sigChan)

	sig := <-c.sigChan
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		// cancel the context to trigger shutdown
		c.exitChan <- struct{}{}
		fmt.Println("received signal:", sig)
		return
	}
}
