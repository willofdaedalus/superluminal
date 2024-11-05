// NOTE; https://stackoverflow.com/a/49580791
// TODO; add a queueing system that keeps track of headers sent and their appropriate responses
// to remove manual check all the time

package backend

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
	"sync"
	"syscall"
	"time"
	"willofdaedalus/superluminal/internal/client"
	"willofdaedalus/superluminal/internal/pipeline"
	u "willofdaedalus/superluminal/internal/utils"

	"github.com/google/uuid"
)

const (
	debugPassCount  = 1
	actualPassCount = 6
)

const (
	maxTries   = 3
	passTries  = 3
	maxTimeout = time.Minute * 1
)

type Server struct {
	OwnerName string

	listener     net.Listener
	masterID     string
	clients      map[string]*client.Client
	ip, port     string
	pipeline     *pipeline.Pipeline
	pass         string
	currentHash  string
	passRotated  bool
	timeStarted  time.Time
	outgoingChan chan []byte
	sigChan      chan os.Signal
	signals      []os.Signal
	hashTimeout  time.Duration
	maxClients   int
	mu           sync.Mutex
}

func CreateServer(owner string, maxConns int) (*Server, error) {
	clients := make(map[string]*client.Client, maxConns)
	signals := []os.Signal{syscall.SIGINT, syscall.SIGKILL}

	pass, hash, err := genPassAndHash()
	if err != nil {
		return nil, err
	}
	log.Println("your pass is", pass)

	// TODO;the user gets to input their name here in the future
	masterClient := client.CreateClient(owner, "")
	masterID := uuid.NewString()
	clients[masterID] = masterClient

	// listen on tcp port 42024 on all available p addresses of the local system.
	// godocs recommeds not assigning a host https://pkg.go.dev/net#Listen
	listener, err := net.Listen("tcp", ":42024")
	if err != nil {
		return nil, err
	}

	p, err := pipeline.NewPipeline(maxConns)
	if err != nil {
		return nil, fmt.Errorf("couldn't start a new pipeline")
	}
	p.Subscribe(&masterClient.Conn)

	return &Server{
		OwnerName:   masterClient.Name,
		listener:    listener,
		clients:     clients,
		signals:     signals,
		masterID:    masterID,
		timeStarted: time.Now(),
		hashTimeout: time.Minute * 5,
		currentHash: hash,
		maxClients:  maxConns + 1, // plus one to account for master client
		pipeline:    p,
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

func (s *Server) Run() {
	// cancel to propagate shutdown signal to all child contexts
	s.sigChan = make(chan os.Signal, 1)
	signal.Notify(s.sigChan, s.signals...)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		signal.Stop(s.sigChan)
		cancel()
		s.ShutdownServer()
	}()

	go s.newPass()

	go func() {
		for {
			log.Println("waiting for someone to join...")
			conn, err := s.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}

				log.Println("err: from client ", err)
				continue
			}

			if len(s.clients) >= s.maxClients {
				u.TryWrite(ctx,
					&u.WriteStruct{
						Conn:     conn,
						MaxTries: maxTries,
						HdrVal:   u.HdrErrVal,
						HdrMsg:   u.ErrServerFull,
						Message:  []byte("server full"),
					})

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
						// FIXME; THIS IS JUST TEMPORARY CODE!
						s.sigChan <- syscall.SIGINT
						close(errChan)
					}
					return
				case <-ctx.Done():
					if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						conn.Close()
					}
				}
			}(conn)
		}
	}()
	<-s.sigChan
	cancel()
}

// regenerate the hash and passphrase every hashTimeout minutes
func (s *Server) newPass() {
	for range time.Tick(s.hashTimeout) {
		s.pass, s.currentHash, _ = genPassAndHash()
		if !s.passRotated {
			s.passRotated = true
		}

		// TODO; remove this line
		fmt.Println("your new pass:", s.pass)
	}
}

func (s *Server) ShutdownServer() {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	// defer cancel()

	// add code to close all other clients
	log.Println("shutting down server now")
	// this gives the client a chance to detach itself and close the connection
	// to the server but if the client is unresponsive the server has no choice
	// than to close the connection by force
	for k, v := range s.clients {
		if k != s.masterID {
			s.pipeline.Unsubscribe(&v.Conn)
			// err := u.TryWrite(ctx, &u.WriteStruct{
			err := u.TryWrite(context.TODO(), &u.WriteStruct{
				Conn:     v.Conn,
				HdrVal:   u.HdrInfoVal,
				HdrMsg:   u.InfoShutdown,
				MaxTries: maxTries,
			})

			// force close the connection as cleanup as we assume the
			// client has already disconnected prematurely
			if err != nil {
				v.Conn.Close()
			}

			delete(s.clients, k)
		}
	}

	// select {
	// case <-ctx.Done():
	// 	for k, v := range s.clients {
	// 		if k != s.masterID {
	// 			v.Conn.Close()
	// 		}
	// 		delete(s.clients, k)
	// 	}
	// }
	s.pipeline.Close()
	s.listener.Close()
}

// authenticate and prepare client for joining the pipeline
func (s *Server) handleNewClient(ctx context.Context, conn net.Conn, errChan chan<- error) {
	s.mu.Lock()
	var client client.Client
	var clientData bytes.Buffer
	handleCtx, cancel := context.WithTimeout(ctx, maxTimeout)

	defer func() {
		defer s.mu.Unlock()
		defer cancel()
	}()

	go func() {
		select {
		case <-handleCtx.Done():
			// only send an err across the channel when the deadline is exceeded
			if errors.Is(handleCtx.Err(), context.DeadlineExceeded) {
				errChan <- u.ErrCtxTimeOut
				return
			}
		}
	}()

	// send the server identification message to the client
	if err := s.sendServerAuth(handleCtx, conn); err != nil {
		errChan <- err
		return
	}

	// read the information from the client
	data, err := u.TryRead(handleCtx, conn, maxTries)
	if err != nil {
		errChan <- err
		return
	}

	// NOTE; this is a very important func for reacting to headers
	_, _, msg, err := u.ParseIncomingMsg(data)
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

	// maybe pass a message to the client about the wrong pass
	if err := s.tryValidatePass(handleCtx, conn, client.PassUsed); err != nil {
		u.TryWrite(context.TODO(), &u.WriteStruct{
			Conn:     conn,
			MaxTries: 3,
			HdrVal:   u.HdrErrVal,
			HdrMsg:   u.ErrServerKick,
			Message:  []byte("wrong_pass"),
		})
		errChan <- fmt.Errorf("client failed authentication")
		return
	}
	// add client to list of clients now
	client.Conn = conn
	s.clients[uuid.NewString()] = &client
	// s.pipeline.Subscribe(&conn)

	fmt.Println("welcome", client.Name)

	return
}

// TODO; current ctx runs for about 5s and we can always increment that duration BUT
// what if the client doesn't enter a pass as fast as possible and the server kicks
// them for timeout;
// POSSIBLE SOLUTIONS
// * once the conn to the server is dead, upon entering a passphrase (any) try reconning
// to the server and submit the same details for verification; this time around the counter
// has reset to 0 (THAT COULD BE A MASSIVE EXPLOIT!)
//   - fix for exploit(potential); recognize the ip of the client that keeps this behaviour
//     up and apply a cooldown of about 5 minutes to reduce spamming the server
//
// * don't run the ctx until user verifies code but that leaves the server's port waiting
// which means we need a maximum server open window which the ctx already solves
func (s *Server) tryValidatePass(ctx context.Context, conn net.Conn, clientPass string) error {
	for tries := 0; tries < passTries; tries++ {
		if u.CheckPassphrase(s.currentHash, clientPass) {
			return nil
		}

		if tries == passTries-1 {
			return u.ErrWrongPass
		}

		err := u.TryWrite(ctx,
			&u.WriteStruct{
				Conn:     conn,
				MaxTries: maxTries,
				HdrVal:   u.HdrErrVal,
				HdrMsg:   u.ErrWrongPassphrase,
				Message:  []byte("wrong_pass"),
			})
		if err != nil {
			return err
		}

		p, err := u.TryRead(ctx, conn, maxTries)
		if err != nil {
			return err
		}

		// parse and extract the relevant message from the client's message
		_, _, msg, err := u.ParseIncomingMsg(p)
		if err != nil {
			return err
		}

		clientPass = string(msg)
	}

	return nil
}

func (s *Server) sendServerAuth(ctx context.Context, conn net.Conn) error {
	for tries := 0; tries < maxTries; tries++ {
		select {
		case <-ctx.Done():
			return u.ErrCtxTimeOut
		default:
			err := u.TryWrite(ctx,
				&u.WriteStruct{
					Conn:     conn,
					MaxTries: maxTries,
					HdrVal:   u.HdrAckVal,
					HdrMsg:   u.AckSelfReport,
					Message:  []byte(""),
				})

			if err != nil {
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
