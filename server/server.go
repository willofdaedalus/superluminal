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
	header             = "v1.sprlmnl."
	retryAttempts      = 4
	passphraseAttempts = 3
)

type Server struct {
	Clients       []c.Client
	Owner         c.Client
	Buffer        bytes.Buffer
	numConns      int
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

	return &Server{
		addr:          addr,
		port:          port,
		Owner:         *masterClient,                   //client that started the server
		Clients:       []c.Client{*masterClient},       // append master to the list of clients
		serverStarted: time.Now(),                      // timestampo for start of server
		hashTimeOut:   time.Now().Add(time.Minute * 5), // hash times out after 5mins
		maxConns:      maxConns + 1,                    // plus one to account for master client
		listener:      listener,
		signals:       sigs,
		currentHash:   hash,
	}, nil
}

func (s *Server) StartServer() {
	var fromClient bytes.Buffer
	var incomingClient c.TransportClient

	ctx := context.Background()
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

		// check that the server doesn't take more than it can handle
		if len(s.Clients) >= s.maxConns {
			sendMessage(ctx, conn, config.PreErr, config.ServerFull)
			conn.Close()
			continue
		}

		var correctPass, headerSent bool
		func(ctx context.Context, conn net.Conn) {
			// every decoder needs to be new
			enc := gob.NewDecoder(&fromClient)
			handshakeCtx, cancel := context.WithTimeout(ctx, config.MaxServerHandshakeTime)
			defer cancel()

			// send header
			clientHeader := fmt.Sprintf("%s%d", header, time.Now().Unix())
			for i := 0; i < retryAttempts; i++ {
				if err := sendMessage(handshakeCtx, conn, config.PreHeader, clientHeader); err != nil {
					log.Println("Error:", err)
					log.Println("Retrying handshake attempt", i+1)
					continue
				}

				// Successfully sent message, break out of loop
				fmt.Println("sent header")
				headerSent = true
				break
			}

			if !headerSent {
				if err := conn.Close(); err != nil {
					log.Println("Error closing connection:", err)
				}

				return
			}

			// Read incoming data
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Println("client disconnected")
					return
				}
				log.Println("Error reading from connection:", err)
				return
			}

			fromClient.Write(buf[:n])

			// decode incoming client data
			err = enc.Decode(&incomingClient)
			if err != nil {
				log.Println("Error decoding client data:", err)
				return
			}

			// check the passphrase
			for i := 0; i < passphraseAttempts; i++ {
				if utils.CheckPassphrase(s.currentHash, incomingClient.PassUsed) {
					correctPass = true
					break
				}

				if i == passphraseAttempts-1 {
					return
				}

				// if passphrase is incorrect, retry sending an error
				for j := 0; j < retryAttempts; j++ {
					if err := sendMessage(handshakeCtx, conn, config.PreErr, config.RejectedPass); err != nil {
						log.Println("Error sending rejected pass message:", err)
						log.Println("Retrying message send", j+1)
						continue
					}
					break
				}

				// read another attempt from the client
				fmt.Println("reading pass")
				n, err := conn.Read(buf)
				if err != nil {
					if errors.Is(err, io.EOF) {
						log.Println("client disconnected")
						return
					}
					log.Println("Error reading from connection:", err)
					return
				}
				incomingClient.PassUsed = string(buf[:n])
			}
		}(ctx, conn)


		// if the password was incorrect after attempts close the connection if
		// we can't send the clientauthfailed message to the client
		if !correctPass {
			for j := 0; j < retryAttempts; j++ {
				if err := sendMessage(ctx, conn, config.PreErr, config.ClientAuthFailed); err != nil {
					log.Println("Error sending auth failure message:", err)
					log.Println("Retrying to notify client", j+1)
					continue
				}
				// successfully notified client
				break
			}

			// if message sending failed after retries, close the connection
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection:", err)
			}

			log.Println("sprlmnl: kicked client for wrong passphrase")
		} else {
			// Add the authenticated client to the list
			var newClient c.Client
			newClient.TransportClient = incomingClient
			newClient.Conn = conn
			s.Clients = append(s.Clients, newClient)
			incomingClient = c.TransportClient{}

			// handle the new client connection
			go handleNewClient(conn)
		}
	}
}

func (s *Server) ShutdownServer() {
	fmt.Println("server shutting down...")
	for _, client := range s.Clients {
		if !client.Master {
			client.Conn.Write([]byte("SHT:"))
		}
	}

	s.listener.Close()
}

func (s *Server) AcceptNewClient(client *c.Client) {
	s.Clients = append(s.Clients, *client)
}

func (s *Server) canAcceptConnections(conn net.Conn) bool {
	// +1 to offset the master client
	if len(s.Clients) >= s.maxConns+1 {
		conn.Write([]byte(config.ServerFull))
		conn.Close()
		return false
	}
	return true
}
