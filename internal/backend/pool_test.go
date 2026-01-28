package backend

import (
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	p := NewPool()
	if p.Count() != 0 {
		t.Errorf("Expected empty pool, got %d backends", p.Count())
	}
}

func TestAddBackend(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")
	b2 := New("proxy2.example.com:1080", "Proxy 2")

	p.Add(b1)
	if p.Count() != 1 {
		t.Errorf("Expected 1 backend, got %d", p.Count())
	}

	p.Add(b2)
	if p.Count() != 2 {
		t.Errorf("Expected 2 backends, got %d", p.Count())
	}
}

func TestRemoveBackend(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")
	b2 := New("proxy2.example.com:1080", "Proxy 2")

	p.Add(b1)
	p.Add(b2)

	removed := p.Remove("proxy1.example.com:1080")
	if !removed {
		t.Error("Expected successful removal")
	}
	if p.Count() != 1 {
		t.Errorf("Expected 1 backend after removal, got %d", p.Count())
	}

	removed = p.Remove("nonexistent.example.com:1080")
	if removed {
		t.Error("Expected failed removal for nonexistent backend")
	}
}

func TestGetAll(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")
	b2 := New("proxy2.example.com:1080", "Proxy 2")

	p.Add(b1)
	p.Add(b2)

	all := p.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(all))
	}
}

func TestGetHealthy(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")
	b2 := New("proxy2.example.com:1080", "Proxy 2")
	b3 := New("proxy3.example.com:1080", "Proxy 3")

	p.Add(b1)
	p.Add(b2)
	p.Add(b3)

	b2.SetHealthy(false)

	healthy := p.GetHealthy()
	if len(healthy) != 2 {
		t.Errorf("Expected 2 healthy backends, got %d", len(healthy))
	}

	for _, b := range healthy {
		if !b.IsHealthy() {
			t.Errorf("Expected only healthy backends, got unhealthy: %s", b.Address())
		}
	}
}

func TestGetByAddress(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")

	p.Add(b1)

	found, err := p.GetByAddress("proxy1.example.com:1080")
	if err != nil {
		t.Fatalf("Expected to find backend, got error: %v", err)
	}
	if found.Address() != "proxy1.example.com:1080" {
		t.Errorf("Expected proxy1.example.com:1080, got %s", found.Address())
	}

	_, err = p.GetByAddress("nonexistent.example.com:1080")
	if err == nil {
		t.Error("Expected error for nonexistent backend")
	}
}

func TestCountHealthy(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")
	b2 := New("proxy2.example.com:1080", "Proxy 2")
	b3 := New("proxy3.example.com:1080", "Proxy 3")

	p.Add(b1)
	p.Add(b2)
	p.Add(b3)

	if p.CountHealthy() != 3 {
		t.Errorf("Expected 3 healthy backends, got %d", p.CountHealthy())
	}

	b2.SetHealthy(false)
	if p.CountHealthy() != 2 {
		t.Errorf("Expected 2 healthy backends, got %d", p.CountHealthy())
	}

	b1.SetHealthy(false)
	b3.SetHealthy(false)
	if p.CountHealthy() != 0 {
		t.Errorf("Expected 0 healthy backends, got %d", p.CountHealthy())
	}
}

func TestSortByLatency(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")
	b2 := New("proxy2.example.com:1080", "Proxy 2")
	b3 := New("proxy3.example.com:1080", "Proxy 3")

	b1.SetLatency(100 * time.Millisecond)
	b2.SetLatency(50 * time.Millisecond)
	b3.SetLatency(150 * time.Millisecond)

	p.Add(b1)
	p.Add(b2)
	p.Add(b3)

	sorted := p.SortByLatency()
	if len(sorted) != 3 {
		t.Errorf("Expected 3 backends, got %d", len(sorted))
	}

	if sorted[0].Latency() != 50*time.Millisecond {
		t.Errorf("Expected first backend latency 50ms, got %v", sorted[0].Latency())
	}
	if sorted[1].Latency() != 100*time.Millisecond {
		t.Errorf("Expected second backend latency 100ms, got %v", sorted[1].Latency())
	}
	if sorted[2].Latency() != 150*time.Millisecond {
		t.Errorf("Expected third backend latency 150ms, got %v", sorted[2].Latency())
	}
}

func TestSortByLatencyOnlyHealthy(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")
	b2 := New("proxy2.example.com:1080", "Proxy 2")
	b3 := New("proxy3.example.com:1080", "Proxy 3")

	b1.SetLatency(100 * time.Millisecond)
	b2.SetLatency(50 * time.Millisecond)
	b3.SetLatency(75 * time.Millisecond)

	b2.SetHealthy(false)

	p.Add(b1)
	p.Add(b2)
	p.Add(b3)

	sorted := p.SortByLatency()
	if len(sorted) != 2 {
		t.Errorf("Expected 2 healthy backends, got %d", len(sorted))
	}

	for _, b := range sorted {
		if !b.IsHealthy() {
			t.Errorf("Expected only healthy backends in sorted list, got unhealthy: %s", b.Address())
		}
	}
}

func TestUpdateLatency(t *testing.T) {
	p := NewPool()
	b1 := New("proxy1.example.com:1080", "Proxy 1")

	p.Add(b1)

	err := p.UpdateLatency("proxy1.example.com:1080", 200*time.Millisecond)
	if err != nil {
		t.Fatalf("Expected successful latency update, got error: %v", err)
	}

	if b1.Latency() != 200*time.Millisecond {
		t.Errorf("Expected latency 200ms, got %v", b1.Latency())
	}

	err = p.UpdateLatency("nonexistent.example.com:1080", 100*time.Millisecond)
	if err == nil {
		t.Error("Expected error for nonexistent backend")
	}
}
