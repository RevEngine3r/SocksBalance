package backend

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Pool manages a collection of backend servers
type Pool struct {
	mu       sync.RWMutex
	backends []*Backend
}

// NewPool creates a new backend pool
func NewPool() *Pool {
	return &Pool{
		backends: make([]*Backend, 0),
	}
}

// Add adds a backend to the pool
func (p *Pool) Add(backend *Backend) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.backends = append(p.backends, backend)
}

// Remove removes a backend by address
func (p *Pool) Remove(address string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, b := range p.backends {
		if b.Address() == address {
			p.backends = append(p.backends[:i], p.backends[i+1:]...)
			return true
		}
	}
	return false
}

// GetAll returns all backends (copy of slice)
func (p *Pool) GetAll() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Backend, len(p.backends))
	copy(result, p.backends)
	return result
}

// GetHealthy returns only healthy backends
func (p *Pool) GetHealthy() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Backend, 0)
	for _, b := range p.backends {
		if b.IsHealthy() {
			result = append(result, b)
		}
	}
	return result
}

// GetByAddress finds a backend by address
func (p *Pool) GetByAddress(address string) (*Backend, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, b := range p.backends {
		if b.Address() == address {
			return b, nil
		}
	}
	return nil, fmt.Errorf("backend not found: %s", address)
}

// Count returns total number of backends
func (p *Pool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.backends)
}

// CountHealthy returns number of healthy backends
func (p *Pool) CountHealthy() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, b := range p.backends {
		if b.IsHealthy() {
			count++
		}
	}
	return count
}

// SortByLatency sorts healthy backends by latency (ascending)
func (p *Pool) SortByLatency() []*Backend {
	healthy := p.GetHealthy()

	sort.Slice(healthy, func(i, j int) bool {
		return healthy[i].Latency() < healthy[j].Latency()
	})

	return healthy
}

// UpdateLatency updates latency for a specific backend
func (p *Pool) UpdateLatency(address string, latency time.Duration) error {
	backend, err := p.GetByAddress(address)
	if err != nil {
		return err
	}
	backend.SetLatency(latency)
	return nil
}
