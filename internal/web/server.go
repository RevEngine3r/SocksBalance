package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

// Server provides HTTP endpoints for monitoring backend status
type Server struct {
	addr    string
	pool    *backend.Pool
	httpSrv *http.Server
	mu      sync.Mutex
	running bool
}

// NewServer creates a new web server instance
func NewServer(addr string, pool *backend.Pool) *Server {
	return &Server{
		addr:    addr,
		pool:    pool,
		running: false,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/", s.handleIndex)

	s.httpSrv = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.running = true

	// Start server in goroutine
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[WEB] HTTP server error: %v", err)
		}
	}():

	log.Printf("[WEB] Server started on %s", s.addr)
	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.httpSrv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.httpSrv.Shutdown(ctx)
	s.running = false

	if err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("[WEB] Server stopped gracefully")
	return nil
}

// handleHealth returns a simple health check response
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleStats returns backend statistics (placeholder for STEP2)
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Placeholder response - will be implemented in STEP2
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Stats endpoint - to be implemented in STEP2",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handleIndex serves the dashboard UI (placeholder for STEP3)
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Placeholder HTML - will be implemented in STEP3
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<title>SocksBalance Dashboard</title>
	<meta charset="UTF-8">
</head>
<body>
	<h1>SocksBalance Dashboard</h1>
	<p>Dashboard UI will be implemented in STEP3</p>
	<p><a href="/api/stats">View Stats API</a></p>
	<p><a href="/health">View Health Check</a></p>
</body>
</html>`)
}
