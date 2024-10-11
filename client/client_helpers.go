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

func handleServerReadErr(c *Client, err error) error {
	opErr, _ := err.(*net.OpError)

	switch {
	case errors.Is(err, io.EOF):
		// this could be a kick
		return fmt.Errorf("server closed unexpectedly")
		// c.Conn.Close()
	case errors.Is(opErr.Err, os.ErrDeadlineExceeded):
		// deadline for reading from the server time out
		// MAKE SURE THIS DOESN'T KEEP THE CLIENT "CONNECTED" EVEN
		// AFTER THERE'S A NETWORK PROBLEM WITH THE CLIENT
		// log.Println("read timed out. reseting...")
		c.Conn.SetDeadline(time.Now().Add(config.MaxConnectionTime))
		return nil
	default:
		return handleReadError(err)
	}
}

func validateHeader(header []byte) (bool, error) {
	if len(header) == 0 {
		return false, fmt.Errorf("no header received from server")
	}

	if strings.Contains(string(header), config.ServerFull) {
		return false, fmt.Errorf("server is at capacity. contact the session owner")
	}

	version, identifier, timestampBytes, err := parseHeader(header)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(version, []byte("v1")) {
		return false, fmt.Errorf("invalid version")
	}

	if !bytes.Equal(identifier, []byte("sprlmnl")) {
		return false, fmt.Errorf("invalid identifier")
	}

	return validateTimestamp(timestampBytes)
}

func parseHeader(header []byte) (version, identifier, timestampBytes []byte, err error) {
	firstDelim := bytes.IndexByte(header, '.')
	if firstDelim == -1 {
		return nil, nil, nil, fmt.Errorf("invalid header format")
	}

	version = header[:firstDelim]

	secondDelim := bytes.IndexByte(header[firstDelim+1:], '.') + firstDelim + 1
	if secondDelim == firstDelim {
		return nil, nil, nil, fmt.Errorf("invalid header format")
	}

	identifier = header[firstDelim+1 : secondDelim]
	timestampBytes = header[secondDelim+1:]

	if len(timestampBytes) == 0 {
		return nil, nil, nil, fmt.Errorf("timestamp missing")
	}

	return version, identifier, timestampBytes, nil
}

func validateTimestamp(timestampBytes []byte) (bool, error) {
	timestamp, err := strconv.ParseInt(string(timestampBytes), 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse timestamp: %v", err)
	}

	unixTime := time.Unix(timestamp, 0)
	now := time.Since(unixTime).Seconds()
	if now > float64(config.MaxConnectionTime) {
		return false, fmt.Errorf("time difference too large. rejected")
	}

	return true, nil
}

func readAndValidateHeader(conn net.Conn) error {
	header := make([]byte, 21)
	_, err := conn.Read(header)
	if err != nil {
		return handleReadError(err)
	}

	if ok, err := validateHeader(header); !ok {
		return err
	}

	return nil
}

func handleReadError(err error) error {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return err
	}

	switch {
	case strings.Contains(opErr.Error(), config.ServerClosed):
		return fmt.Errorf("server not accepting connections; didn't receive authentication key")
	case strings.Contains(opErr.Error(), config.ConnectionReset):
		return fmt.Errorf("server reset connection because it shut down; didn't receive authentication key")
	case strings.Contains(opErr.Error(), config.ServerIO):
		return fmt.Errorf("deadline for reading authentication key exceeded due to i/o timeout")
	default:
		return err
	}
}
