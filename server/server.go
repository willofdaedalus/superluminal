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
	maxTries   = 3
	maxClients = 32
)

type Server struct {
	listener net.Listener
	masterID string
	clients  map[string]*client.Client
	ip, port string

	msgStack     *stack
	timeStarted  time.Time
	outgoingChan chan []byte
	sigChan      chan os.Signal
	signals      []os.Signal
}

func CreateServer() (*Server, error) {
	clients := make(map[string]*client.Client, maxClients)
	signals := []os.Signal{syscall.SIGINT, syscall.SIGKILL}

	st := newStack()

	// listen on tcp port 42024 on all available p addresses of the local system.
	// godocs recommeds not assigning a host https://pkg.go.dev/net#Listen
	listener, err := net.Listen("tcp", ":42024")
	if err != nil {
		return nil, err
	}

	// TODO;the user gets to input their name here in the future
	masterClient := client.CreateClient("master")
	masterID := uuid.NewString()
	clients[masterID] = masterClient

	return &Server{
		listener: listener,
		clients:  clients,
		signals:  signals,
		msgStack: st,
	}, nil
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

			go func(conn net.Conn) {
				errChan := make(chan error)
				go s.handleNewClient(ctx, conn, errChan)

				select {
				case err := <-errChan:
					if err != nil {
						log.Println("client err: ", err)
						// TODO; THIS IS JUST FOR DEBUG PURPOSES; REMOVE WHEN DONE!
						s.sigChan <- syscall.SIGINT
					}
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
	// return
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

	// put the recent hdr on list for cross checking when a response is received
	s.msgStack.push(u.HdrAckVal)

	return u.ErrFailedServerAuth
}
