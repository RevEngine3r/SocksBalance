package proxy

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

func TestNew(t *testing.T) {
	pool := backend.NewPool()
	server := New("127.0.0.1:0", pool)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	if server.Address() != "127.0.0.1:0" {
		t.Errorf("Expected address 127.0.0.1:0, got %s", server.Address())
	}
	if server.IsRunning() {
		t.Error("Expected server not to be running initially")
	}
}

func TestStartStop(t *testing.T) {
	pool := backend.NewPool()
	server := New("127.0.0.1:0", pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !server.IsRunning() {
		t.Error("Expected server to be running")
	}

	time.Sleep(100 * time.Millisecond)

	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if server.IsRunning() {
		t.Error("Expected server to be stopped")
	}
}

func TestStartAlreadyRunning(t *testing.T) {
	pool := backend.NewPool()
	server := New("127.0.0.1:0", pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	err = server.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running server")
	}
}

func TestStopNotRunning(t *testing.T) {
	pool := backend.NewPool()
	server := New("127.0.0.1:0", pool)

	err := server.Stop()
	if err == nil {
		t.Error("Expected error when stopping non-running server")
	}
}

func TestConnectionWithNoBackends(t *testing.T) {
	pool := backend.NewPool()
	server := New("127.0.0.1:0", pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	addr := server.listener.Addr().String()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)
}

func TestConnectionWithHealthyBackend(t *testing.T) {
	backendServer := startMockBackend(t)
	defer backendServer.Close()

	pool := backend.NewPool()
	b := backend.New(backendServer.Addr().String(), "Mock Backend")
	pool.Add(b)

	server := New("127.0.0.1:0", pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	addr := server.listener.Addr().String()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	testData := []byte("hello")
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write to server: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, len(testData))
	n, err := io.ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("Failed to read from server: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected %d bytes, got %d", len(testData), n)
	}
	if string(buf) != string(testData) {
		t.Errorf("Expected %q, got %q", testData, buf)
	}
}

func TestGracefulShutdown(t *testing.T) {
	backendServer := startMockBackend(t)
	defer backendServer.Close()

	pool := backend.NewPool()
	b := backend.New(backendServer.Addr().String(), "Mock Backend")
	pool.Add(b)

	server := New("127.0.0.1:0", pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	addr := server.listener.Addr().String()
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)

	done := make(chan error, 1)
	go func() {
		done <- server.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Stop returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Graceful shutdown timed out")
	}
}

func startMockBackend(t *testing.T) net.Listener {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock backend: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	return listener
}
