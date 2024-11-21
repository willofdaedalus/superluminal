package client

import (
	"context"
	"testing"
	"willofdaedalus/superluminal/internal/backend"
)

const (
	name  = "haha"
	owner = "admin"
)

func TestListenForMessage(t *testing.T) {
	session, err := backend.NewSession(owner, 0)
	if err != nil {
		t.Fatal("failed to start server")
	}
	go session.Start()

	errChan := make(chan error, 1)
	c := New(name)
	if err := c.ConnectToSession(context.Background(), "localhost", "42024"); err != nil {
		t.Fatal("failed to connect to a session")
	}

	go c.ListenForMessages(errChan)

	for {
		select {
		case retErr := <-errChan:
			if retErr == nil {
				t.Fatalf("expected an error from Listen got %v", err)
			}
		}
	}
}
