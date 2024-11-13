package backend

import (
	"bytes"
	"net"
	"testing"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
)

const (
	adminName = "admin"
)

func TestNewSession(t *testing.T) {
	session, err := NewSession(adminName, 1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer session.End()

	if session.Owner != adminName {
		t.Fatalf("expected %s got %s", adminName, session.Owner)
	}
}

func TestSessionFullMessage(t *testing.T) {
	payloadBytes := make([]byte, 256)

	testSession, err := NewSession(adminName, 1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	testSession.maxConns = 0
	ready := make(chan struct{})

	go func() {
		go testSession.Start()
		close(ready)
	}()
	<-ready

	conn, err := net.Dial("tcp", "localhost:42024")
	if err != nil {
		t.Fatalf("%v", err)
	}

	n, err := conn.Read(payloadBytes)
	if err != nil {
		t.Fatalf("%v", err)
	}

	payload, err := base.DecodePayload(payloadBytes[:n])
	if err != nil {
		t.Fatal("couldn't decoded the payload\n")
	}
	t.Log(n)

	if payload.Header != common.Header_HEADER_ERROR {
		t.Fatalf("expected error header got %s", payload.GetHeader().String())
		t.Fail()
	}

	likelyErr, ok := payload.GetContent().(*base.Payload_Error)
	if !ok {
		t.Fatalf("unexpected payload %v", payload.GetContent())
	}

	if !bytes.Equal(likelyErr.Error.GetMessage(), []byte("server_full")) {
		t.Fatal("received wrong message from the session")
	}

	conn.Close()
}
