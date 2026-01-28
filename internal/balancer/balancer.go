package balancer

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

// Balancer distributes connections across backends using round-robin
type Balancer struct {
	pool             *backend.Pool
	counter          uint32
	maxLatency       time.Duration
	stickySessionTTL time.Duration
	stickySessions   map[string]*stickySession
	mu               sync.RWMutex
}

// stickySession tracks client IP to backend mapping
type stickySession struct {
	backend   *backend.Backend
	expiry    time.Time
}

// New creates a new load balancer
func New(pool *backend.Pool, maxLatency, stickySessionTTL time.Duration) *Balancer {
	b := &Balancer{
		pool:             pool,
		counter:          0,
		maxLatency:       maxLatency,
		stickySessionTTL: stickySessionTTL,
		stickySessions:   make(map[string]*stickySession),
	}

	// Start cleanup goroutine for expired sessions
	if stickySessionTTL > 0 {
		go b.cleanupExpiredSessions()
	}

	return b
}

// Next selects the next backend using round-robin on latency-sorted backends
// Supports sticky sessions based on client IP
// Returns nil if no healthy backends are available
func (b *Balancer) Next(clientAddr string) *backend.Backend {
	// Extract client IP from address
	clientIP := extractIP(clientAddr)

	// Check for sticky session
	if b.stickySessionTTL > 0 && clientIP != "" {
		b.mu.RLock()
		session, exists := b.stickySessions[clientIP]
		b.mu.RUnlock()

		if exists && time.Now().Before(session.expiry) {
			// Check if backend is still healthy and meets latency requirement
			if session.backend.IsHealthy() {
				if b.maxLatency == 0 || session.backend.Latency() <= b.maxLatency {
					// Extend session TTL
					b.mu.Lock()
					session.expiry = time.Now().Add(b.stickySessionTTL)
					b.mu.Unlock()
					return session.backend
				}
			}
			// Backend no longer valid, remove session
			b.mu.Lock()
			delete(b.stickySessions, clientIP)
			b.mu.Unlock()
		}
	}

	// Get healthy backends sorted by latency (lowest first)
	backends := b.pool.SortByLatency()
	if len(backends) == 0 {
		return nil
	}

	// Filter by max latency if configured
	if b.maxLatency > 0 {
		filtered := make([]*backend.Backend, 0)
		for _, be := range backends {
			if be.Latency() <= b.maxLatency {
				filtered = append(filtered, be)
			}
		}
		// If filtering removes all backends, use all healthy backends
		if len(filtered) > 0 {
			backends = filtered
		}
	}

	if len(backends) == 0 {
		return nil
	}

	// Round-robin selection on filtered backends
	idx := atomic.AddUint32(&b.counter, 1)
	selectedIdx := int(idx-1) % len(backends)
	selected := backends[selectedIdx]

	// Create sticky session if enabled
	if b.stickySessionTTL > 0 && clientIP != "" {
		b.mu.Lock()
		b.stickySessions[clientIP] = &stickySession{
			backend: selected,
			expiry:  time.Now().Add(b.stickySessionTTL),
		}
		b.mu.Unlock()
	}

	return selected
}

// extractIP extracts IP address from "ip:port" format
func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr // Return as-is if parsing fails
	}
	return host
}

// cleanupExpiredSessions periodically removes expired sticky sessions
func (b *Balancer) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		b.mu.Lock()
		for ip, session := range b.stickySessions {
			if now.After(session.expiry) {
				delete(b.stickySessions, ip)
			}
		}
		b.mu.Unlock()
	}
}

// GetPool returns the backend pool for testing
func (b *Balancer) GetPool() *backend.Pool {
	return b.pool
}

// GetStickySessionCount returns number of active sticky sessions (for monitoring)
func (b *Balancer) GetStickySessionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.stickySessions)
}
