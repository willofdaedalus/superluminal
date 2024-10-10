package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
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
	header = "v1.sprlmnl"
)

type Server struct {
	Clients       []c.Client
	Owner         c.Client
	Buffer        bytes.Buffer
	maxConns      int
	port          string
	addr          string
	currentHash   string
	hashTimeOut   time.Time
	serverStarted time.Time
	listener      net.Listener
	sig           chan os.Signal
	signals       []os.Signal
	shutdownSent  bool
}

func CreateServer(name string, maxConns int) (*Server, error) {
	masterClient := c.CreateClient(name, true, nil)

	listener, err := net.Listen("tcp", "localhost:42024")
	if err != nil {
		return nil, fmt.Errorf("couldn't start server for superluminal")
	}

	sigs := []os.Signal{syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT}

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("couldn't start server for superluminal")
	}

	addr, err := getIpAddr()
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

	return &Server{
		addr:          addr,
		port:          port,
		Owner:         *masterClient,                   //client that started the server
		Clients:       []c.Client{*masterClient},       // append master to the list of clients
		serverStarted: time.Now(),                      // timestampo for start of server
		hashTimeOut:   time.Now().Add(time.Minute * 5), // hash times out after 5mins
		listener:      listener,
		signals:       sigs,
		currentHash:   hash,
		maxConns:      maxConns,
	}, nil
}

func (s *Server) StartServer() {
	var fromClient bytes.Buffer
	var incomingClient c.Client

	ctx := context.Background()
	enc := gob.NewDecoder(&fromClient)
	s.handleSignals()
	fmt.Printf("server started at %q\n", fmt.Sprintf("%s:%s", s.addr, s.port))

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
				return
			}
			log.Println("error is:", err)
		}

		// send the header as bytes to the client on connect to confirm it's the
		// correct port and server
		ok, err := sendHeader(ctx, conn, header)
		if !ok {
			log.Println(err)
			continue
		}

		buf := make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("client disconnected")
				continue
			}
			log.Println(err)
			continue
		}
		fromClient.Write(buf)

		err = enc.Decode(&incomingClient)
		if err != nil {
			log.Println("error decoding:", err)
		}

		if !utils.CheckPassphrase(s.currentHash, incomingClient.PassUsed) {
			conn.Write([]byte("server rejected your passphrase. check and re-enter\n"))
			log.Printf("rejected %q for wrong passphrase", incomingClient.Name)
			// in the future allow up to 3 tries before closing the client's
			// connection to allow that flexibility and great UX
			conn.Close()
			continue
		}
		newClient := incomingClient
		newClient.Conn = conn
		s.Clients = append(s.Clients, newClient)

		go handleNewClient(conn)
	}
}

func (s *Server) ShutdownServer() {
	fmt.Println("server shutting down...")
	for _, client := range s.Clients {
		if !client.Master {
			client.Conn.Write([]byte(config.ShutdownMsg))
		}
	}

	s.listener.Close()
}

func (s *Server) AcceptNewClient(client *c.Client) {
	s.Clients = append(s.Clients, *client)
}
