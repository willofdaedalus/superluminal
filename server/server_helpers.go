package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"willofdaedalus/superluminal/utils"
)

func sendMessage(ctx context.Context, conn net.Conn, header, msg string) error {
	finalMessage := []byte(fmt.Sprintf("%s%s", header, msg))

	errChan := make(chan error)
	defer close(errChan)
	go func() {
		_, err := conn.Write(finalMessage)
		errChan <- err
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return utils.ErrCtxTimeOut
	}
}

func (s *Server) handleSignals(outSig chan<- bool) {
	s.sigChan = make(chan os.Signal)

	defer func() {
		signal.Stop(s.sigChan)
		close(s.sigChan)
	}()

	for range s.sigChan {
		outSig <- true
		return
	}
}
