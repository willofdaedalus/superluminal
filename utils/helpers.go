package utils

import (
	"context"
	"fmt"
	"net"
)

func SendMessage(ctx context.Context, conn net.Conn, msgHdr string, msg string, retErr error) error {
	errChan := make(chan error, 1)
	go func() {
		msg := fmt.Sprintf("%s%s", msgHdr, msg)
		_, err := conn.Write([]byte(msg))
		errChan <- err
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("couldn't send header: %v", err)
		}
		return nil
	case <-ctx.Done():
		return retErr
	}
}
