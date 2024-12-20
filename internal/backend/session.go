package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/payload/info"
	"willofdaedalus/superluminal/internal/pipeline"
	"willofdaedalus/superluminal/internal/utils"

	err1 "willofdaedalus/superluminal/internal/payload/error"

	"golang.org/x/sync/errgroup"
)

const (
	debugPassCount = 1
	maxAuthChances = 3
	maxTries       = 3

	heartbeatTimeout      = time.Second * 30
	clientKickTimeout     = time.Second * 30
	maxHandleTime         = time.Minute * 1
	passRegenTimeout      = time.Minute * 5
	serverShutdownTimeout = time.Minute * 1
)

func NewSession(owner string, maxConns uint8) (*session, error) {
	var reader bytes.Reader
	clients := make(map[string]*sessionClient, maxConns)

	listener, err := net.Listen("tcp", ":42024") // Listen on all interfaces (IPv4)
	if err != nil {
		return nil, err
	}

	p, err := pipeline.NewPipeline(maxConns)
	if err != nil {
		return nil, err
	}

	pass, hash, err := genPassAndHash(debugPassCount)
	if err != nil {
		return nil, err
	}
	log.Println("your pass is", pass)

	master := createClient(owner, nil, true)
	clients[master.uuid] = master

	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}

	return &session{
		Owner:         owner,
		maxConns:      maxConns + 1,
		clients:       clients,
		listener:      listener,
		pass:          pass,
		hash:          hash,
		pipeline:      p,
		reader:        reader,
		heartbeatTime: heartbeatTimeout,
		passRegenTime: passRegenTimeout,
		signals:       signals,
		tracker:       utils.NewSyncTracker(),
	}, nil
}

func (s *session) Start() error {
	// our shutdown can come from a signal for instance
	ctx, cancel := signal.NotifyContext(context.Background(), s.signals...)
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	defer func() {
		s.pipeline.Close()
		cancel()
		// defer close(doneChan)
		close(errChan)
	}()

	go s.pipeline.ReadStdin()
	go s.pipeline.Start()
	go s.regenPassLoop(ctx)
	go s.listen(ctx, doneChan, errChan)

	// wait for and handle errors
	select {
	case err := <-errChan:
		log.Printf("server err: %v", err)
		if err != nil {
			fmt.Println("exiting by err chan...")
			return err
		}
	case <-doneChan:
		fmt.Println("exiting by done chan...")
		return nil
	case <-ctx.Done():
		fmt.Println("exiting by context done...")
		return s.End()
	}

	return nil
}

func (s *session) listen(ctx context.Context, doneChan chan<- struct{}, errChan chan error) {
	idChan := make(chan string, 1)

	log.Println("server started...")
	defer close(doneChan)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			errChan <- fmt.Errorf("accept error: %w", err)
			return
		}

		// upon successful connection
		go func(conn net.Conn) {
			fmt.Println("new connection...")
			if len(s.clients) >= int(s.maxConns) {
				tempCtx, tempCancel := context.WithTimeout(ctx, clientKickTimeout)
				defer tempCancel()

				errorMsg := base.GenerateError(
					err1.ErrorMessage_ERROR_SERVER_FULL,
					[]byte("server_full"),
					[]byte("server is full"),
				)

				errPayload, err := base.EncodePayload(common.Header_HEADER_ERROR, errorMsg)
				if err != nil {
					conn.Close()
					return
				}

				// need a minute to write to the client; if it's not possible don't bother
				err = s.writeToClient(tempCtx, conn, errPayload)
				if err != nil {
					if errors.Is(err, io.EOF) {
						log.Println("sprlmnl: client is closed")
					} else {
						log.Printf("write error: %v", err)
					}
				}

				conn.Close()
				log.Println("rejected client with server_full error")
				return
			}

			clientID, err := s.handleNewConn(ctx, conn)
			if err != nil {
				errChan <- err
				idChan <- ""
				return
			}

			idChan <- clientID
		}(conn)

		// probably not the best way to go at this but a check in handleClientIO
		// will return early if idChan passes an empty string due to an error
		go s.handleClientIO(ctx, <-idChan)
	}
}

func (s *session) kickAndCloseClient(
	ctx context.Context, conn net.Conn, errType err1.ErrorMessage_ErrorCode, details []string) {
	kickCtx, cancel := context.WithTimeout(ctx, clientKickTimeout)
	defer func() {
		cancel()
		conn.Close()
	}()

	errPayload := base.GenerateError(errType, []byte(details[0]), []byte(details[1]))
	payload, err := base.EncodePayload(common.Header_HEADER_ERROR, errPayload)
	if err != nil {
		return
	}

	err = s.writeToClient(kickCtx, conn, payload)
	if err != nil {
		return
	}
}

// End sends a message to all connected clients about the shutdown and then
// proceeds to shutdown.
// with the current implementation, End doesn't wait for a departing message
// from each client before shutting down
func (s *session) End() error {
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()
	log.Println("server is shutting down...")

	// prevent new connections
	s.listener.Close()

	// wait for active operations to complete
	for i := 0; i < maxTries; i++ {
		if !s.tracker.AnyActionInProgress() {
			break
		}
		select {
		case <-time.After(time.Second * 5):
			continue
		case <-ctx.Done():
			log.Println("Shutdown timeout reached before active operations completed")
			break
		}
	}

	// notify and close clients
	if len(s.clients) > 1 {
		group, gCtx := errgroup.WithContext(ctx)
		group.SetLimit(len(s.clients))

		s.mu.Lock()
		defer s.mu.Unlock()

		for _, client := range s.clients {
			if client.isOwner {
				continue
			}

			// client := client // capture loop variable
			group.Go(func() error {
				// Prepare shutdown notification
				infoPayload := base.GenerateInfo(info.Info_INFO_SHUTDOWN, "Server is shutting down")
				payload, err := base.EncodePayload(common.Header_HEADER_INFO, infoPayload)
				if err != nil {
					return err
				}

				// attempt to send shutdown message
				err = s.writeToClient(gCtx, client.conn, payload)
				if err != nil &&
					!errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
					log.Printf("Error sending shutdown message to client: %v", err)
				}

				// ensure connection is closed
				s.pipeline.Unsubscribe(client.conn)
				client.conn.Close()
				delete(s.clients, client.uuid)
				return nil
			})
		}

		// wait for all client shutdown attempts
		return group.Wait()
	}

	return nil
}

func (s *session) handleNewConn(ctx context.Context, conn net.Conn) (string, error) {
	name, err := s.authenticateClient(ctx, conn)
	if err != nil {
		s.kickAndCloseClient(ctx,
			conn,
			err1.ErrorMessage_ERROR_AUTH_FAILED, []string{"failed_auth", "couldn't pass auth"},
		)
		return "", err
	}

	s.mu.Lock()
	newClient := createClient(name, conn, false)
	s.clients[newClient.uuid] = newClient

	s.pipeline.Subscribe(newClient.conn)
	s.mu.Unlock()

	// send a congratulatory message to the client
	infoPayload := base.GenerateInfo(info.Info_INFO_AUTH_SUCCESS, "welcome to the session")
	payload, err := base.EncodePayload(common.Header_HEADER_INFO, infoPayload)
	if err != nil {
		return "", fmt.Errorf("failed to encode welcome payload: %w", err)
	}

	err = s.writeToClient(ctx, conn, payload)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("client disconnected: %w", err)
		}
		return "", fmt.Errorf("failed to send welcome message: %w", err)
	}

	log.Println("hello client", newClient.uuid)
	return newClient.uuid, nil
}

func (s *session) kickClientGracefully(clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	curClient, ok := s.clients[clientID]
	if !ok {
		return fmt.Errorf("failed to find client in kickClientGracefully")
	}

	infoPayload := base.GenerateInfo(info.Info_INFO_REQ_ACK, "goodbye...")
	payload, err := base.EncodePayload(common.Header_HEADER_INFO, infoPayload)
	if err == nil {
		writeCtx, cancel := context.WithTimeout(context.Background(), clientKickTimeout)
		s.writeToClient(writeCtx, curClient.conn, payload)
		cancel()
	}

	s.pipeline.Unsubscribe(curClient.conn)
	delete(s.clients, clientID)
	return nil
}

func (s *session) handleClientIO(ctx context.Context, clientID string) {
	client, ok := s.clients[clientID]
	if !ok {
		return
	}

	key := clientUniqID("client_id")
	procCtx := context.WithValue(ctx, key, clientID)

	readErrChan := make(chan error, 1)
	readDataChan := make(chan []byte, 1)
	errChan := make(chan error, 1) // Updated to be properly synchronized
	var wg sync.WaitGroup

	defer func() {
		wg.Wait()      // Wait for all goroutines to finish
		close(errChan) // Only close after all usage is done
		close(readErrChan)
		close(readDataChan)
	}()

	// Reading goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				read, err := s.readFromClient(ctx, client.conn)
				if err != nil {
					readErrChan <- err
					return
				}
				readDataChan <- read
			}
		}
	}()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-readErrChan:
			if errors.Is(err, utils.ErrConnectionClosed) {
				log.Println("client has already shut down")
			} else if errors.Is(err, utils.ErrFailedAfterRetries) {
				log.Println("failed to read from the server")
			}
			return // Stop processing for this client
		case read := <-readDataChan:
			wg.Add(1)
			go func(data []byte) {
				defer wg.Done()
				s.processPayload(procCtx, data, errChan)
			}(read)
		}
	}
}

func (s *session) processPayload(ctx context.Context, data []byte, errChan chan<- error) {
	payload, err := base.DecodePayload(data)
	if err != nil {
		errChan <- err
		return
	}

	switch payload.GetHeader() {
	case common.Header_HEADER_HEARTBEAT:
		errChan <- s.handleHeartbeatResp()
	case common.Header_HEADER_INFO:
		infoPayload, ok := payload.GetContent().(*base.Payload_Info)
		if !ok {
			errChan <- fmt.Errorf("couldn't assert info payload")
		}
		errChan <- s.handleClientInfoMsg(ctx, infoPayload)
	}

}

func (s *session) handleClientInfoMsg(ctx context.Context, infoPayload *base.Payload_Info) error {
	switch infoPayload.Info.InfoType {
	case info.Info_INFO_SHUTDOWN:
		id, ok := ctx.Value(clientUniqID("client_id")).(string)
		if !ok {
			return fmt.Errorf("unable to get value")
		}
		log.Println("closing client", id)
		return s.kickClientGracefully(id)
	}

	return nil
}

func (*session) handleHeartbeatResp() error {

	return nil
}
