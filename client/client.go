package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
	"willofdaedalus/superluminal/utils"
)

const (
	maxConnTries = 3
)

type Client struct {
	Name     string
	PassUsed string
}

func CreateClient(name string) *Client {
	return &Client{
		Name: name,
	}
}

func (c *Client) ConnectToServer(host, port string) error {
	buf := make([]byte, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// to get the dialwithctx func
	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		buf, err = tryRead(ctx, conn)
		if err != nil {
			// we assume we couldn't read from the server then
			return err
		}
		log.Println(string(buf))
	}
}

func tryRead(ctx context.Context, conn net.Conn) ([]byte, error) {
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
					// Don't retry on timeout
					buf <- nil
					errChan <- utils.ErrCtxTimeOut
					return
				}

				// Check context before sleeping
				select {
				case <-readCtx.Done():
					buf <- nil
					errChan <- readCtx.Err()
					return
				case <-time.After(time.Millisecond * 500):
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

		// Max retries reached
		buf <- nil
		errChan <- fmt.Errorf("failed to read after %d attempts", maxConnTries)
	}()

	// wait for either read completion or context cancellation
	select {
	case <-readCtx.Done():
		// // note: we still need to receive from channels to prevent goroutine leak
		// go func() {
		<-buf
		<-errChan
		// }()
		return nil, fmt.Errorf("read operation cancelled: %w", readCtx.Err())
	case data := <-buf:
		err := <-errChan
		return data, err
	}
}

// func try(ctx context.Context, conn net.Conn) ([]byte, error) {
// 	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
// 	defer cancel()

// 	buf := make(chan []byte)
// 	errChan := make(chan error)
// 	go func() {
// 		var err error

// 		successfulRead := false
// 		readBuf := make([]byte, 1024)
// 		n := 0
// 		for tries := 0; tries < maxConnTries; tries++ {
// 			n, err = conn.Read(readBuf)
// 			if err != nil {
// 				if errors.Is(err, io.EOF) {
// 					// server shutdown
// 					buf <- nil
// 					return
// 				}
// 				errChan <- err
// 				log.Println("err:", err)
// 				time.Sleep(time.Millisecond * 500)
// 				continue
// 			}
// 			successfulRead = true
// 		}
// 		if !successfulRead {
// 			buf <- nil
// 		}

// 		buf <- readBuf[:n]
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		// context is done
// 		buf <- nil
// 		errChan <- utils.ErrCtxTimeOut
// 	}

// 	return <-buf, <-errChan
// }
