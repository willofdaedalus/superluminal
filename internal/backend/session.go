package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/payload/info"
	"willofdaedalus/superluminal/internal/utils"

	err1 "willofdaedalus/superluminal/internal/payload/error"
)

const (
	debugPassCount = 1
	MaxAuthChances = 3

	heartbeatTimeout  = time.Second * 30
	clientKickTimeout = time.Second * 30
	maxHandleTime     = time.Minute * 1
	passRegenTimeout  = time.Minute * 5
)

func NewSession(owner string, maxConns uint8) (*session, error) {
	var reader bytes.Reader
	clients := make(map[string]*sessionClient, maxConns)

	listener, err := net.Listen("tcp", ":42024")
	if err != nil {
		return nil, err
	}

	pass, hash, err := genPassAndHash(debugPassCount)
	if err != nil {
		return nil, err
	}
	log.Println("your pass is", pass)

	master := createClient(owner, nil, true)
	clients[master.uuid] = master

	return &session{
		Owner:         owner,
		maxConns:      maxConns + 1,
		clients:       clients,
		listener:      listener,
		pass:          pass,
		hash:          hash,
		reader:        reader,
		heartbeatTime: heartbeatTimeout,
		passRegenTime: passRegenTimeout,
	}, nil
}

func (s *session) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	for range time.Tick(s.passRegenTime) {
		s.pass, s.hash, _ = genPassAndHash(debugPassCount)
		fmt.Println(s.pass)
	}

	go s.listen(ctx, doneChan, errChan)

	// Wait for and handle errors
	select {
	case <-doneChan:
		return nil
	}
}

func (s *session) listen(ctx context.Context, doneChan chan<- struct{}, errChan chan error) {
	log.Println("server started...")
	defer close(doneChan)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			errChan <- fmt.Errorf("accept error: %w", err)
			return
		}

		// Use a buffered channel to manage concurrency
		go func(conn net.Conn) {
			if len(s.clients) >= int(s.maxConns) {
				errorMsg := base.GenerateError(
					err1.ErrorMessage_ERROR_SERVER_FULL,
					[]byte("server_full"),
					[]byte("server is full"),
				)

				errPayload, err := base.EncodePayload(common.Header_HEADER_ERROR, errorMsg)
				if err != nil {
					conn.Close()
					return
				}

				// need a minute to write to the client; if it's not possible don't bother
				tempCtx, tempCancel := context.WithTimeout(ctx, clientKickTimeout)
				err = utils.TryWriteCtx(tempCtx, conn, errPayload)
				tempCancel()
				if err != nil {
					if errors.Is(err, io.EOF) {
						log.Println("sprlmnl: client is closed")
					} else {
						log.Printf("write error: %v", err)
					}
				}

				conn.Close()
				log.Println("rejected client with server_full error")
				return
			}

			if err := s.handleNewConn(ctx, conn); err != nil {
				errChan <- err
			}
			errChan <- nil
		}(conn)

		select {
		case err := <-errChan:
			if err != nil {
				log.Println("err from client: ", err)
			}
			continue
		}
	}

}

func (s *session) kickAndCloseClient(ctx context.Context, errType err1.ErrorMessage_ErrorCode, conn net.Conn) {
	kickCtx, cancel := context.WithTimeout(ctx, clientKickTimeout)
	defer func() {
		cancel()
		conn.Close()
	}()

	errPayload := base.GenerateError(errType, []byte("failed_auth"), []byte("failed to authenticate with session"))
	payload, err := base.EncodePayload(common.Header_HEADER_ERROR, errPayload)
	if err != nil {
		return
	}

	err = utils.TryWriteCtx(kickCtx, conn, payload)
	if err != nil {
		return
	}
}

func (s *session) End() {
	for k := range s.clients {
		delete(s.clients, k)
	}
	s.listener.Close()
}

func (s *session) handleNewConn(ctx context.Context, conn net.Conn) error {
	name, err := s.authenticateClient(ctx, conn)
	if err != nil {
		s.kickAndCloseClient(ctx, err1.ErrorMessage_ERROR_AUTH_FAILED, conn)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	newClient := createClient(name, conn, false)
	s.clients[newClient.uuid] = newClient

	// send a congratulatory message to the client
	infoPayload := base.GenerateInfo(info.Info_INFO_AUTH_SUCCESS, "welcome to the session")
	payload, err := base.EncodePayload(common.Header_HEADER_INFO, infoPayload)
	if err != nil {
		return fmt.Errorf("failed to encode welcome payload: %w", err)
	}

	err = utils.TryWriteCtx(ctx, conn, payload)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("client disconnected: %w", err)
		}
		return fmt.Errorf("failed to send welcome message: %w", err)
	}

	log.Println("hello client", newClient.uuid)
	return nil
}
