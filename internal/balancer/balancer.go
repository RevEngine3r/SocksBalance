package balancer

import (
	"sync/atomic"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

// Balancer distributes connections across backends using round-robin
type Balancer struct {
	pool    *backend.Pool
	counter uint32
}

// New creates a new load balancer
func New(pool *backend.Pool) *Balancer {
	return &Balancer{
		pool:    pool,
		counter: 0,
	}
}

// Next selects the next backend using round-robin on latency-sorted backends
// Returns nil if no healthy backends are available
func (b *Balancer) Next() *backend.Backend {
	// Get healthy backends sorted by latency
	backends := b.pool.GetHealthy()
	if len(backends) == 0 {
		return nil
	}

	// Sort by latency (lowest first)
	b.pool.SortByLatency(backends)

	// Round-robin selection
	idx := atomic.AddUint32(&b.counter, 1)
	selectedIdx := int(idx-1) % len(backends)

	return backends[selectedIdx]
}

// GetPool returns the backend pool for testing
func (b *Balancer) GetPool() *backend.Pool {
	return b.pool
}
