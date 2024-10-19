package dispatcher

import (
	"fmt"
	"os"
)

type Dispatcher struct {

}

func (d *Dispatcher) readFromStdin(outStream chan []byte) error {
	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Println("err:", err)
			continue
		}

		outStream<-buf[:n]
	}
}

func (d *Dispatcher) WriteTo(pts *os.File) {

}
