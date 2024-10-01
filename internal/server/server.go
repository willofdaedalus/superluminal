package server

import (
	"fmt"
	"net"
	"time"
	c "willofdaedalus/superluminal/internal/client"
)

type Server struct {
	Clients []c.Client
	Owner   c.Client

	port          string
	addr        string
	currentHash   string
	hashTimeOut   time.Time
	serverStarted time.Time
	listener      net.Listener
}

func CreateServer() (*Server, error) {
	masterClient := c.CreateClient("manny", true)

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("couldn't start server for superluminal")
	}

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("couldn't start server for superluminal")
	}

	addr, err := getIpAddr()
	if err != nil {
		return nil, err
	}

	return &Server{
		addr:        addr,
		port:          port,
		Owner:         *masterClient,                   //client that started the server
		Clients:       []c.Client{*masterClient},       // append master to the list of clients
		serverStarted: time.Now(),                      // timestampo for start of server
		hashTimeOut:   time.Now().Add(time.Minute * 5), // hash times out after 5mins
		listener:      listener,
	}, nil
}

func (s *Server) AcceptNewClient(client *c.Client) {
	s.Clients = append(s.Clients, *client)
}
