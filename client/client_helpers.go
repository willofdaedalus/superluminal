package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"willofdaedalus/superluminal/config"
)

func handleMessage(message string) error {
	switch {
	case strings.Contains(message, config.RejectedPass):
		return fmt.Errorf(message)
	case message == config.ShutdownMsg:
		// fmt.Println("server is shutting down")
		return fmt.Errorf("server is shutting down")
	default:
		fmt.Println(message)
		return nil
	}
}

func validateHeader(header []byte) error {
	if len(header) == 0 {
		return config.ErrEmptyHeader
	}

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

	if len(tsBytes) == 0 {
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
	case errors.Is(err, io.EOF):
		return config.ErrServerShutdown
	}

	return err
}

// func handleServerReadErr(c *Client, err error) error {
// 	opErr, _ := err.(*net.OpError)
//
// 	switch {
// 	case errors.Is(err, io.EOF):
// 		// this could be a kick
// 		return fmt.Errorf("server closed unexpectedly")
// 		// c.Conn.Close()
// 	case errors.Is(opErr.Err, os.ErrDeadlineExceeded):
// 		// deadline for reading from the server time out
// 		// MAKE SURE THIS DOESN'T KEEP THE CLIENT "CONNECTED" EVEN
// 		// AFTER THERE'S A NETWORK PROBLEM WITH THE CLIENT
// 		// log.Println("read timed out. reseting...")
// 		c.Conn.SetDeadline(time.Now().Add(config.MaxConnectionTime))
// 		return nil
// 	default:
// 		return handleReadError(err)
// 	}
// }
