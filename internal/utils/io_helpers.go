package utils

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	// MaxConnTime    = time.Second * 30
	MaxConnTime    = time.Minute * 1
	MaxBackoffTime = time.Second * 7
	maxTries       = 5
	baseBackoff    = 100 * time.Millisecond
)

// TryWriteCtx attempts to write to the conn passed and retries up to a number of times
// defined until it gives up and returns an error
// it respects the context passed to it
func TryWriteCtx(ctx context.Context, conn net.Conn, data []byte) error {
	if conn == nil {
		return fmt.Errorf("nil connection")
	}
	if len(data) == 0 {
		return nil // Nothing to write
	}

	// Ensure deadline is reset at the end
	defer conn.SetWriteDeadline(time.Time{})

	for tries := 0; tries < maxTries; tries++ {
		// Check context before attempting write
		select {
		case <-ctx.Done():
			return fmt.Errorf("write cancelled: %w", ctx.Err())
		default:
		}

		// Set deadline from context
		if deadline, ok := ctx.Deadline(); ok {
			if err := conn.SetWriteDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return ErrCtxTimeOut
				}
				return fmt.Errorf("set deadline failed: %w", err)
			}
		}

		// attempt write
		n, err := conn.Write(data)
		if err == nil && n == len(data) {
			return nil
		}

		// handle write errors
		if err != nil {
			// non-retryable errors
			if errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded) {
				return err
			}

			// for partial writes, adjust data slice
			if n > 0 {
				data = data[n:]
			}

			// only retry if we haven't exhausted our attempts
			if tries == maxTries-1 {
				return fmt.Errorf("failed after %d retries: %w", maxTries, err)
			}

			// calculate backoff with jitter
			backoff := baseBackoff * time.Duration(1<<uint(tries))
			if backoff > MaxBackoffTime {
				backoff = MaxBackoffTime
			}
			jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
			backoff = backoff + jitter

			// wait for backoff period or context cancellation
			log.Printf("retrying write after error: %v (try %d/%d, waiting %v)",
				err, tries+1, maxTries, backoff)

			select {
			case <-ctx.Done():
				return fmt.Errorf("write cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
				continue
			}
		}
	}

	return ErrFailedAfterRetries
}

func ReadFull(ctx context.Context, conn net.Conn, tracker *SyncTracker) ([]byte, error) {
	tracker.IncrementRead()
	defer tracker.DecrementRead()

	done := make(chan struct{})
	var result []byte
	var readErr error

	go func() {
		defer close(done)

		// set the read deadline
		deadline, ok := ctx.Deadline()
		if ok {
			if err := conn.SetReadDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					readErr = ErrCtxTimeOut
					return
				}
				readErr = ErrSetDeadlineUnsuccessful
				return
			}
		}

		// read initial header
		headerBuffer := make([]byte, 4)
		n, err := io.ReadFull(conn, headerBuffer)
		if err != nil {
			log.Printf("Failed to read header: error=%v, bytes_read=%d", err, n)
			readErr = io.ErrUnexpectedEOF
			return
		}

		payloadLen := binary.BigEndian.Uint32(headerBuffer)
		// sanity check on payload length
		if payloadLen > MaxPayloadSize {
			readErr = fmt.Errorf("payload length exceeds maximum allowed size: %d", payloadLen)
			return
		}

		// allocate space for the full payload
		actualPayload := make([]byte, payloadLen)
		// use io.readfull to read the entire payload
		if _, err := io.ReadFull(conn, actualPayload); err != nil {
			readErr = io.ErrUnexpectedEOF
			return
		}

		// reset the read deadline
		conn.SetReadDeadline(time.Time{})
		result = actualPayload
	}()

	// wait for either the context to be done or the read to complete
	select {
	case <-ctx.Done():
		conn.SetReadDeadline(time.Now())
		<-done
		return nil, ctx.Err()
	case <-done:
		// the goroutine has naturally exited so we can return whatever we
		// got from the assignments and everything
		return result, readErr
	}
}

func WriteFull(ctx context.Context, conn net.Conn, tracker *SyncTracker, data []byte) error {
	if tracker != nil {
		tracker.IncrementWrite()
		defer tracker.DecrementWrite()
	}

	payload := PrependLength(data)
	return TryWriteCtx(ctx, conn, payload)
}
