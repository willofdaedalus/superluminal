package dispatcher

import (
	"os"
	"testing"
)

func TestReadTo(t *testing.T) {
	var d Dispatcher
	var pts *os.File
	var buf []byte

	testMsg := "hello world"

	os.Stdin.Write([]byte(testMsg))
	d.ReadTo(os.Stdin, pts)
	pts.Read(buf)

	if string(buf) != testMsg {
		t.Fatalf("expected %s but got %s", testMsg, string(buf))
	}
}
