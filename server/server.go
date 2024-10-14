package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"
	c "willofdaedalus/superluminal/client"
	"willofdaedalus/superluminal/config"
	"willofdaedalus/superluminal/utils"
)

const (
	header             = "v1.sprlmnl."
	retryAttempts      = 4
	passphraseAttempts = 3
)

type Server struct {
	buffer        bytes.Buffer
	clients       map[string]*c.Client
	owner         string
	numConns      int
	maxConns      int
	port          string
	addr          string
	currentHash   string
	hashTimeOut   time.Time
	serverStarted time.Time
	listener      net.Listener
	signal        chan os.Signal
	signals       []os.Signal
	shutdownSent  bool
}

func CreateServer(name string, maxConns int) (*Server, error) {

	listener, err := net.Listen("tcp", "localhost:42024")
	if err != nil {
		return nil, fmt.Errorf("couldn't start server for superluminal")
	}

	initialClientID := make(map[string]*c.Client)
	masterClient := c.CreateClient(name, nil)
	masterID := uuid.New()

	sigs := []os.Signal{syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT}

	addr, port, err := getServerDetails(listener)
	if err != nil {
		return nil, err
	}

	pass, err := utils.GeneratePassphrase()
	if err != nil {
		return nil, err
	}

	fmt.Println("your pass is:", pass)

	hash, err := utils.HashPassphrase(pass)
	if err != nil {
		return nil, err
	}

	initialClientID[masterID.String()] = masterClient
	return &Server{
		clients:       initialClientID,
		owner:         masterID.String(),
		addr:          addr,
		port:          port,
		serverStarted: time.Now(),                      // timestampo for start of server
		hashTimeOut:   time.Now().Add(time.Minute * 5), // hash times out after 5mins
		maxConns:      maxConns + 1,                    // plus one to account for master client
		listener:      listener,
		signals:       sigs,
		currentHash:   hash,
	}, nil
}

func (s *Server) StartServer() {
	ctx := context.Background()
	s.handleSignals()
	fmt.Printf("server started at %q\n", fmt.Sprintf("%s:%s", s.addr, s.port))

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
				// this might cause issues in the future
				log.Println("non-temporary error, closing server:", err)
				return
			}
			log.Println("error accepting connection:", err)
			continue
		}

		// Check if the server can handle more connections
		if len(s.clients) >= s.maxConns {
			s.rejectClient(ctx, conn, config.ServerFull)
			continue
		}

		go s.handleNewConnection(ctx, conn)
	}
}

// Refactor for handling new connection
func (s *Server) handleNewConnection(ctx context.Context, conn net.Conn) {
	var fromClient bytes.Buffer
	var incomingClient c.TransportClient
	var correctPass, headerSent bool

	handshakeCtx, cancel := context.WithTimeout(ctx, config.MaxServerHandshakeTime)
	defer cancel()

	headerSent = s.sendHeader(handshakeCtx, conn)
	if !headerSent {
		conn.Close()
		return
	}

	// read client data
	if err := s.readClientData(conn, &fromClient, &incomingClient); err != nil {
		conn.Close()
		return
	}

	// Authenticate passphrase
	correctPass = s.authenticateClient(ctx, conn, &incomingClient)
	if !correctPass {
		s.kickClient(ctx, conn, config.ClientAuthFailed)
		return
	}

	// Add the authenticated client
	clientID := s.addClient(conn, &incomingClient)
	go s.handleNewClient(conn, clientID)
}

// Send a header to the client
func (s *Server) sendHeader(ctx context.Context, conn net.Conn) bool {
	clientHeader := fmt.Sprintf("%s%d", header, time.Now().Unix())

	for i := 0; i < retryAttempts; i++ {
		if err := sendMessage(ctx, conn, config.PreHeader, clientHeader); err != nil {
			log.Printf("error sending header (attempt %d): %v", i+1, err)
			continue
		}
		fmt.Println("header sent successfully")
		return true
	}

	return false
}

// Read and decode client data
func (s *Server) readClientData(conn net.Conn, fromClient *bytes.Buffer, incomingClient *c.TransportClient) error {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("client disconnected")
		} else {
			log.Println("error reading from connection:", err)
		}
		return err
	}

	fromClient.Write(buf[:n])
	enc := gob.NewDecoder(fromClient)
	if err := enc.Decode(incomingClient); err != nil {
		log.Println("error decoding client data:", err)
		return err
	}

	return nil
}

// Authenticate client passphrase
func (s *Server) authenticateClient(ctx context.Context, conn net.Conn, incomingClient *c.TransportClient) bool {
	for i := 0; i < passphraseAttempts; i++ {
		if utils.CheckPassphrase(s.currentHash, incomingClient.PassUsed) {
			return true
		}

		// if passphrase incorrect, notify and allow retry
		if i < passphraseAttempts-1 {
			if err := sendMessage(ctx, conn, config.PreErr, config.RejectedPass); err != nil {
				log.Println("error sending rejected pass message:", err)
			}
			s.readNewPassphrase(conn, incomingClient)
		}
	}

	return false
}

// Read new passphrase from the client
func (s *Server) readNewPassphrase(conn net.Conn, incomingClient *c.TransportClient) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("error reading new passphrase:", err)
		return
	}
	incomingClient.PassUsed = string(buf[:n])
}

// Kick client for incorrect passphrase
func (s *Server) kickClient(ctx context.Context, conn net.Conn, errMsg string) {
	for j := 0; j < retryAttempts; j++ {
		if err := sendMessage(ctx, conn, config.PreErr, errMsg); err != nil {
			log.Println("error sending auth failure message:", err)
			continue
		}
		break
	}
	conn.Close()
	log.Println("client kicked")
}

// Add authenticated client to the server's list
func (s *Server) addClient(conn net.Conn, incomingClient *c.TransportClient) string {
	newClient := c.Client{
		TransportClient: *incomingClient,
		Conn:            conn,
		TimeJoined:      time.Now().Unix(),
	}

	clientID := uuid.New().String()
	fmt.Println("new client id:", clientID)
	s.clients[clientID] = &newClient

	return clientID
}

// Reject client if server is full
func (s *Server) rejectClient(ctx context.Context, conn net.Conn, errMsg string) {
	utils.SendMessage(ctx, conn, config.PreErr, errMsg, config.ErrServerCtxTimeout)
	conn.Close()
	log.Println("client rejected: server full")
}

// fix the code below
func (s *Server) ShutdownServer() {
	ctx, cancel := context.WithTimeout(context.Background(), config.MaxConnectionTime)
	defer cancel()

	for k, v := range s.clients {
		if k != s.owner {
			sentMsg := false
			for j := 0; j < retryAttempts; j++ {
				err := utils.SendMessage(
					ctx,
					v.Conn,
					config.PreShutdown,
					"server is shutting down",
					config.ErrServerCtxTimeout,
				)
				if err != nil {
					log.Println("Error sending auth failure message:", err)
					log.Println("Retrying to notify client", j+1)
					continue
				}
				// successfully notified client
				sentMsg = true
				break
			}

			// couldn't send message after retries
			if !sentMsg {
				v.Conn.Close()
			}
		}
	}

	s.listener.Close()
}
