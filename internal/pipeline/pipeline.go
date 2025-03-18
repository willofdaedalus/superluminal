package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/utils"

	"github.com/google/uuid"
)

type Pipeline struct {
	pty           *os.File
	logFile       *os.File
	mainClient    *os.File
	consumers     map[net.Conn]struct{}
	consumerCount uint8
	lastMsg       int
	mu            sync.Mutex
	stopChan      chan struct{}
}

// creates a new pipeline to bridge the pty and the rest of the world
func NewPipeline(maxConns uint8) (*Pipeline, error) {
	pty, err := createSession()
	if err != nil {
		return nil, err
	}
	file, _ := os.OpenFile("./log.output", os.O_CREATE|os.O_WRONLY, 0644)
	return &Pipeline{
		pty:       pty,
		consumers: make(map[net.Conn]struct{}, maxConns),
		stopChan:  make(chan struct{}),
		logFile:   file,
	}, nil
}

// starts broadcasting pty output to all connected consumers
func (p *Pipeline) Start(done chan<- struct{}) {
	// create and grow a buffer of 10MB
	receiveBuffer := bytes.NewBuffer(nil)
	receiveBuffer.Grow(receiveBufferInitSize)

	go func() {
		// clear the screen
		// p.WriteTo([]byte("\x0C"))
		for {
			select {
			case <-p.stopChan:
				return
			default:
				// read from pty
				receiveBuffer.Write(p.ReadFrom())
				// fmt.Println("got from pty", receiveBuffer.Len())
				// read exactly `networkSendSize` bytes or whatever is available
				buf := receiveBuffer.Next(min(networkSendSize, receiveBuffer.Len()))
				// fmt.Println("got from buf", len(buf))

				if len(buf) == 0 {
					// either the pty got the "exit" command and quit which in that case returned
					// nothing or the pty crashed (highly unlikely) but either case we
					// signal the done chan and exit from the function effectively ending the
					// session. session handler makes sure everyone is closed
					done <- struct{}{}
					return
				}

				// buf := parser.EncodeData(fromPty)

				// this is for the client facing side so that they "see" what's happening
				p.writeDataToScreen(buf)

				termPayload := base.GenerateTermContent(uuid.NewString(), uint32(len(buf)), buf)
				payload, err := base.EncodePayload(common.Header_HEADER_TERMINAL_DATA, &termPayload)
				if err != nil {
					log.Println("failed to encode the terminal payload in pipeline.Start")
					log.Println(err)
					continue
					// not quite sure what do with the error yet
				}

				// broadcast to all consumers
				p.mu.Lock()
				for conn := range p.consumers {
					// non-blocking write to prevent slow consumers from blocking
					select {
					case <-p.stopChan:
						// p.mu.Unlock()
						return
					default:
						writeCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
						writeErr := utils.WriteFull(writeCtx, conn, nil, payload)
						if writeErr != nil {
							fmt.Printf("Error writing to consumer: %v\n", writeErr)
							// consider removing the failing consumer
							delete(p.consumers, conn)
						}
						cancel()
					}
				}
				p.mu.Unlock()
			}
		}
	}()
}

func (p *Pipeline) writeDataToScreen(data []byte) {
	fmt.Printf("%v", string(data))
	p.logFile.Write(data)
}

// Add a new client to the pipeline
func (p *Pipeline) Subscribe(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consumers[conn] = struct{}{}
	p.consumerCount += 1
}

// Remove a client from the pipeline
func (p *Pipeline) Unsubscribe(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.consumers, conn)
	if p.consumerCount > 0 {
		p.consumerCount -= 1
	}
}

// Closes the pipeline and stops broadcasting
func (p *Pipeline) Close() {
	p.mu.Lock()
	// signal the broadcasting goroutine to stop
	close(p.stopChan)
	p.logFile.Close()

	for k := range p.consumers {
		delete(p.consumers, k)
		k.Close()
	}
	p.mu.Unlock()

	p.pty.Close()
}

// Writes whatever is passed to the PTY
// Used to pass commands to the PTY
func (p *Pipeline) WriteTo(stuff []byte) {
	if len(stuff) == 0 || stuff == nil {
		return
	}
	p.pty.Write(stuff)
}

// reads whatever is in the pty and returns it
func (p *Pipeline) ReadFrom() []byte {
	buf := make([]byte, networkSendSize)
	n, err := p.pty.Read(buf)
	if err != nil {
		return nil
	}

	return buf[:n]
}

func (p *Pipeline) ReadStdin() {
	// Buffer for reading a single byte
	buf := make([]byte, 1)
	for {
		// Read a single byte from stdin
		n, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Println("Error reading standard input:", err)
			return
		}

		// Ensure the byte read is sent to the PTY
		if n > 0 {
			p.WriteTo(buf)
		}
	}
}

// func (p *Pipeline) ReadStdin() {
// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		// Get the line of input
// 		line := scanner.Text()

// 		// Append a newline, as PTY expects terminal-like input
// 		lineWithNewline := line + "\n"

// 		// Write to the PTY
// 		p.WriteTo([]byte(lineWithNewline))
// 	}

// 	// Check for errors during scanning
// 	if err := scanner.Err(); err != nil {
// 		fmt.Println("Error reading standard input:", err)
// 	}
// }
