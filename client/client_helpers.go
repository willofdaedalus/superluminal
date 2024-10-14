package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"willofdaedalus/superluminal/config"
	"willofdaedalus/superluminal/utils"
)

func validateHeader(header []byte) error {
	if len(header) != 21 {
		return config.ErrHeaderShort
	}

	// move this out of this function
	if strings.Contains(string(header), config.ServerFull) {
		return config.ErrServerFull
	}

	v, id, tsBytes, err := parseHeader(header)
	if err != nil {
		return err
	}

	if !bytes.Equal(v, []byte("v1")) {
		return config.ErrInvalidHeaderFormat
	}

	if !bytes.Equal(id, []byte("sprlmnl")) {
		return config.ErrInvalidHeaderFormat
	}

	return validateTimestamp(tsBytes)
}

func parseHeader(header []byte) (v, id, tsBytes []byte, err error) {
	firstDelim := bytes.IndexByte(header, '.')
	if firstDelim == -1 {
		return nil, nil, nil, config.ErrInvalidHeaderFormat
	}

	v = header[:firstDelim]

	secondDelim := bytes.IndexByte(header[firstDelim+1:], '.') + firstDelim + 1
	if secondDelim == firstDelim {
		return nil, nil, nil, config.ErrInvalidHeaderFormat
	}

	id = header[firstDelim+1 : secondDelim]
	tsBytes = header[secondDelim+1:]

	// make sure the timestamp is valid
	if len(tsBytes) < 10 {
		return nil, nil, nil, config.ErrInvalidTimestamp
	}

	return v, id, tsBytes, nil
}

func validateTimestamp(tsBytes []byte) error {
	timestamp, err := strconv.ParseInt(string(tsBytes), 10, 64)
	if err != nil {
		return config.ErrInvalidTimestamp
	}

	unixTime := time.Unix(timestamp, 0)
	now := time.Since(unixTime).Seconds()
	if now > float64(config.MaxConnectionTime) {
		return config.ErrTimeDifference
	}

	return nil
}

func readAndValidateHeader(conn net.Conn) error {
	header := make([]byte, 21)
	_, err := conn.Read(header)
	if err != nil {
		fmt.Println("read and validate")
		return handleReadError(nil, err)
	}

	if err := validateHeader(header); err != nil {
		return err
	}

	return nil
}

func handleReadError(c *Client, err error) error {
	var opErr *net.OpError

	if errors.As(err, &opErr) {
		switch {
		case strings.Contains(opErr.Error(), config.ServerClosed):
			return config.ErrServerClosed
		case strings.Contains(opErr.Error(), config.ConnectionReset):
			return config.ErrServerReset
		case strings.Contains(opErr.Error(), config.ServerIO):
			return config.ErrDeadlineTimeout
		case errors.Is(opErr, os.ErrDeadlineExceeded):
			c.Conn.SetDeadline(time.Now().Add(config.MaxConnectionTime))
			return nil
		default: // specific error not parsed
			return err
		}
	}

	switch {
	// make sure the server sends a shutdown message
	case errors.Is(err, io.EOF):
		return config.ErrServerShutdown
	}

	return err
}

func handleSignals(conn net.Conn) {
	sig := make(chan os.Signal, 1)

	signals := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}

	signal.Notify(sig, signals...)

	for {
		select {
		case s := <-sig:
			switch s {
			case syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP:
				err := utils.SendMessage(context.TODO(), conn, config.PreShutdown, "", nil)
				if err != nil {
					log.Println("error sending shutdown message:", err)
				}
				if err := conn.Close(); err != nil {
					log.Println("error closing connection:", err)
				}
				return
				// return fmt.Errorf("exiting client")
			}
		}
	}
}
