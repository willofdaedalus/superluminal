package backend

import (
	"bytes"
	"context"
	"errors"
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
	DebugPassCount = 1
	MaxHandleTime  = time.Minute * 1
	MaxAuthChances = 3
)

func NewSession(owner string, maxConns uint8) (*session, error) {
	var reader bytes.Reader
	clients := make(map[string]*sessionClient, maxConns)

	listener, err := net.Listen("tcp", ":42024")
	if err != nil {
		return nil, err
	}

	pass, hash, err := genPassAndHash(DebugPassCount)
	if err != nil {
		return nil, err
	}
	log.Println("your pass is", pass)

	master := createClient(owner, nil, true)
	clients[master.uuid] = master

	return &session{
		Owner:    owner,
		maxConns: maxConns + 1,
		clients:  clients,
		listener: listener,
		pass:     pass,
		hash:     hash,
		reader:   reader,
	}, nil
}

func (s *session) Start() {
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		close(errChan)
	}()

	log.Println("server started...")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}

		// send an error message to the client and close the connection
		if len(s.clients) >= int(s.maxConns) {
			errorMsg := base.GenerateError(
				err1.ErrorMessage_ERROR_SERVER_FULL,
				[]byte("server_full"),
				[]byte("server is full"),
			)

			errPayload, err := base.EncodePayload(common.Header_HEADER_ERROR, errorMsg)
			if err != nil {
				conn.Close()
				continue
			}

			err = utils.TryWriteCtx(ctx, conn, errPayload)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Println("sprlmnl: client is closed")
				}

				log.Printf("%v", err)
			}
			conn.Close()
			log.Println("rejected client with server_full error")
			continue
		}

		go s.handleNewConn(ctx, conn, errChan)

		go func() {
			for {
				select {
				case err := <-errChan:
					if err != nil {
						if errors.Is(err, utils.ErrFailedServerAuth) {
							s.kickAndCloseClient(err1.ErrorMessage_ERROR_AUTH_FAILED, conn)
						}
						log.Println(err.Error())
					}
				}
			}
		}()
	}
}

func (s *session) kickAndCloseClient(errType err1.ErrorMessage_ErrorCode, conn net.Conn) {
	defer conn.Close()
	errPayload := base.GenerateError(errType, []byte("failed_auth"), []byte("failed to authenticate with session"))
	payload, err := base.EncodePayload(common.Header_HEADER_ERROR, errPayload)
	if err != nil {
		return
	}

	err = utils.TryWriteCtx(context.Background(), conn, payload)
	if err != nil {
		return
	}
}

func (s *session) handleNewConn(ctx context.Context, conn net.Conn, errChan chan<- error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// handleCtx, cancel := context.WithTimeout(ctx, MaxHandleTime)
	// defer cancel()

	name, err := s.authenticateClient(ctx, conn)
	if err != nil {
		errChan <- err
		return
	}

	newClient := createClient(name, conn, false)
	s.clients[newClient.uuid] = newClient

	// send a congratulatory message to the client
	infoPayload := base.GenerateInfo(info.Info_INFO_AUTH_SUCCESS, "welcome to the session")
	payload, err := base.EncodePayload(common.Header_HEADER_INFO, infoPayload)
	if err != nil {
		errChan <- err
		return
	}

	err = utils.TryWriteCtx(ctx, conn, payload)
	if err != nil {
		if errors.Is(err, io.EOF) {
			errChan <- err
			return
		}
		return
	}
}

func (s *session) End() {
	for k := range s.clients {
		delete(s.clients, k)
	}
	s.listener.Close()
}
