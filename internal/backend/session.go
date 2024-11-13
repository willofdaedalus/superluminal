package backend

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"

	err1 "willofdaedalus/superluminal/internal/payload/error"
)

const (
	debugPassCount = 1
)

func NewSession(owner string, maxConns uint8) (*Session, error) {
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

	master := createClient(owner, "", nil, true)
	clients[master.uuid] = master

	return &Session{
		Owner:    owner,
		maxConns: maxConns + 1,
		clients:  clients,
		listener: listener,
		pass:     pass,
		hash:     hash,
		reader:   reader,
	}, nil
}

func (s *Session) Start() {
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

			_, err = conn.Write(errPayload)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Println("sprlmnl: client is closed")
				}
			}
			conn.Close()
			continue
		}
	}
}

func (s *Session) End() {
	for k := range s.clients {
		delete(s.clients, k)
	}
	s.listener.Close()
}
