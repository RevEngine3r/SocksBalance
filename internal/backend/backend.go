package backend

import (
	"sync"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/health"
)

// ConnectionMetrics tracks real connection outcomes
type ConnectionMetrics struct {
	mu               sync.RWMutex
	totalAttempts    int64
	successCount     int64
	failureCount     int64
	timeoutCount     int64
	lastSuccess      time.Time
	lastFailure      time.Time
	avgResponseTime  time.Duration
	recentOutcomes   []bool // Sliding window of last N outcomes (true=success)
	windowSize       int
}

// NewConnectionMetrics creates a new metrics tracker
func NewConnectionMetrics(windowSize int) *ConnectionMetrics {
	if windowSize <= 0 {
		windowSize = 10 // Default window size
	}
	return &ConnectionMetrics{
		recentOutcomes: make([]bool, 0, windowSize),
		windowSize:     windowSize,
	}
}

// RecordSuccess records a successful connection
func (cm *ConnectionMetrics) RecordSuccess(responseTime time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.totalAttempts++
	cm.successCount++
	cm.lastSuccess = time.Now()

	// Update average response time (exponential moving average)
	if cm.avgResponseTime == 0 {
		cm.avgResponseTime = responseTime
	} else {
		// EMA: new_avg = alpha * new_value + (1-alpha) * old_avg
		alpha := 0.3
		cm.avgResponseTime = time.Duration(float64(responseTime)*alpha + float64(cm.avgResponseTime)*(1-alpha))
	}

	// Update sliding window
	if len(cm.recentOutcomes) >= cm.windowSize {
		cm.recentOutcomes = cm.recentOutcomes[1:]
	}
	cm.recentOutcomes = append(cm.recentOutcomes, true)
}

// RecordFailure records a failed connection
func (cm *ConnectionMetrics) RecordFailure(isTimeout bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.totalAttempts++
	cm.failureCount++
	cm.lastFailure = time.Now()

	if isTimeout {
		cm.timeoutCount++
	}

	// Update sliding window
	if len(cm.recentOutcomes) >= cm.windowSize {
		cm.recentOutcomes = cm.recentOutcomes[1:]
	}
	cm.recentOutcomes = append(cm.recentOutcomes, false)
}

// GetSuccessRate returns success rate from sliding window (0.0 to 1.0)
func (cm *ConnectionMetrics) GetSuccessRate() float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if len(cm.recentOutcomes) == 0 {
		return 1.0 // No data = assume healthy
	}

	successCount := 0
	for _, success := range cm.recentOutcomes {
		if success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(cm.recentOutcomes))
}

// GetStats returns current metrics
func (cm *ConnectionMetrics) GetStats() MetricsStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return MetricsStats{
		TotalAttempts:   cm.totalAttempts,
		SuccessCount:    cm.successCount,
		FailureCount:    cm.failureCount,
		TimeoutCount:    cm.timeoutCount,
		LastSuccess:     cm.lastSuccess,
		LastFailure:     cm.lastFailure,
		AvgResponseTime: cm.avgResponseTime,
		SuccessRate:     cm.GetSuccessRate(),
	}
}

// MetricsStats contains connection statistics
type MetricsStats struct {
	TotalAttempts   int64
	SuccessCount    int64
	FailureCount    int64
	TimeoutCount    int64
	LastSuccess     time.Time
	LastFailure     time.Time
	AvgResponseTime time.Duration
	SuccessRate     float64
}

// Backend represents a SOCKS5 backend server with health and latency tracking
type Backend struct {
	mu            sync.RWMutex
	address       string
	name          string
	healthy       bool
	latency       time.Duration
	failureCount  int
	lastChecked   time.Time
	circuit       *health.CircuitBreaker
	metrics       *ConnectionMetrics
	inUse         bool // Flag to indicate if backend is actively handling requests
}

// New creates a new Backend instance
func New(address, name string) *Backend {
	return &Backend{
		address: address,
		name:    name,
		healthy: true,
		latency: 0,
		circuit: health.NewCircuitBreaker(3), // Default: 3 failures
		metrics: NewConnectionMetrics(10),    // Track last 10 connections
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
	return b.healthy && b.circuit.IsAvailable()
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

// RecordConnectionSuccess records a successful real connection
func (b *Backend) RecordConnectionSuccess(responseTime time.Duration) {
	b.metrics.RecordSuccess(responseTime)
	b.circuit.RecordSuccess()

	// Also update health status
	b.mu.Lock()
	b.healthy = true
	b.failureCount = 0
	b.inUse = true
	b.mu.Unlock()
}

// RecordConnectionFailure records a failed real connection
func (b *Backend) RecordConnectionFailure(isTimeout bool) {
	b.metrics.RecordFailure(isTimeout)
	state := b.circuit.RecordFailure()

	// Update health based on circuit state
	b.mu.Lock()
	b.failureCount++
	if state == health.StateOpen {
		b.healthy = false
	}
	b.mu.Unlock()
}

// GetCircuitState returns the current circuit breaker state
func (b *Backend) GetCircuitState() health.CircuitState {
	return b.circuit.State()
}

// GetConnectionMetrics returns connection metrics
func (b *Backend) GetConnectionMetrics() MetricsStats {
	return b.metrics.GetStats()
}

// IsInUse returns whether backend is actively handling requests
func (b *Backend) IsInUse() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.inUse
}

// SetInUse marks backend as actively in use
func (b *Backend) SetInUse(inUse bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inUse = inUse
}

// TryRecovery attempts to transition circuit from OPEN to HALF_OPEN
func (b *Backend) TryRecovery() bool {
	return b.circuit.TryReset()
}
