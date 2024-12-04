package pipeline

import (
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
	cs := make(map[net.Conn]struct{}, maxConns)
	return &Pipeline{
		pty:       pty,
		consumers: cs,
		stopChan:  make(chan struct{}),
	}, nil
}

// starts broadcasting pty output to all connected consumers
func (p *Pipeline) Start() {
	go func() {
		p.WriteTo([]byte("\x0C"))
		for {
			select {
			case <-p.stopChan:
				return
			default:
				// read from pty
				buf, err := p.ReadFrom()
				if err != nil || buf == nil {
					log.Printf("there was an error %v or the buf was nil", err)
					continue
				}

				// this is for the client facing side so that they "see" what's happening
				writeDataToScreen(buf)

				// broadcast to all consumers
				p.mu.Lock()
				for conn := range p.consumers {
					// non-blocking write to prevent slow consumers from blocking
					select {
					case <-p.stopChan:
						p.mu.Unlock()
						return
					default:
						termPayload := base.GenerateTermContent(uuid.NewString(), buf)
						payload, err := base.EncodePayload(common.Header_HEADER_TERMINAL_DATA, &termPayload)
						if err != nil {
							log.Println("failed to encode the terminal payload in pipeline.Start")
							log.Println(err)
							continue
						}

						_, writeErr := conn.Write(payload)
						if writeErr != nil {
							fmt.Printf("Error writing to consumer: %v\n", writeErr)
							// consider removing the failing consumer
							delete(p.consumers, conn)
						}
						home, _ := os.UserHomeDir()
						utils.LogBytes("wrote", home+"/superluminal.log", payload)
					}
				}
				p.mu.Unlock()
			}
		}
	}()
}

func writeDataToScreen(data []byte) {
	fmt.Printf(string(data))
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
	// Signal the broadcasting goroutine to stop
	close(p.stopChan)

	// Clear out the consumers
	p.mu.Lock()
	for k := range p.consumers {
		delete(p.consumers, k)
		k.Close()
	}
	p.mu.Unlock()

	// Close the PTY
	p.pty.Close()
}

// Writes whatever is passed to the PTY
// Used to pass commands to the PTY
func (p *Pipeline) WriteTo(stuff []byte) {
	if len(stuff) == 0 {
		return
	}
	p.pty.Write(stuff)
}

// reads whatever is in the pty and returns it
func (p *Pipeline) ReadFrom() ([]byte, error) {
	buf := make([]byte, 10240)
	n, err := p.pty.Read(buf)
	if err != nil {
		fmt.Println("Couldn't read from the PTY:", err)
		return nil, err
	}

	return buf[:n], nil
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
