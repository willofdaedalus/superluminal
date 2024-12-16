package client

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
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
	//passEntryTimeout = time.Minute * 2
	passEntryTimeout   = time.Second * 5
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

	if err := c.writeToServer(authCtx, payload); err != nil {
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

func (c *client) startCleanup(ctx context.Context) {
	cleanCtx, cancel := context.WithTimeout(ctx, cleanupTime)

	// safely close connections and channels
	defer func() {
		// close these only if they haven't been closed already
		utils.SafeClose(c.sigChan)
		utils.SafeClose(c.exitChan)
		c.serverConn.Close()
		cancel()
	}()

	// Only send shutdown message if approved
	if c.isApproved {
		infoPayload := base.GenerateInfo(info.Info_INFO_SHUTDOWN, "client_shutdown")
		payload, err := base.EncodePayload(common.Header_HEADER_INFO, infoPayload)
		if err != nil {
			return
		}

		// Use the cleanup context for writing
		err = c.writeToServer(cleanCtx, payload)
		if err != nil {
			// Log or handle write errors as needed
			log.Println("Error during shutdown write:", err)
		}

		fmt.Println("wrote to server")

		c.isApproved = false
	}
}

func (c *client) handleSignals() {
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

// writeToServer provides a way to synchronize writes across the entire backend
func (c *client) writeToServer(ctx context.Context, data []byte) error {
	c.tracker.IncrementWrite()
	defer c.tracker.DecrementWrite()

	return utils.TryWriteCtx(ctx, c.serverConn, data)
}

func (c *client) readFromServer(ctx context.Context) ([]byte, error) {
	c.tracker.IncrementRead()
	defer c.tracker.DecrementRead()
	readCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Set a read deadline if context has a timeout
	if deadline, ok := readCtx.Deadline(); ok {
		if err := c.serverConn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	// Read initial header
	headerBuffer := make([]byte, 4)
	if _, err := io.ReadFull(c.serverConn, headerBuffer); err != nil {
		return nil, fmt.Errorf("header read error: %w", err)
	}

	// Extract payload length
	payloadLen := binary.BigEndian.Uint32(headerBuffer)

	// Sanity check on payload length
	if payloadLen > utils.MaxPayloadSize {
		return nil, fmt.Errorf("payload length exceeds maximum allowed size: %d", payloadLen)
	}

	// Log payload size for debugging
	log.Printf("Reading payload of size: %d bytes", payloadLen)

	// Allocate space for the full payload
	actualPayload := make([]byte, payloadLen)

	// Use io.ReadFull to read the entire payload
	if _, err := io.ReadFull(c.serverConn, actualPayload); err != nil {
		return nil, fmt.Errorf("payload read error: %w", err)
	}

	return actualPayload, nil
}

// // readFromServer provides a way to synchronize reads across the client
// func (c *client) readFromServer(ctx context.Context) ([]byte, error) {
// 	c.tracker.IncrementRead()
// 	defer c.tracker.DecrementRead()

// 	return utils.TryReadCtx(ctx, c.serverConn)
// }
