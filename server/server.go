package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"time"
	c "willofdaedalus/superluminal/client"
)

type Server struct {
	Clients       []c.Client
	Owner         c.Client
	Buffer        bytes.Buffer
	port          string
	addr          string
	currentHash   string
	hashTimeOut   time.Time
	serverStarted time.Time
	listener      net.Listener
	sig           chan os.Signal
	signals       []os.Signal
}

func CreateServer() (*Server, error) {
	masterClient := c.CreateClient("manny", true, nil)

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

	return &Server{
		addr:          addr,
		port:          port,
		Owner:         *masterClient,                   //client that started the server
		Clients:       []c.Client{*masterClient},       // append master to the list of clients
		serverStarted: time.Now(),                      // timestampo for start of server
		hashTimeOut:   time.Now().Add(time.Minute * 5), // hash times out after 5mins
		listener:      listener,
		signals:       sigs,
	}, nil
}

func (s *Server) StartServer() {
	s.handleSignals()
	fmt.Printf("server started at %q\n", fmt.Sprintf("%s:%s", s.addr, s.port))

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
				fmt.Println("server shutting down")
				return
			}

			log.Println(err)
			continue
		}

		go handleNewClient(conn)
	}
}

func (s *Server) ShutdownServer() {
	fmt.Println("server shutting down...")
	s.listener.Close()
}

func (s *Server) AcceptNewClient(client *c.Client) {
	s.Clients = append(s.Clients, *client)
}
