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
	"time"
)

func TryWrite(ctx context.Context, conn net.Conn, maxConnTries int, header, message []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	endChan := make(chan error, 1)
	finalMsg := bytes.Join([][]byte{header, message}, []byte(""))

	go func() {
		for tries := 0; tries < maxConnTries; tries++ {
			deadline, ok := writeCtx.Deadline()
			if ok {
				if err := conn.SetWriteDeadline(deadline); err != nil {
					endChan <- ErrDeadlineUnsuccessful
					return
				}
			}

			_, err := conn.Write(finalMsg)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					endChan <- ErrCtxTimeOut
					return
				}

				if errors.Is(err, io.EOF) {
					endChan <- io.EOF
					return
				}

				retryTime := time.Millisecond * 500 * (1 << uint(tries))
				select {
				case <-writeCtx.Done():
					endChan <- writeCtx.Err()
					return
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
	readCtx, cancel := context.WithTimeout(ctx, time.Second*5)
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
				retryTime := time.Millisecond * 500 * (1 << uint(tries))
				select {
				case <-readCtx.Done():
					buf <- nil
					errChan <- readCtx.Err()
					return
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

// take some bytes and return their equivalent value index
func GetHeaderType(header []byte) int {
	switch string(header) {
	case "inf":
		return HdrInfoVal
	case "ack":
		return HdrAckVal
	case "err":
		return HdrErrVal
	case "res":
		return HdrResVal
	default:
		return HdrUnknownVal
	}
}

func ParseHeader(header []byte) (int, error) {
	if len(header) != 11 {
		return HdrUnknownVal, ErrInvalidHeader
	}

	split := bytes.Split(header, []byte("+"))
	if !bytes.Contains(split[0], []byte("sprlmnl")) || len(split[1]) != 3 {
		return HdrUnknownVal, ErrInvalidHeader
	}

	log.Println(string(split[1]))
	headerType := GetHeaderType(split[1])
	if headerType == HdrUnknownVal {
		return HdrUnknownVal, ErrUnknownHeader
	}

	return headerType, nil
}

func ParseIncomingMsg(msg []byte) (int, []byte, error) {
	header, message, ok := bytes.Cut(msg, []byte(":"))
	if !ok {
		return HdrUnknownVal, nil, ErrInvalidHeader
	}

	headerType, err := ParseHeader(header)
	if err != nil {
		return headerType, nil, err
	}

	// switch headerType {
	// case ack:
	// 	return c.parseAckMessage(message)
	// case info:
	// 	log.Printf("%s", message)
	// 	return nil
	// }

	return headerType, message, nil
}
