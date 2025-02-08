package pipeline

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
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
	go func() {
		// clear the screen
		// p.WriteTo([]byte("\x0C"))
		for {
			select {
			case <-p.stopChan:
				return
			default:
				// read from pty
				buf := p.ReadFrom()
				if buf == nil {
					done <- struct{}{}
					return
				}

				// this is for the client facing side so that they "see" what's happening
				p.writeDataToScreen(buf)

				// broadcast to all consumers
				p.mu.Lock()

				bufCopy := make([]byte, len(buf))
				copy(bufCopy, buf)
				termPayload := base.GenerateTermContent(uuid.NewString(), uint32(len(buf)), buf)
				payload, err := base.EncodePayload(common.Header_HEADER_TERMINAL_DATA, &termPayload)
				if err != nil {
					log.Println("failed to encode the terminal payload in pipeline.Start")
					log.Println(err)
					continue
					// not quite sure what do with the error yet
				}

				for conn := range p.consumers {
					// non-blocking write to prevent slow consumers from blocking
					select {
					case <-p.stopChan:
						p.mu.Unlock()
						return
					default:
						writeErr := utils.WriteFull(context.TODO(), conn, nil, payload)
						if writeErr != nil {
							fmt.Printf("Error writing to consumer: %v\n", writeErr)
							// consider removing the failing consumer
							delete(p.consumers, conn)
						}
					}
				}
				p.mu.Unlock()
			}
		}
	}()
}

func (p *Pipeline) writeDataToScreen(data []byte) {
	fmt.Printf("%s", string(data))
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
	buf := make([]byte, 10240)
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
