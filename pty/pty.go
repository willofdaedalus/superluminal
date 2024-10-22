package pty

import (
	"fmt"
	"os"
)

type Tether struct {
	pty *os.File
}

// creates a new Tether to bridge the pty and the rest of the world
func NewTether() (*Tether, error) {
	var err error

	t := &Tether{}
	t.pty, err = createSession()
	if err != nil {
		return nil, err
	}

	return t, nil
}

// writes whatever is passed to the PTY
func (t *Tether) WriteTo(stuff []byte) {
	if len(stuff) == 0 {
		return
	}

	t.pty.Write(stuff)
}

// reads whatever is in the pty and returns it
func (t *Tether) ReadFrom() []byte {
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
