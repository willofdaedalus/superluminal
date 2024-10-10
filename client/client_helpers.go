package client

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"willofdaedalus/superluminal/config"
)

func validateHeader(header []byte) (bool, error) {
	if len(header) == 0 {
		return false, fmt.Errorf("no header received")
	}

	// look for the first delimiter (.)
	firstDelim := bytes.IndexByte(header, '.')
	if firstDelim == -1 {
		return false, fmt.Errorf("invalid header format")
	}

	// check version bytes
	version := header[:firstDelim]
	if !bytes.Equal(version, []byte("v1")) {
		return false, fmt.Errorf("invalid version")
	}

	// look for the second delimiter (.)
	secondDelim := bytes.IndexByte(header[firstDelim+1:], '.') + firstDelim + 1
	if secondDelim == firstDelim {
		return false, fmt.Errorf("invalid header format")
	}

	// check identifier bytes
	identifier := header[firstDelim+1 : secondDelim]
	if !bytes.Equal(identifier, []byte("sprlmnl")) {
		return false, fmt.Errorf("invalid identifier")
	}

	// extract timestamp bytes
	timestampBytes := header[secondDelim+1:]
	if len(timestampBytes) == 0 {
		return false, fmt.Errorf("timestamp missing")
	}

	// convert timestamp bytes to int64
	timestamp, err := strconv.ParseInt(string(timestampBytes), 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse timestamp: %v", err)
	}

	// validate timestamp
	unixTime := time.Unix(timestamp, 0)
	now := time.Since(unixTime).Seconds()
	if now > float64(config.MaxConnectionTime) {
		return false, fmt.Errorf("time difference too large. rejected")
	}

	return true, nil
}

// Function to handle read errors
func handleReadError(err error) error {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return err
	}

	if strings.Contains(opErr.Error(), config.ServerClosed) {
		return fmt.Errorf("server not accepting connections; didn't receive authentication key")
	} else if strings.Contains(opErr.Error(), config.ConnectionReset) {
		return fmt.Errorf("server reset connection because it shut down; didn't receive authentication key")
	} else if strings.Contains(opErr.Error(), config.ServerIO) {
		return fmt.Errorf("deadline for reading authentication key exceeded due to i/o timeout")
	}

	return err
}
