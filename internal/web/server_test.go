package web

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

func TestNewServer(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":8080", pool)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.addr != ":8080" {
		t.Errorf("Expected addr :8080, got %s", server.addr)
	}

	if server.pool != pool {
		t.Error("Pool reference not set correctly")
	}

	if server.running {
		t.Error("Server should not be running initially")
	}
}

func TestServerStartStop(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":18080", pool)

	ctx := context.Background()

	// Start server
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	if !server.running {
		t.Error("Server should be running")
	}

	// Stop server
	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if server.running {
		t.Error("Server should not be running after stop")
	}
}

func TestServerDoubleStart(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":18081", pool)

	ctx := context.Background()

	// First start
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("First start failed: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Second start should fail
	err = server.Start(ctx)
	if err == nil {
		t.Error("Expected error on double start, got nil")
	}
}

func TestServerStopWithoutStart(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":18082", pool)

	// Stop without starting should not error
	err := server.Stop()
	if err != nil {
		t.Errorf("Stop without start should not error: %v", err)
	}
}

func TestHealthEndpoint(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":18083", pool)

	ctx := context.Background()
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	resp, err := http.Get("http://localhost:18083/health")
	if err != nil {
		t.Fatalf("Failed to request health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", result["status"])
	}
}

func TestStatsEndpoint(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":18084", pool)

	ctx := context.Background()
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Test stats endpoint
	resp, err := http.Get("http://localhost:18084/api/stats")
	if err != nil {
		t.Fatalf("Failed to request stats endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if result["message"] == nil {
		t.Error("Expected 'message' field in response")
	}
}

func TestIndexEndpoint(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":18085", pool)

	ctx := context.Background()
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Test index endpoint
	resp, err := http.Get("http://localhost:18085/")
	if err != nil {
		t.Fatalf("Failed to request index endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html; charset=utf-8, got %s", contentType)
	}
}

func TestGracefulShutdown(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":18086", pool)

	ctx := context.Background()
	err := server.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Start a request in the background
	done := make(chan bool)
	go func() {
		_, err := http.Get("http://localhost:18086/health")
		if err != nil {
			t.Logf("Request during shutdown: %v", err)
		}
		done <- true
	}()

	time.Sleep(50 * time.Millisecond)

	// Shutdown should complete within timeout
	shutdownStart := time.Now()
	err = server.Stop()
	shutdownDuration := time.Since(shutdownStart)

	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	if shutdownDuration > 6*time.Second {
		t.Errorf("Shutdown took too long: %v", shutdownDuration)
	}

	<-done
}
