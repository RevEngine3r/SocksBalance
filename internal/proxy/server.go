package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

// Server represents the TCP proxy server
type Server struct {
	address  string
	pool     *backend.Pool
	listener net.Listener
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

// New creates a new proxy server
func New(address string, pool *backend.Pool) *Server {
	return &Server{
		address: address,
		pool:    pool,
	}
}

// Start begins listening for connections
func (s *Server) Start(ctx context.Context) error {
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

	log.Printf("[INFO] Proxy server listening on %s", s.address)

	go s.acceptLoop(ctx)

	return nil
}

// acceptLoop accepts incoming connections
func (s *Server) acceptLoop(ctx context.Context) {
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

// handleConnection handles a single client connection
func (s *Server) handleConnection(ctx context.Context, clientConn net.Conn) {
	defer s.wg.Done()
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("[INFO] New connection from %s", clientAddr)

	backends := s.pool.GetHealthy()
	if len(backends) == 0 {
		log.Printf("[ERROR] No healthy backends available for %s", clientAddr)
		return
	}

	backend := backends[0]
	log.Printf("[INFO] Routing %s to backend %s", clientAddr, backend.Address())

	backendConn, err := net.DialTimeout("tcp", backend.Address(), 5*time.Second)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to backend %s: %v", backend.Address(), err)
		backend.MarkFailure(3)
		return
	}
	defer backendConn.Close()

	log.Printf("[INFO] Connected to backend %s for client %s", backend.Address(), clientAddr)

	s.relay(ctx, clientConn, backendConn)
}

// relay bidirectionally forwards data between client and backend
func (s *Server) relay(ctx context.Context, client, backend net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.copyData(client, backend, "client->backend")
	}()

	go func() {
		defer wg.Done()
		s.copyData(backend, client, "backend->client")
	}()

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

// copyData copies data from src to dst
func (s *Server) copyData(dst, src net.Conn, direction string) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			_, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("server not running")
	}
	s.running = false
	s.mu.Unlock()

	log.Printf("[INFO] Stopping proxy server...")

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	s.wg.Wait()
	log.Printf("[INFO] Proxy server stopped")

	return nil
}

// Address returns the listening address
func (s *Server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.address
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
