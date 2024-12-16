package utils

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const (
	// MaxConnTime    = time.Second * 30
	MaxConnTime    = time.Minute * 1
	MaxBackoffTime = time.Second * 7
	maxTries       = 5
)

// TryWriteCtx attempts to write to the conn passed and retries up to a number of times
// defined until it gives up and returns an error
// it respects the context passed to it
func TryWriteCtx(ctx context.Context, conn net.Conn, data []byte) error {
	for tries := 0; tries < maxTries; tries++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// proceed with the write
		}

		// set deadline from context
		if deadline, ok := ctx.Deadline(); ok {
			if err := conn.SetWriteDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return ErrCtxTimeOut
				}
				return ErrDeadlineUnsuccessful
			}
		}

		// handle partial writes
		bytesWritten := 0
		for bytesWritten < len(data) {
			n, err := conn.Write(data[bytesWritten:])
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return ErrCtxTimeOut
				}
				if errors.Is(err, io.EOF) {
					return io.EOF
				}

				// retry with exponential backoff
				retryTime := time.Second * (1 << uint(tries))
				if retryTime > MaxBackoffTime {
					retryTime = MaxBackoffTime
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryTime):
					log.Printf("retrying connection after error: %v (try %d/%d)", err, tries+1, maxTries)
					break // retry outer loop
				}
			}
			bytesWritten += n
		}

		// successful write
		conn.SetWriteDeadline(time.Time{})
		return nil
	}

	return ErrFailedAfterRetries
}

// TryReadCtx relies on external timeouts and deadlines in order to function
// properly. this makes it robust for all situations including those that
// are not in any hurry to read something from a connection
func TryReadCtx(ctx context.Context, conn net.Conn) ([]byte, error) {
	var data bytes.Buffer
	buf := make([]byte, MaxPayloadSize)
	for tries := 0; tries < maxTries; tries++ {
		// Exit if the context is canceled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// continue to attempt reading
		}

		// set the deadline based on the context
		deadline, ok := ctx.Deadline()
		if ok {
			if err := conn.SetReadDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return nil, ErrCtxTimeOut
				}
				return nil, ErrDeadlineUnsuccessful
			}
		}

		// attempt to read from the connection
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return nil, err
			}
			if errors.Is(err, io.EOF) {
				return nil, ErrConnectionClosed
			}
			// backoff before retrying
			retryTime := time.Second * (1 << uint(tries))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryTime):
				continue
			}
		}

		// only write and return if we actually read something
		if n > 0 {
			data.Write(buf[:n])
			conn.SetReadDeadline(time.Time{})
			home, _ := os.UserHomeDir()
			LogBytes("read", home+"/superluminal.log", data.Bytes())
			return data.Bytes(), nil
		}
	}
	return nil, ErrFailedAfterRetries
}
