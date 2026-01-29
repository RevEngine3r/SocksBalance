package health

import (
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3)

	if cb.State() != StateClosed {
		t.Errorf("Expected initial state CLOSED, got %v", cb.State())
	}

	if !cb.IsAvailable() {
		t.Error("Circuit should be available when CLOSED")
	}
}

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(3)

	// Record 2 failures - should stay CLOSED
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateClosed {
		t.Errorf("Expected state CLOSED after 2 failures, got %v", cb.State())
	}

	// Third failure should OPEN the circuit
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("Expected state OPEN after 3 failures, got %v", cb.State())
	}

	if cb.IsAvailable() {
		t.Error("Circuit should not be available when OPEN")
	}
}

func TestCircuitBreakerResetsOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3)

	// Record 2 failures
	cb.RecordFailure()
	cb.RecordFailure()

	// Record success - should reset failure count
	cb.RecordSuccess()

	stats := cb.GetStats()
	if stats.ConsecutiveFails != 0 {
		t.Errorf("Expected consecutive failures to reset, got %d", stats.ConsecutiveFails)
	}

	// Should still be CLOSED
	if cb.State() != StateClosed {
		t.Errorf("Expected state CLOSED, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(2)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("Expected state OPEN, got %v", cb.State())
	}

	// Try to reset (should transition to HALF_OPEN after timeout)
	time.Sleep(11 * time.Second) // Wait for initial timeout

	if !cb.TryReset() {
		t.Error("TryReset should succeed after timeout")
	}

	if cb.State() != StateHalfOpen {
		t.Errorf("Expected state HALF_OPEN, got %v", cb.State())
	}

	if !cb.IsAvailable() {
		t.Error("Circuit should be available when HALF_OPEN")
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	cb := NewCircuitBreaker(2)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Transition to HALF_OPEN
	time.Sleep(11 * time.Second)
	cb.TryReset()

	if cb.State() != StateHalfOpen {
		t.Fatalf("Expected state HALF_OPEN, got %v", cb.State())
	}

	// Record success - should close the circuit
	cb.RecordSuccess()

	if cb.State() != StateClosed {
		t.Errorf("Expected state CLOSED after recovery success, got %v", cb.State())
	}
}

func TestCircuitBreakerFailedRecovery(t *testing.T) {
	cb := NewCircuitBreaker(2)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Transition to HALF_OPEN
	time.Sleep(11 * time.Second)
	cb.TryReset()

	if cb.State() != StateHalfOpen {
		t.Fatalf("Expected state HALF_OPEN, got %v", cb.State())
	}

	// Record failure - should go back to OPEN
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("Expected state OPEN after failed recovery, got %v", cb.State())
	}
}

func TestCircuitBreakerStats(t *testing.T) {
	cb := NewCircuitBreaker(3)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	stats := cb.GetStats()

	if stats.State != StateClosed {
		t.Errorf("Expected state CLOSED, got %v", stats.State)
	}

	if stats.FailureCount != 2 {
		t.Errorf("Expected 2 total failures, got %d", stats.FailureCount)
	}

	if stats.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", stats.SuccessCount)
	}

	if stats.ConsecutiveFails != 0 {
		t.Errorf("Expected 0 consecutive failures after success, got %d", stats.ConsecutiveFails)
	}
}

func TestCircuitBreakerForceReset(t *testing.T) {
	cb := NewCircuitBreaker(2)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("Expected state OPEN, got %v", cb.State())
	}

	// Force reset
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("Expected state CLOSED after reset, got %v", cb.State())
	}

	stats := cb.GetStats()
	if stats.FailureCount != 0 || stats.ConsecutiveFails != 0 {
		t.Error("Expected all counters to be reset")
	}
}

func TestCircuitStateString(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF_OPEN"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, got)
		}
	}
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	cb := NewCircuitBreaker(10)
	done := make(chan bool)

	// Simulate concurrent requests
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				if j%2 == 0 {
					cb.RecordSuccess()
				} else {
					cb.RecordFailure()
				}
				cb.IsAvailable()
				cb.GetStats()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not panic
}
