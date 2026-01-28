package backend

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")

	if b.Address() != "proxy1.example.com:1080" {
		t.Errorf("Expected address proxy1.example.com:1080, got %s", b.Address())
	}
	if b.Name() != "Proxy 1" {
		t.Errorf("Expected name 'Proxy 1', got %s", b.Name())
	}
	if !b.IsHealthy() {
		t.Error("Expected new backend to be healthy")
	}
	if b.Latency() != 0 {
		t.Errorf("Expected latency 0, got %v", b.Latency())
	}
	if b.FailureCount() != 0 {
		t.Errorf("Expected failure count 0, got %d", b.FailureCount())
	}
}

func TestSetHealthy(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")

	b.SetHealthy(false)
	if b.IsHealthy() {
		t.Error("Expected backend to be unhealthy")
	}

	b.SetHealthy(true)
	if !b.IsHealthy() {
		t.Error("Expected backend to be healthy")
	}
}

func TestLatency(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")

	latency := 50 * time.Millisecond
	b.SetLatency(latency)

	if b.Latency() != latency {
		t.Errorf("Expected latency %v, got %v", latency, b.Latency())
	}
}

func TestFailureCount(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")

	b.IncrementFailureCount()
	if b.FailureCount() != 1 {
		t.Errorf("Expected failure count 1, got %d", b.FailureCount())
	}

	b.IncrementFailureCount()
	if b.FailureCount() != 2 {
		t.Errorf("Expected failure count 2, got %d", b.FailureCount())
	}

	b.ResetFailureCount()
	if b.FailureCount() != 0 {
		t.Errorf("Expected failure count 0 after reset, got %d", b.FailureCount())
	}
}

func TestMarkSuccess(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")

	b.SetHealthy(false)
	b.IncrementFailureCount()
	b.IncrementFailureCount()

	latency := 100 * time.Millisecond
	b.MarkSuccess(latency)

	if !b.IsHealthy() {
		t.Error("Expected backend to be healthy after success")
	}
	if b.Latency() != latency {
		t.Errorf("Expected latency %v, got %v", latency, b.Latency())
	}
	if b.FailureCount() != 0 {
		t.Errorf("Expected failure count 0 after success, got %d", b.FailureCount())
	}
	if b.LastChecked().IsZero() {
		t.Error("Expected LastChecked to be set")
	}
}

func TestMarkFailure(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")
	threshold := 3

	b.MarkFailure(threshold)
	if !b.IsHealthy() {
		t.Error("Expected backend to remain healthy after 1 failure")
	}
	if b.FailureCount() != 1 {
		t.Errorf("Expected failure count 1, got %d", b.FailureCount())
	}

	b.MarkFailure(threshold)
	if !b.IsHealthy() {
		t.Error("Expected backend to remain healthy after 2 failures")
	}

	b.MarkFailure(threshold)
	if b.IsHealthy() {
		t.Error("Expected backend to be unhealthy after reaching threshold")
	}
	if b.FailureCount() != 3 {
		t.Errorf("Expected failure count 3, got %d", b.FailureCount())
	}
}

func TestLastChecked(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")

	if !b.LastChecked().IsZero() {
		t.Error("Expected LastChecked to be zero initially")
	}

	before := time.Now()
	b.SetHealthy(true)
	after := time.Now()

	lastChecked := b.LastChecked()
	if lastChecked.Before(before) || lastChecked.After(after) {
		t.Errorf("Expected LastChecked between %v and %v, got %v", before, after, lastChecked)
	}
}

func TestThreadSafety(t *testing.T) {
	b := New("proxy1.example.com:1080", "Proxy 1")
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.SetHealthy(true)
			b.IsHealthy()
			b.SetLatency(50 * time.Millisecond)
			b.Latency()
			b.IncrementFailureCount()
			b.FailureCount()
		}()
	}

	wg.Wait()
}
