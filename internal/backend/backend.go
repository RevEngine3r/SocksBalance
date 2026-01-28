package backend

import (
	"sync"
	"time"
)

// Backend represents a SOCKS5 backend server with health and latency tracking
type Backend struct {
	mu           sync.RWMutex
	address      string
	name         string
	healthy      bool
	latency      time.Duration
	failureCount int
	lastChecked  time.Time
}

// New creates a new Backend instance
func New(address, name string) *Backend {
	return &Backend{
		address: address,
		name:    name,
		healthy: true,
		latency: 0,
	}
}

// Address returns the backend address
func (b *Backend) Address() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.address
}

// Name returns the backend name
func (b *Backend) Name() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.name
}

// IsHealthy returns whether the backend is healthy
func (b *Backend) IsHealthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.healthy
}

// SetHealthy updates the backend health status
func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.healthy = healthy
	b.lastChecked = time.Now()
}

// Latency returns the current latency
func (b *Backend) Latency() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.latency
}

// SetLatency updates the backend latency
func (b *Backend) SetLatency(latency time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.latency = latency
}

// FailureCount returns the consecutive failure count
func (b *Backend) FailureCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.failureCount
}

// IncrementFailureCount increments the failure counter
func (b *Backend) IncrementFailureCount() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failureCount++
}

// ResetFailureCount resets the failure counter to zero
func (b *Backend) ResetFailureCount() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failureCount = 0
}

// LastChecked returns the last health check timestamp
func (b *Backend) LastChecked() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastChecked
}

// MarkSuccess marks a successful check (resets failures, sets healthy)
func (b *Backend) MarkSuccess(latency time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.healthy = true
	b.latency = latency
	b.failureCount = 0
	b.lastChecked = time.Now()
}

// MarkFailure increments failures and optionally marks unhealthy
func (b *Backend) MarkFailure(threshold int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failureCount++
	b.lastChecked = time.Now()
	if b.failureCount >= threshold {
		b.healthy = false
	}
}
