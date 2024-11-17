package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

// createTestListenerAndConn sets up a test TCP listener and returns a client connection.
// It also returns a cleanup function to close resources.
func createTestListenerAndConn(t *testing.T) (net.Listener, net.Conn, func()) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to create client connection: %v", err)
	}

	cleanup := func() {
		conn.Close()
		listener.Close()
	}

	return listener, conn, cleanup
}

// createMockConnWithDelay creates a mock connection that delays writes to simulate network slowness.
func createMockConnWithDelay(delay time.Duration) net.Conn {
	server, client := net.Pipe()

	go func() {
		defer server.Close()
		buf := make([]byte, 1024)
		for {
			_, err := server.Read(buf)
			if err != nil {
				return
			}
			time.Sleep(delay) // Introduce delay
		}
	}()

	return client
}

// createClosedMockConn returns a connection that simulates a closed state, triggering EOF.
func createClosedMockConn() net.Conn {
	server, client := net.Pipe()
	server.Close() // Close immediately to simulate a closed connection.
	return client
}

// createMockConnWithPersistentErrors creates a mock connection that always returns errors on write.
func createMockConnWithPersistentErrors() net.Conn {
	server, client := net.Pipe()

	ch := make(chan struct{})
	go func() {
		defer server.Close()
		buf := make([]byte, 1024)
		for {
			_, err := server.Read(buf)
			if err != nil {
				return
			}
			// Simulate a persistent write error by closing the server side.
			server.Close()
			ch <- struct{}{}
		}
	}()
	<-ch

	return client
}

// Additional test helpers
func createMockConnWithWriteDeadlineError() net.Conn {
	return &mockConn{
		setWriteDeadlineErr: fmt.Errorf("failed to set deadline"),
	}
}

type mockConn struct {
	net.Conn
	setWriteDeadlineErr error
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return m.setWriteDeadlineErr
}

// Test for deadline setting failure
func TestTryWriteWithCtx_SetDeadlineError(t *testing.T) {
	ctx := context.Background()
	conn := createMockConnWithWriteDeadlineError()
	data := []byte("test data")

	err := TryWriteCtx(ctx, conn, data)
	if !errors.Is(err, ErrDeadlineUnsuccessful) {
		t.Fatalf("expected ErrDeadlineUnsuccessful, got %v", err)
	}
}

// Test for context cancellation during retry
func TestTryWriteWithCtx_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	conn := createMockConnWithPersistentErrors()
	data := []byte("test data")

	// Cancel context after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := TryWriteCtx(ctx, conn, data)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// Test for partial write scenario
func TestTryWriteWithCtx_PartialWrite(t *testing.T) {
	ctx := context.Background()
	listener, conn, cleanup := createTestListenerAndConn(t)
	defer cleanup()

	largeData := make([]byte, 1<<20) // 1MB of data
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	var receivedBytes int
	go func() {
		clientConn, err := listener.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer clientConn.Close()

		buf := make([]byte, 1024)
		for {
			n, err := clientConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					t.Error(err)
				}
				return
			}
			receivedBytes += n
		}
	}()

	err := TryWriteCtx(ctx, conn, largeData)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Give some time for all data to be received
	time.Sleep(100 * time.Millisecond)
	if receivedBytes != len(largeData) {
		t.Errorf("partial write detected: expected %d bytes, got %d", len(largeData), receivedBytes)
	}
}

// Test for retry backoff timing
func TestTryWriteWithCtx_RetryBackoff(t *testing.T) {
	ctx := context.Background()
	start := time.Now()

	conn := createMockConnWithPersistentErrors()
	data := []byte("test data")

	err := TryWriteCtx(ctx, conn, data)
	duration := time.Since(start)

	// Calculate expected minimum duration based on retry backoff
	// First retry: 1s, Second: 2s, Third: 4s, etc.
	expectedMinDuration := time.Second * 7 // Sum of first few retries

	if duration < expectedMinDuration {
		t.Errorf("retry backoff too short: got %v, expected at least %v", duration, expectedMinDuration)
	}

	if !errors.Is(err, ErrFailedAfterRetries) {
		t.Fatalf("expected ErrFailedAfterRetries, got %v", err)
	}
}
