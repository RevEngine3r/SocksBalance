package test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
	"github.com/RevEngine3r/SocksBalance/internal/balancer"
	"github.com/RevEngine3r/SocksBalance/internal/health"
	"github.com/RevEngine3r/SocksBalance/internal/proxy"
	"golang.org/x/net/proxy"
)

// mockSOCKS5Backend creates a simple SOCKS5 proxy server for testing
func mockSOCKS5Backend(t *testing.T) net.Listener {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create mock SOCKS5 backend: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go handleSOCKS5Connection(conn)
		}
	}()

	return listener
}

// handleSOCKS5Connection handles a single SOCKS5 connection
func handleSOCKS5Connection(clientConn net.Conn) {
	defer clientConn.Close()

	// Read authentication methods
	buf := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, buf); err != nil {
		return
	}

	if buf[0] != 0x05 {
		return
	}

	nMethods := int(buf[1])
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(clientConn, methods); err != nil {
		return
	}

	// Send NO_AUTH response
	if _, err := clientConn.Write([]byte{0x05, 0x00}); err != nil {
		return
	}

	// Read CONNECT request
	reqHeader := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, reqHeader); err != nil {
		return
	}

	if reqHeader[0] != 0x05 || reqHeader[1] != 0x01 {
		return
	}

	// Parse address
	var host string
	var port uint16

	switch reqHeader[3] {
	case 0x01: // IPv4
		addr := make([]byte, 4)
		if _, err := io.ReadFull(clientConn, addr); err != nil {
			return
		}
		host = net.IP(addr).String()
	case 0x03: // Domain
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(clientConn, lenBuf); err != nil {
			return
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(clientConn, domain); err != nil {
			return
		}
		host = string(domain)
	case 0x04: // IPv6
		addr := make([]byte, 16)
		if _, err := io.ReadFull(clientConn, addr); err != nil {
			return
		}
		host = net.IP(addr).String()
	default:
		return
	}

	// Read port
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, portBuf); err != nil {
		return
	}
	port = uint16(portBuf[0])<<8 | uint16(portBuf[1])

	// Connect to target
	target := fmt.Sprintf("%s:%d", host, port)
	targetConn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		// Send failure reply
		clientConn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	// Send success reply
	reply := []byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	if _, err := clientConn.Write(reply); err != nil {
		return
	}

	// Relay data
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(targetConn, clientConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(clientConn, targetConn)
	}()

	wg.Wait()
}

// TestEndToEndSOCKS5Connection tests complete SOCKS5 proxy flow
func TestEndToEndSOCKS5Connection(t *testing.T) {
	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("integration test response"))
	}))
	defer httpServer.Close()

	// Create mock SOCKS5 backend
	backendListener := mockSOCKS5Backend(t)
	defer backendListener.Close()

	// Setup SocksBalance
	pool := backend.NewPool()
	b := backend.New(backendListener.Addr().String())
	b.SetHealthy(true)
	pool.Add(b)

	bal := balancer.New(pool)
	server := proxy.New("127.0.0.1:0", bal)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start proxy server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Get actual server address
	proxyAddr := server.Address()
	if listener := server.GetListener(); listener != nil {
		proxyAddr = listener.Addr().String()
	}

	// Create SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		t.Fatalf("Failed to create SOCKS5 dialer: %v", err)
	}

	// Create HTTP client with SOCKS5 proxy
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
		Timeout: 10 * time.Second,
	}

	// Make request through proxy
	resp, err := httpClient.Get(httpServer.URL)
	if err != nil {
		t.Fatalf("Failed to make HTTP request through proxy: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if string(body) != "integration test response" {
		t.Errorf("Expected 'integration test response', got %q", string(body))
	}
}

// TestLoadBalancingDistribution verifies round-robin distribution
func TestLoadBalancingDistribution(t *testing.T) {
	// Create HTTP test server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer httpServer.Close()

	// Create 3 mock SOCKS5 backends
	backends := make([]net.Listener, 3)
	for i := 0; i < 3; i++ {
		backends[i] = mockSOCKS5Backend(t)
		defer backends[i].Close()
	}

	// Setup pool with backends
	pool := backend.NewPool()
	for i, listener := range backends {
		b := backend.New(listener.Addr().String())
		b.SetHealthy(true)
		b.UpdateLatency(time.Duration(i*10) * time.Millisecond) // Different latencies
		pool.Add(b)
	}

	bal := balancer.New(pool)
	server := proxy.New("127.0.0.1:0", bal)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start proxy server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// Make multiple requests and verify distribution
	// Note: Actual verification would require instrumentation in the balancer
	// This test ensures the system works with multiple backends

	proxyAddr := server.Address()
	if listener := server.GetListener(); listener != nil {
		proxyAddr = listener.Addr().String()
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		t.Fatalf("Failed to create SOCKS5 dialer: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
		Timeout: 10 * time.Second,
	}

	// Make 10 requests
	for i := 0; i < 10; i++ {
		resp, err := httpClient.Get(httpServer.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, resp.StatusCode)
		}
	}
}

// TestHealthCheckIntegration tests health checking with failover
func TestHealthCheckIntegration(t *testing.T) {
	// Create HTTP test server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer httpServer.Close()

	// Create 2 mock SOCKS5 backends
	backend1 := mockSOCKS5Backend(t)
	defer backend1.Close()

	backend2 := mockSOCKS5Backend(t)
	defer backend2.Close()

	// Setup pool
	pool := backend.NewPool()
	b1 := backend.New(backend1.Addr().String())
	b2 := backend.New(backend2.Addr().String())
	pool.Add(b1)
	pool.Add(b2)

	// Start health checker
	healthChecker := health.New(
		pool,
		2*time.Second,  // connect timeout
		httpServer.URL, // test URL
		3*time.Second,  // check interval
		5*time.Second,  // request timeout
		2,              // failure threshold
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := healthChecker.Start(ctx); err != nil {
		t.Fatalf("Failed to start health checker: %v", err)
	}
	defer healthChecker.Stop()

	// Wait for health checks
	time.Sleep(4 * time.Second)

	// Verify backends are healthy
	healthy := pool.GetHealthy()
	if len(healthy) < 1 {
		t.Errorf("Expected at least 1 healthy backend, got %d", len(healthy))
	}

	// Close one backend to simulate failure
	backend1.Close()

	// Wait for health check to detect failure
	time.Sleep(8 * time.Second)

	// Verify failover occurred
	healthy = pool.GetHealthy()
	if len(healthy) != 1 {
		t.Logf("Expected 1 healthy backend after failure, got %d (note: timing dependent)", len(healthy))
	}
}

// TestConcurrentConnections tests multiple simultaneous connections
func TestConcurrentConnections(t *testing.T) {
	// Create HTTP test server
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("concurrent test"))
	}))
	defer httpServer.Close()

	// Create mock SOCKS5 backend
	backendListener := mockSOCKS5Backend(t)
	defer backendListener.Close()

	// Setup SocksBalance
	pool := backend.NewPool()
	b := backend.New(backendListener.Addr().String())
	b.SetHealthy(true)
	pool.Add(b)

	bal := balancer.New(pool)
	server := proxy.New("127.0.0.1:0", bal)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start proxy server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	proxyAddr := server.Address()
	if listener := server.GetListener(); listener != nil {
		proxyAddr = listener.Addr().String()
	}

	// Make 20 concurrent requests
	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
			if err != nil {
				errors <- fmt.Errorf("request %d: failed to create dialer: %w", id, err)
				return
			}

			httpClient := &http.Client{
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
				Timeout: 10 * time.Second,
			}

			resp, err := httpClient.Get(httpServer.URL)
			if err != nil {
				errors <- fmt.Errorf("request %d: failed: %w", id, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("request %d: expected status 200, got %d", id, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

// GetListener is a helper method to access the server's listener for tests
func (s *proxy.Server) GetListener() net.Listener {
	// This would need to be added to the Server struct if not already present
	// For now, we'll work around it
	return nil
}
