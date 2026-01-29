package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/balancer"
)

// Server represents the TCP proxy server
type Server struct {
	address  string
	balancer *balancer.Balancer
	listener net.Listener
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

// New creates a new proxy server
func New(address string, bal *balancer.Balancer) *Server {
	return &Server{
		address:  address,
		balancer: bal,
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

	log.Printf("[INFO] SOCKS5 proxy server listening on %s", s.address)

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

// handleConnection handles a single client connection with SOCKS5 protocol
func (s *Server) handleConnection(ctx context.Context, clientConn net.Conn) {
	defer s.wg.Done()
	defer clientConn.Close()

	clientAddr := clientConn.RemoteAddr().String()
	log.Printf("[INFO] New SOCKS5 connection from %s", clientAddr)

	// Perform SOCKS5 handshake
	target, err := handleSOCKS5(clientConn)
	if err != nil {
		log.Printf("[ERROR] SOCKS5 handshake failed for %s: %v", clientAddr, err)
		return
	}

	log.Printf("[INFO] SOCKS5 target for %s: %s", clientAddr, target)

	// Try to connect with automatic failover
	const maxRetries = 3
	var backend interface{}
	var backendConn net.Conn
	var connectErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get backend from load balancer (with sticky session support)
		backend = s.balancer.Next(clientAddr)
		if backend == nil {
			log.Printf("[ERROR] No healthy backends available for %s (attempt %d/%d)", clientAddr, attempt+1, maxRetries)
			if attempt == maxRetries-1 {
				sendReply(clientConn, replyHostUnreachable)
				return
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Type assertion to get backend methods
		type BackendInterface interface {
			Address() string
			Latency() time.Duration
			RecordConnectionSuccess(time.Duration)
			RecordConnectionFailure(bool)
			MarkFailure(int)
		}

		be, ok := backend.(BackendInterface)
		if !ok {
			log.Printf("[ERROR] Backend does not implement required interface")
			return
		}

		log.Printf("[INFO] Routing %s through backend %s (latency: %v) to %s (attempt %d/%d)",
			clientAddr, be.Address(), be.Latency(), target, attempt+1, maxRetries)

		// Connect to backend SOCKS5 server with timeout
		connectStart := time.Now()
		backendConn, connectErr = net.DialTimeout("tcp", be.Address(), 5*time.Second)
		connectDuration := time.Since(connectStart)

		if connectErr != nil {
			// Check if it's a timeout
			isTimeout := false
			if netErr, ok := connectErr.(net.Error); ok && netErr.Timeout() {
				isTimeout = true
				log.Printf("[ERROR] Timeout connecting to backend %s: %v", be.Address(), connectErr)
			} else {
				log.Printf("[ERROR] Failed to connect to backend %s: %v", be.Address(), connectErr)
			}

			// Record failure and trigger circuit breaker
			be.RecordConnectionFailure(isTimeout)
			be.MarkFailure(3)

			log.Printf("[WARN] Backend %s failed, trying another backend...", be.Address())

			// Try next backend
			if attempt < maxRetries-1 {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// All retries exhausted
			log.Printf("[ERROR] All backends failed for %s after %d attempts", clientAddr, maxRetries)
			sendReply(clientConn, replyHostUnreachable)
			return
		}

		log.Printf("[INFO] Connected to backend %s in %v", be.Address(), connectDuration)

		// Perform SOCKS5 handshake with backend
		handshakeStart := time.Now()
		if err := performBackendHandshake(backendConn, target); err != nil {
			log.Printf("[ERROR] Backend SOCKS5 handshake failed: %v", err)
			backendConn.Close()
			be.RecordConnectionFailure(false)
			be.MarkFailure(3)

			log.Printf("[WARN] Backend %s handshake failed, trying another backend...", be.Address())

			if attempt < maxRetries-1 {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			log.Printf("[ERROR] All backends failed handshake for %s", clientAddr)
			sendReply(clientConn, replyHostUnreachable)
			return
		}

		handshakeDuration := time.Since(handshakeStart)
		log.Printf("[INFO] Backend handshake successful in %v, relaying data for %s", handshakeDuration, clientAddr)

		// Connection successful - record success
		totalSetupTime := connectDuration + handshakeDuration
		be.RecordConnectionSuccess(totalSetupTime)

		// Relay data between client and backend
		s.relay(ctx, clientConn, backendConn, be)

		log.Printf("[INFO] Connection closed for %s", clientAddr)
		return
	}
}

// performBackendHandshake performs SOCKS5 handshake with backend server
func performBackendHandshake(conn net.Conn, target string) error {
	// Send authentication methods (NO_AUTH)
	if _, err := conn.Write([]byte{socks5Version, 1, authNone}); err != nil {
		return fmt.Errorf("failed to send auth methods: %w", err)
	}

	// Read auth response
	response := make([]byte, 2)
	if _, err := conn.Read(response); err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if response[0] != socks5Version || response[1] != authNone {
		return fmt.Errorf("backend rejected authentication")
	}

	// Parse target address
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		return fmt.Errorf("invalid target address: %w", err)
	}

	// Build CONNECT request
	req := []byte{socks5Version, cmdConnect, 0x00}

	// Add address
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			req = append(req, addrTypeIPv4)
			req = append(req, ip4...)
		} else {
			req = append(req, addrTypeIPv6)
			req = append(req, ip...)
		}
	} else {
		req = append(req, addrTypeDomain)
		req = append(req, byte(len(host)))
		req = append(req, []byte(host)...)
	}

	// Add port
	portNum := uint16(0)
	fmt.Sscanf(port, "%d", &portNum)
	req = append(req, byte(portNum>>8), byte(portNum&0xff))

	// Send CONNECT request
	if _, err := conn.Write(req); err != nil {
		return fmt.Errorf("failed to send CONNECT: %w", err)
	}

	// Read response
	resp := make([]byte, 4)
	if _, err := conn.Read(resp); err != nil {
		return fmt.Errorf("failed to read CONNECT response: %w", err)
	}

	if resp[1] != replySuccess {
		return fmt.Errorf("backend CONNECT failed with code: %d", resp[1])
	}

	// Skip bind address and port
	var addrLen int
	switch resp[3] {
	case addrTypeIPv4:
		addrLen = 4
	case addrTypeIPv6:
		addrLen = 16
	case addrTypeDomain:
		lenBuf := make([]byte, 1)
		conn.Read(lenBuf)
		addrLen = int(lenBuf[0])
	}

	// Read and discard bind address and port
	discard := make([]byte, addrLen+2)
	conn.Read(discard)

	return nil
}

// relay bidirectionally forwards data between client and backend
func (s *Server) relay(ctx context.Context, client, backend net.Conn, be interface{}) {
	var wg sync.WaitGroup
	wg.Add(2)

	var relayErr error
	var relayMu sync.Mutex

	go func() {
		defer wg.Done()
		err := s.copyData(client, backend, "client->backend")
		if err != nil {
			relayMu.Lock()
			if relayErr == nil {
				relayErr = err
			}
			relayMu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		err := s.copyData(backend, client, "backend->client")
		if err != nil {
			relayMu.Lock()
			if relayErr == nil {
				relayErr = err
			}
			relayMu.Unlock()
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Check if relay failed due to backend issues
		if relayErr != nil {
			// Record failure if error occurred during data transfer
			if beTyped, ok := be.(interface {
				RecordConnectionFailure(bool)
			}); ok {
				beTyped.RecordConnectionFailure(false)
			}
		}
	case <-ctx.Done():
	}
}

// copyData copies data from src to dst and returns any error
func (s *Server) copyData(dst, src net.Conn, direction string) error {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			_, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			return err
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

// GetListener returns the underlying network listener (for testing)
func (s *Server) GetListener() net.Listener {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener
}
