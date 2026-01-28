package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/balancer"
)

// TransparentServer is a zero-copy TCP proxy that forwards raw bytes
// Client → TransparentServer → Backend SOCKS5 → Target
type TransparentServer struct {
	address  string
	balancer *balancer.Balancer
	listener net.Listener
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

// NewTransparent creates a transparent TCP forwarding proxy
func NewTransparent(address string, bal *balancer.Balancer) *TransparentServer {
	return &TransparentServer{
		address:  address,
		balancer: bal,
	}
}

// Start begins listening for connections
func (s *TransparentServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.listener = listener
	s.running = true
	s.mu.Unlock()

	log.Printf("[INFO] Transparent TCP proxy listening on %s", s.address)
	log.Printf("[INFO] Mode: Zero-copy forwarding (no SOCKS5 decoding)")

	go s.acceptLoop(ctx)

	return nil
}

// acceptLoop accepts incoming connections
func (s *TransparentServer) acceptLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()

			if !running {
				return
			}
			log.Printf("[ERROR] Failed to accept connection: %v", err)
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(ctx, conn)
	}
}

// handleConnection performs transparent TCP forwarding
func (s *TransparentServer) handleConnection(ctx context.Context, clientConn net.Conn) {
	defer s.wg.Done()
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("[INFO] New connection from %s", clientAddr)

	// Select backend via load balancer
	backend := s.balancer.Next()
	if backend == nil {
		log.Printf("[ERROR] No healthy backends available for %s", clientAddr)
		return
	}

	log.Printf("[INFO] Forwarding %s → backend %s", clientAddr, backend.Address())

	// Connect to backend
	backendConn, err := net.DialTimeout("tcp", backend.Address(), 5*time.Second)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to backend %s: %v", backend.Address(), err)
		backend.MarkFailure(3)
		return
	}
	defer backendConn.Close()

	log.Printf("[INFO] Connected to backend %s for %s", backend.Address(), clientAddr)

	// Transparent bidirectional forwarding (zero-copy)
	s.pipe(ctx, clientConn, backendConn)

	log.Printf("[INFO] Connection closed: %s ↔ %s", clientAddr, backend.Address())
}

// pipe performs zero-copy bidirectional forwarding
func (s *TransparentServer) pipe(ctx context.Context, client, backend net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Client → Backend
	go func() {
		defer wg.Done()
		io.Copy(backend, client)
		backend.(*net.TCPConn).CloseWrite() // Half-close
	}()

	// Backend → Client
	go func() {
		defer wg.Done()
		io.Copy(client, backend)
		client.(*net.TCPConn).CloseWrite() // Half-close
	}()

	// Wait for both directions or context cancellation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

// Stop gracefully shuts down the server
func (s *TransparentServer) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("server not running")
	}
	s.running = false
	s.mu.Unlock()

	log.Printf("[INFO] Stopping transparent proxy...")

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	s.wg.Wait()
	log.Printf("[INFO] Transparent proxy stopped")

	return nil
}

// Address returns the listening address
func (s *TransparentServer) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.address
}

// IsRunning returns whether the server is running
func (s *TransparentServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetListener returns the underlying network listener (for testing)
func (s *TransparentServer) GetListener() net.Listener {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener
}
