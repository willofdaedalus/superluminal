package server

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"willofdaedalus/superluminal/client"
	"willofdaedalus/superluminal/utils"

	"github.com/google/uuid"
)

const (
	maxConnTries = 3
)

type Server struct {
	listener     net.Listener
	masterID     string
	clients      map[string]*client.Client
	ip, port     string
	timeStarted  time.Time
	outgoingChan chan []byte

	sigChan chan os.Signal
	signals []os.Signal
}

func CreateServer() (*Server, error) {
	clients := make(map[string]*client.Client)
	signals := []os.Signal{syscall.SIGINT, syscall.SIGKILL}

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
	}, nil
}

func (s *Server) Start() {
	// cancel to propagate shutdown signal to all child contexts
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
				go handleNewClient(ctx, conn, errChan)

				select {
				case err := <-errChan:
					if err != nil {
						log.Println("client err: ", err)
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
	s.ShutdownServer()
}

func (s *Server) ShutdownServer() {
	// add code to close all other clients
	log.Println("shutting down server now")
	s.listener.Close()
}

func handleNewClient(ctx context.Context, conn net.Conn, errChan chan<- error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// send the server identification message to the client
	if err := sendServerAuth(ctx, conn); err != nil {
		errChan <- err
		return
	}

	select {
	case <-ctx.Done():
		errChan <- utils.ErrCtxTimeOut
		return
	}
}

func sendServerAuth(ctx context.Context, conn net.Conn) error {
	for tries := 0; tries < maxConnTries; tries++ {
		select {
		case <-ctx.Done():
			return utils.ErrCtxTimeOut
		default:
			if err := sendMessage(ctx, conn, utils.HdrInfo, "superluminal_server"); err != nil {
				// in the event the context timed out before client got server auth
				// return that custom error we returned from sendMessage
				if errors.Is(err, utils.ErrCtxTimeOut) {
					return err
				}

				if tries == maxConnTries-1 {
					return utils.ErrClientExchFailed
				}

				// check on the context while counting down to retry the connection
				select {
				case <-ctx.Done():
					return utils.ErrCtxTimeOut
				case <-time.After(time.Millisecond * 500):
					log.Println("failed to send message. retrying...")
					continue
				}
			}
			log.Println("sent server authentication to client; waiting for client details")
			return nil
		}
	}

	return utils.ErrFailedServerAuth
}
