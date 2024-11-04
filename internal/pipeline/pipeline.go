package pipeline

import (
	"fmt"
	"net"
	"os"
	"sync"
)

type consumer struct {
	conn      net.Conn
	lastMsgId int
}

type Pipeline struct {
	pty           *os.File
	mainClient    *os.File
	consumers     map[*net.Conn]struct{}
	consumerCount int
	lastMsg       int
	mu            sync.Mutex
}

// creates a new Pipeline to bridge the pty and the rest of the world
func NewPipeline(maxConns int) (*Pipeline, error) {
	tty, err := createSession()
	if err != nil {
		return nil, err
	}

	cs := make(map[*net.Conn]struct{}, maxConns)

	return &Pipeline{
		pty:       tty,
		consumers: cs,
	}, nil
}

func (p *Pipeline) Start() {

}

// add a new client to the pipeline
func (p *Pipeline) Subscribe(conn *net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// c := &consumer{
	// 	conn:      *conn,
	// 	lastMsgId: p.lastMsg,
	// }

	p.consumers[conn] = struct{}{}
	p.consumerCount += 1

}

// remove a client from the pipeline
func (p *Pipeline) Unsubscribe(conn *net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// not decrementing the consumer count for sanity check
	for c := range p.consumers {
		if c == conn {
			delete(p.consumers, c)
		}
	}
}

func (p *Pipeline) Close() {
	// clear out the consumers
	for k := range p.consumers {
		delete(p.consumers, k)
	}

	p.pty.Close()
}

// writes whatever is passed to the PTY
// used to pass commands to the PTY
func (t *Pipeline) WriteTo(stuff []byte) {
	if len(stuff) == 0 {
		return
	}

	t.pty.Write(stuff)
}

// reads whatever is in the pty and returns it
func (t *Pipeline) ReadFrom() []byte {
	buf := make([]byte, 1024)
	for {
		n, err := t.pty.Read(buf)
		if err != nil {
			fmt.Println("couldn't read from the pty:", err)
			continue
		}

		return buf[:n]
	}
}
