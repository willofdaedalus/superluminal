package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	// MaxConnTime    = time.Second * 30
	MaxConnTime    = time.Minute * 1
	MaxBackoffTime = time.Second * 7
	maxTries       = 5
)

type WriteStruct struct {
	Conn     net.Conn
	MaxTries int
	HdrVal   int
	HdrMsg   int
	Message  []byte
}

func (ws *WriteStruct) headerMsgByte() []byte {
	hdr := strconv.Itoa(ws.HdrVal)
	msg := strconv.Itoa(ws.HdrMsg)
	fin := fmt.Sprintf("%s+%s", hdr, msg)

	return []byte(fin)
}

// TryWriteCtx attempts to write to the conn passed and retries up to a number of times
// defined until it gives up and returns an error
// it respects the context passed to it
func TryWriteCtx(ctx context.Context, conn net.Conn, data []byte) error {
	writeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for tries := 0; tries < maxTries; tries++ {
		deadline, ok := writeCtx.Deadline()
		if ok {
			if err := conn.SetWriteDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return ErrCtxTimeOut
				}

				return ErrDeadlineUnsuccessful
			}
		}

		if tries == maxTries-1 {
			return ErrFailedAfterRetries
		}

		_, err := conn.Write(data)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return ErrCtxTimeOut
			}

			if errors.Is(err, io.EOF) {
				return io.EOF
			}

			retryTime := time.Second * (1 << uint(tries))
			if retryTime > MaxBackoffTime {
				retryTime = MaxBackoffTime
			}

			select {
			case <-writeCtx.Done():
				if errors.Is(writeCtx.Err(), context.DeadlineExceeded) {
					return writeCtx.Err()
				}
			case <-time.After(retryTime):
				log.Printf("retrying connection to server...")
				continue
			}
		}

		// successful write
		conn.SetWriteDeadline(time.Time{}) // Reset the deadline
		return nil
	}

	return ErrFailedAfterRetries
}

// TryReadCtx relies on external timeouts and deadlines in order to function
// properly. this makes it robust for all situations including those that
// are not in any hurry to read something from a connection
func TryReadCtx(ctx context.Context, conn net.Conn) ([]byte, error) {
	var data bytes.Buffer
	writeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	buf := make([]byte, MaxPayloadSize)

	for tries := 0; tries < maxTries; tries++ {
		if tries == maxTries-1 {
			return nil, ErrFailedAfterRetries
		}

		deadline, ok := writeCtx.Deadline()
		if ok {
			if err := conn.SetReadDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return nil, ErrCtxTimeOut
				}
				return nil, ErrDeadlineUnsuccessful
			}
		}

		n, err := conn.Read(buf)
		// Write any data we got before handling errors
		if n > 0 {
			data.Write(buf[:n])
			fmt.Println("data len", len(data.Bytes()))
		}

		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return nil, ErrCtxTimeOut
			}
			if errors.Is(err, io.EOF) {
				return nil, ErrConnectionClosed
			}

			retryTime := time.Second * (1 << uint(tries))
			select {
			case <-writeCtx.Done():
				if errors.Is(writeCtx.Err(), context.DeadlineExceeded) {
					return nil, writeCtx.Err()
				}
			case <-time.After(retryTime):
				log.Printf("retrying connection to server...")
				continue
			}
		}

		// If we got here with no error, we have a complete read
		conn.SetReadDeadline(time.Time{}) // Reset the deadline
		return data.Bytes(), nil
	}
	return nil, ErrFailedAfterRetries
}
