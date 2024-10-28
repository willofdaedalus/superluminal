// NOTE; https://stackoverflow.com/a/49580791
// TODO; add a queueing system that keeps track of headers sent and their appropriate responses
// to remove manual check all the time

package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"willofdaedalus/superluminal/client"
	u "willofdaedalus/superluminal/utils"

	"github.com/google/uuid"
)

const (
	debugPassCount  = 1
	actualPassCount = 6
)

const (
	maxTries  = 3
	passTries = 3
)

type Server struct {
	OwnerName string
	listener  net.Listener
	masterID  string
	clients   map[string]*client.Client
	ip, port  string

	pass         string
	currentHash  string
	passRotated  bool
	timeStarted  time.Time
	outgoingChan chan []byte
	sigChan      chan os.Signal
	signals      []os.Signal
	hashTimeout  time.Duration
}

func CreateServer(owner string, maxConns int) (*Server, error) {
	clients := make(map[string]*client.Client, maxConns)
	signals := []os.Signal{syscall.SIGINT, syscall.SIGKILL}

	pass, hash, err := genPassAndHash()
	if err != nil {
		return nil, err
	}
	log.Println("your pass is", pass)

	// listen on tcp port 42024 on all available p addresses of the local system.
	// godocs recommeds not assigning a host https://pkg.go.dev/net#Listen
	listener, err := net.Listen("tcp", ":42024")
	if err != nil {
		return nil, err
	}

	// TODO;the user gets to input their name here in the future
	masterClient := client.CreateClient(owner, "")
	masterID := uuid.NewString()
	clients[masterID] = masterClient

	return &Server{
		OwnerName:   masterClient.Name,
		listener:    listener,
		clients:     clients,
		signals:     signals,
		masterID:    masterID,
		timeStarted: time.Now(),
		hashTimeout: time.Minute * 5,
		currentHash: hash,
	}, nil
}

// generate a new passphrase, hash it and return the duo
func genPassAndHash() (string, string, error) {
	pass, err := u.GeneratePassphrase(debugPassCount)
	if err != nil {
		return "", "", err
	}

	hash, err := u.HashPassphrase(pass)
	if err != nil {
		return "", "", err
	}

	return pass, hash, nil
}

func (s *Server) Start() {
	// cancel to propagate shutdown signal to all child contexts
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		s.ShutdownServer()
	}()

	s.sigChan = make(chan os.Signal, 1)
	signal.Notify(s.sigChan, s.signals...)
	// regenerate the hash and passphrase every hashTimeout minutes
	go func() {
		for range time.Tick(s.hashTimeout) {
			s.pass, s.currentHash, _ = genPassAndHash()
			if !s.passRotated {
				s.passRotated = true
			}

			// obviously remove this
			fmt.Println("your new pass:", s.pass)
		}
	}()

	go func() {
		for {
			log.Println("waiting for someone to join...")
			conn, err := s.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}

				log.Println("err:", err)
				continue
			}

			if len(s.clients) >= maxClients {
				u.TryWrite(ctx, conn, maxTries, []byte(u.HdrErrMsg), []byte("server_full"))
				conn.Close()
				continue
			}

			go func(conn net.Conn) {
				errChan := make(chan error)
				go s.handleNewClient(ctx, conn, errChan)

				select {
				case err := <-errChan:
					if err != nil {
						// log.Println("client err: ", err)
						s.sigChan <- syscall.SIGINT
					}
					return
				case <-ctx.Done():
				}
				conn.Close()
			}(conn)
		}
	}()
	<-s.sigChan
	cancel()
}

func (s *Server) ShutdownServer() {
	// add code to close all other clients
	log.Println("shutting down server now")
	s.listener.Close()
}

func (s *Server) handleNewClient(ctx context.Context, conn net.Conn, errChan chan<- error) {
	var client client.Client
	var clientData bytes.Buffer
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			errChan <- u.ErrCtxTimeOut
			return
		}
	}()

	// send the server identification message to the client
	if err := s.sendServerAuth(ctx, conn); err != nil {
		errChan <- err
		return
	}

	// read the information from the client
	data, err := u.TryRead(ctx, conn, maxTries)
	if err != nil {
		errChan <- err
		return
	}

	// NOTE; this is a very important func for reacting to headers
	_, msg, err := u.ParseIncomingMsg(data)
	if err != nil {
		errChan <- err
		return
	}

	clientData.Write(msg)
	dec := gob.NewDecoder(&clientData)
	if err := dec.Decode(&client); err != nil {
		errChan <- fmt.Errorf("couldn't decode values")
		return
	}

	log.Printf("%s", client.Name)
	return
}

// TODO; current ctx runs for about 5s and we can always increment that duration BUT
// what if the client doesn't enter a pass as fast as possible and the server kicks
// them for timeout;
// POSSIBLE SOLUTIONS
// * once the conn to the server is dead, upon entering a passphrase (any) try reconning
// to the server and submit the same details for verification; this time around the counter
// has reset to 0 (THAT COULD BE A MASSIVE EXPLOIT!)
// * don't run the ctx until user verifies code but that leaves the server's port waiting
// which means we need a maximum server open window which the ctx solves
func (s *Server) tryValidatePass(clientPass string) error {
	for tries := 0; tries < passTries; tries++ {
	}
	return nil
}

func (s *Server) sendServerAuth(ctx context.Context, conn net.Conn) error {
	for tries := 0; tries < maxTries; tries++ {
		select {
		case <-ctx.Done():
			return u.ErrCtxTimeOut
		default:
			if err := u.TryWrite(ctx, conn, maxTries, []byte(u.HdrAckMsg), []byte(u.AckSelfReport)); err != nil {
				// in the event the context timed out before client got server auth
				// return that custom error we returned from sendMessage
				if errors.Is(err, u.ErrCtxTimeOut) {
					return err
				}

				if tries == maxTries-1 {
					return u.ErrClientExchFailed
				}

				retryTime := time.Millisecond * 500 * (1 << uint(tries))
				// check on the context while counting down to retry the connection
				select {
				case <-ctx.Done():
					return u.ErrCtxTimeOut
				case <-time.After(retryTime):
					log.Println("failed to send message. retrying...")
					continue
				}
			}
			log.Println("sent server authentication to client; waiting for client details")
			return nil
		}
	}

	return u.ErrFailedServerAuth
}
