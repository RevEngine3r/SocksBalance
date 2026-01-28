package health

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

func TestNew(t *testing.T) {
	pool := backend.NewPool()
	checker := New(pool, 5*time.Second, "http://example.com", 10*time.Second, 10*time.Second, 3)

	if checker == nil {
		t.Fatal("Expected non-nil checker")
	}
	if checker.IsRunning() {
		t.Error("Expected checker not to be running initially")
	}
}

func TestStartStop(t *testing.T) {
	pool := backend.NewPool()
	checker := New(pool, 5*time.Second, "", 100*time.Millisecond, 10*time.Second, 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := checker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start checker: %v", err)
	}

	if !checker.IsRunning() {
		t.Error("Expected checker to be running")
	}

	time.Sleep(150 * time.Millisecond)

	err = checker.Stop()
	if err != nil {
		t.Fatalf("Failed to stop checker: %v", err)
	}

	if checker.IsRunning() {
		t.Error("Expected checker to be stopped")
	}
}

func TestStartAlreadyRunning(t *testing.T) {
	pool := backend.NewPool()
	checker := New(pool, 5*time.Second, "", 10*time.Second, 10*time.Second, 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := checker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start checker: %v", err)
	}
	defer checker.Stop()

	err = checker.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running checker")
	}
}

func TestStopNotRunning(t *testing.T) {
	pool := backend.NewPool()
	checker := New(pool, 5*time.Second, "", 10*time.Second, 10*time.Second, 3)

	err := checker.Stop()
	if err == nil {
		t.Error("Expected error when stopping non-running checker")
	}
}

func TestTestConnection(t *testing.T) {
	// Start a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	pool := backend.NewPool()
	checker := New(pool, 1*time.Second, "", 10*time.Second, 10*time.Second, 3)

	// Test successful connection
	if !checker.testConnection(listener.Addr().String()) {
		t.Error("Expected successful connection test")
	}

	// Test failed connection
	if checker.testConnection("127.0.0.1:0") {
		t.Error("Expected failed connection test for invalid address")
	}
}

func TestCheckBackendConnectionOnly(t *testing.T) {
	// Start mock backend
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock backend: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	pool := backend.NewPool()
	b := backend.New(listener.Addr().String(), "Test Backend")
	pool.Add(b)

	// No URL test configured
	checker := New(pool, 1*time.Second, "", 10*time.Second, 10*time.Second, 3)

	checker.checkBackend(b)

	if !b.IsHealthy() {
		t.Error("Expected backend to be healthy after successful connection test")
	}
	if b.FailureCount() != 0 {
		t.Errorf("Expected failure count 0, got %d", b.FailureCount())
	}
}

func TestCheckBackendConnectionFailed(t *testing.T) {
	pool := backend.NewPool()
	b := backend.New("127.0.0.1:0", "Invalid Backend")
	pool.Add(b)

	checker := New(pool, 100*time.Millisecond, "", 10*time.Second, 10*time.Second, 2)

	checker.checkBackend(b)

	if b.FailureCount() != 1 {
		t.Errorf("Expected failure count 1, got %d", b.FailureCount())
	}

	// Check again to reach threshold
	checker.checkBackend(b)

	if b.IsHealthy() {
		t.Error("Expected backend to be unhealthy after reaching failure threshold")
	}
}

func TestCheckAllBackends(t *testing.T) {
	// Start two mock backends
	listener1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock backend 1: %v", err)
	}
	defer listener1.Close()

	listener2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock backend 2: %v", err)
	}
	defer listener2.Close()

	// Accept connections
	for _, l := range []net.Listener{listener1, listener2} {
		go func(listener net.Listener) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}(l)
	}

	pool := backend.NewPool()
	b1 := backend.New(listener1.Addr().String(), "Backend 1")
	b2 := backend.New(listener2.Addr().String(), "Backend 2")
	pool.Add(b1)
	pool.Add(b2)

	checker := New(pool, 1*time.Second, "", 10*time.Second, 10*time.Second, 3)

	checker.checkAll()

	if pool.CountHealthy() != 2 {
		t.Errorf("Expected 2 healthy backends, got %d", pool.CountHealthy())
	}
}

func TestMeasureLatency(t *testing.T) {
	// Skip this test if we can't create a real SOCKS5 proxy
	// This is more of an integration test
	t.Skip("Skipping latency measurement test (requires SOCKS5 backend)")
}

func TestPeriodicChecks(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock backend: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	pool := backend.NewPool()
	b := backend.New(listener.Addr().String(), "Test Backend")
	pool.Add(b)

	// Very short interval for testing
	checker := New(pool, 1*time.Second, "", 100*time.Millisecond, 10*time.Second, 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = checker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start checker: %v", err)
	}

	// Wait for at least 2 check cycles
	time.Sleep(250 * time.Millisecond)

	if !b.IsHealthy() {
		t.Error("Expected backend to be healthy after periodic checks")
	}

	checker.Stop()
}
