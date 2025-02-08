package backend

import (
	"bytes"
	"net"
	"os"
	"sync"
	"testing"
	"time"
	"willofdaedalus/superluminal/internal/payload/base"
	"willofdaedalus/superluminal/internal/payload/common"
	"willofdaedalus/superluminal/internal/pipeline"
	"willofdaedalus/superluminal/internal/utils"
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

func TestSession_Start(t *testing.T) {
	type fields struct {
		Owner         string
		maxConns      uint8
		pass          string
		hash          string
		clients       map[string]*sessionClient
		pipeline      *pipeline.Pipeline
		listener      net.Listener
		signals       []os.Signal
		mu            sync.Mutex
		tracker       *utils.SyncTracker
		passRegenTime time.Duration
		heartbeatTime time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Owner:         tt.fields.Owner,
				maxConns:      tt.fields.maxConns,
				pass:          tt.fields.pass,
				hash:          tt.fields.hash,
				clients:       tt.fields.clients,
				pipeline:      tt.fields.pipeline,
				listener:      tt.fields.listener,
				signals:       tt.fields.signals,
				mu:            tt.fields.mu,
				tracker:       tt.fields.tracker,
				passRegenTime: tt.fields.passRegenTime,
				heartbeatTime: tt.fields.heartbeatTime,
			}
			if err := s.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Session.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_End(t *testing.T) {
	type fields struct {
		Owner         string
		maxConns      uint8
		pass          string
		hash          string
		clients       map[string]*sessionClient
		pipeline      *pipeline.Pipeline
		listener      net.Listener
		signals       []os.Signal
		mu            sync.Mutex
		tracker       *utils.SyncTracker
		passRegenTime time.Duration
		heartbeatTime time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Owner:         tt.fields.Owner,
				maxConns:      tt.fields.maxConns,
				pass:          tt.fields.pass,
				hash:          tt.fields.hash,
				clients:       tt.fields.clients,
				pipeline:      tt.fields.pipeline,
				listener:      tt.fields.listener,
				signals:       tt.fields.signals,
				mu:            tt.fields.mu,
				tracker:       tt.fields.tracker,
				passRegenTime: tt.fields.passRegenTime,
				heartbeatTime: tt.fields.heartbeatTime,
			}
			if err := s.End(); (err != nil) != tt.wantErr {
				t.Errorf("Session.End() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
