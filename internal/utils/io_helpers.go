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
	// this is very generous and is half of the usual and original ctx that
	// is passed from handleNewClient in server.go
	MaxConnTime    = time.Second * 30
	MaxBackoffTime = time.Second * 7
	MaxTries       = 5
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
	writeCtx, cancel := context.WithTimeout(ctx, MaxConnTime)
	defer cancel()

	for tries := 0; tries < MaxTries; tries++ {
		deadline, ok := writeCtx.Deadline()
		if ok {
			if err := conn.SetWriteDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return ErrCtxTimeOut
				}

				return ErrDeadlineUnsuccessful
			}
		}

		if tries == MaxTries-1 {
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

func TryReadCtx(ctx context.Context, conn net.Conn) ([]byte, error) {
	var data bytes.Buffer
	writeCtx, cancel := context.WithTimeout(ctx, MaxConnTime)
	defer cancel()

	buf := make([]byte, MaxPayloadSize)
	for tries := 0; tries < MaxTries; tries++ {
		deadline, ok := writeCtx.Deadline()
		if ok {
			if err := conn.SetReadDeadline(deadline); err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return nil, ErrCtxTimeOut
				}

				return nil, ErrDeadlineUnsuccessful
			}
		}

		if tries == MaxTries-1 {
			return nil, ErrFailedAfterRetries
		}

		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return nil, ErrCtxTimeOut
			}

			if errors.Is(err, io.EOF) {
				return nil, ErrServerClosed
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

		// Successful read
		data.Write(buf[:n])
		conn.SetReadDeadline(time.Time{}) // Reset the deadline
		return data.Bytes(), nil
	}

	return nil, ErrFailedAfterRetries
}

func TryWrite(ctx context.Context, ws *WriteStruct) error {
	writeCtx, cancel := context.WithTimeout(ctx, MaxConnTime)
	defer cancel()

	endChan := make(chan error, 1)
	finalMsg := bytes.Join([][]byte{ws.headerMsgByte(), ws.Message}, []byte(":"))

	go func() {
		for tries := 0; tries < ws.MaxTries; tries++ {
			deadline, ok := writeCtx.Deadline()
			if ok {
				if err := ws.Conn.SetWriteDeadline(deadline); err != nil {
					endChan <- ErrDeadlineUnsuccessful
					return
				}
			}

			_, err := ws.Conn.Write(finalMsg)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					endChan <- ErrCtxTimeOut
					return
				}

				if errors.Is(err, io.EOF) {
					endChan <- io.EOF
					return
				}

				retryTime := time.Second * 1 * (1 << uint(tries))
				select {
				case <-writeCtx.Done():
					if errors.Is(writeCtx.Err(), context.DeadlineExceeded) {
						endChan <- writeCtx.Err()
						return
					}
				case <-time.After(retryTime):
					log.Printf("retrying connection to server again...")
					continue
				}
			}
			endChan <- nil
			return
		}
	}()

	return <-endChan
}

func TryRead(ctx context.Context, conn net.Conn, maxConnTries int) ([]byte, error) {
	readCtx, cancel := context.WithTimeout(ctx, MaxConnTime)
	defer cancel()

	buf := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			close(buf)
			close(errChan)
		}()

		readBuf := make([]byte, 1024)

		for tries := 0; tries < maxConnTries; tries++ {
			// set read deadline based on context
			deadline, ok := readCtx.Deadline()
			if ok {
				if err := conn.SetReadDeadline(deadline); err != nil {
					errChan <- fmt.Errorf("failed to set read deadline: %w", err)
					return
				}
			}

			n, err := conn.Read(readBuf)

			// reset read deadline; it persists if we don't reset and it's best to
			// do it immediately
			if err := conn.SetReadDeadline(time.Time{}); err != nil {
				log.Printf("warning: failed to reset read deadline: %v", err)
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					buf <- nil
					errChan <- io.EOF
					return
				}

				if errors.Is(err, os.ErrDeadlineExceeded) {
					// don't retry on timeout
					buf <- nil
					errChan <- ErrCtxTimeOut
					return
				}

				// check context before sleeping
				retryTime := time.Second * 1 * (1 << uint(tries))
				select {
				case <-readCtx.Done():
					if errors.Is(readCtx.Err(), context.DeadlineExceeded) {
						buf <- nil
						errChan <- readCtx.Err()
						return
					}
				case <-time.After(retryTime):
					log.Printf("retrying connection to server again...")
					continue
				}
			}

			// successful read
			data := make([]byte, n)
			copy(data, readBuf[:n])
			buf <- data
			errChan <- nil
			return
		}

		// max retries reached
		buf <- nil
		errChan <- fmt.Errorf("failed to read after %d attempts", maxConnTries)
	}()

	// wait for either read completion or context cancellation
	select {
	case <-readCtx.Done():
		<-buf
		<-errChan
		return nil, fmt.Errorf("read operation cancelled: %w", readCtx.Err())
	case data := <-buf:
		err := <-errChan
		return data, err
	}
}

// func IntToBytes(buf []byte) []byte {

// }

// func BytesToInt(buf []byte) int {

// }
