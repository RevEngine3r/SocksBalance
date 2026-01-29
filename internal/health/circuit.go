package health

import (
	"sync"
	"time"
)

// CircuitState represents the current state of a circuit breaker
type CircuitState int

const (
	// StateClosed - Normal operation, backend is healthy
	StateClosed CircuitState = iota
	// StateOpen - Backend has failed, removed from rotation
	StateOpen
	// StateHalfOpen - Testing if backend has recovered
	StateHalfOpen
)

// String returns human-readable circuit state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern for backend health
type CircuitBreaker struct {
	mu sync.RWMutex

	state            CircuitState
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	lastStateChange  time.Time
	nextRetryTime    time.Time
	consecutiveFails int

	// Configuration
	failureThreshold int           // Failures before opening circuit
	successThreshold int           // Successes in HALF_OPEN before closing
	timeout          time.Duration // Initial timeout before retry
	maxTimeout       time.Duration // Maximum timeout (for exponential backoff)
}

// NewCircuitBreaker creates a new circuit breaker with default settings
func NewCircuitBreaker(failureThreshold int) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 3 // Default: 3 failures
	}

	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: 1, // Single success in HALF_OPEN closes circuit
		timeout:          10 * time.Second,
		maxTimeout:       60 * time.Second,
		lastStateChange:  time.Now(),
	}
}

// State returns the current circuit state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// IsAvailable returns whether the backend can handle requests
func (cb *CircuitBreaker) IsAvailable() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if it's time to attempt recovery
		if time.Now().After(cb.nextRetryTime) {
			// Will transition to HALF_OPEN on next check
			return false
		}
		return false
	case StateHalfOpen:
		// Allow limited probing
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful connection attempt
func (cb *CircuitBreaker) RecordSuccess() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.consecutiveFails = 0

	switch cb.state {
	case StateClosed:
		// Already healthy, reset failure count
		cb.failureCount = 0

	case StateHalfOpen:
		// Success during recovery - close the circuit
		if cb.successCount >= cb.successThreshold {
			cb.transitionTo(StateClosed)
			cb.failureCount = 0
			cb.successCount = 0
		}

	case StateOpen:
		// Should not happen (open circuits shouldn't receive traffic)
		// But if it does, treat as recovery probe success
		cb.transitionTo(StateHalfOpen)
		cb.successCount = 1
	}

	return cb.state
}

// RecordFailure records a failed connection attempt
func (cb *CircuitBreaker) RecordFailure() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.consecutiveFails++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Check if threshold exceeded
		if cb.consecutiveFails >= cb.failureThreshold {
			cb.transitionTo(StateOpen)
			cb.scheduleRetry()
		}

	case StateHalfOpen:
		// Failed during recovery - back to OPEN
		cb.transitionTo(StateOpen)
		cb.scheduleRetry()
		cb.successCount = 0

	case StateOpen:
		// Already open, just update retry time
		cb.scheduleRetry()
	}

	return cb.state
}

// TryReset attempts to transition from OPEN to HALF_OPEN for recovery probe
func (cb *CircuitBreaker) TryReset() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen && time.Now().After(cb.nextRetryTime) {
		cb.transitionTo(StateHalfOpen)
		cb.successCount = 0
		return true
	}

	return false
}

// GetStats returns current circuit breaker statistics
func (cb *CircuitBreaker) GetStats() CircuitStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitStats{
		State:            cb.state,
		FailureCount:     cb.failureCount,
		SuccessCount:     cb.successCount,
		ConsecutiveFails: cb.consecutiveFails,
		LastFailureTime:  cb.lastFailureTime,
		LastStateChange:  cb.lastStateChange,
		NextRetryTime:    cb.nextRetryTime,
	}
}

// Reset forcefully resets the circuit breaker to CLOSED state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.transitionTo(StateClosed)
	cb.failureCount = 0
	cb.successCount = 0
	cb.consecutiveFails = 0
}

// transitionTo changes circuit state (must be called with lock held)
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	if cb.state != newState {
		cb.state = newState
		cb.lastStateChange = time.Now()
	}
}

// scheduleRetry calculates next retry time with exponential backoff
func (cb *CircuitBreaker) scheduleRetry() {
	// Exponential backoff: 10s, 20s, 40s, 60s (max)
	backoffDuration := cb.timeout * time.Duration(1<<uint(cb.consecutiveFails/cb.failureThreshold))

	if backoffDuration > cb.maxTimeout {
		backoffDuration = cb.maxTimeout
	}

	cb.nextRetryTime = time.Now().Add(backoffDuration)
}

// CircuitStats contains circuit breaker statistics
type CircuitStats struct {
	State            CircuitState
	FailureCount     int
	SuccessCount     int
	ConsecutiveFails int
	LastFailureTime  time.Time
	LastStateChange  time.Time
	NextRetryTime    time.Time
}
