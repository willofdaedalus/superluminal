package server

import (
	"context"
	"fmt"
	"net"
	"time"
	"willofdaedalus/superluminal/utils"
)

func sendMessage(ctx context.Context, conn net.Conn, header, msg string) error {
	finalMessage := []byte(fmt.Sprintf("%s%s", header, msg))

	select {
	case <-ctx.Done():
		return utils.ErrCtxTimeOut
	default:
		deadline, ok := ctx.Deadline()
		if ok {
			// set a write deadline that is in sync with the ctx's
			if err := conn.SetWriteDeadline(deadline); err != nil {
				return fmt.Errorf("failed to write deadline", err)
			}
			defer conn.SetWriteDeadline(time.Time{})
		}

		errChan := make(chan error)
		defer close(errChan)
		go func() {
			_, err := conn.Write(finalMessage)
			errChan <- err
		}()
		return <-errChan
	}
}
