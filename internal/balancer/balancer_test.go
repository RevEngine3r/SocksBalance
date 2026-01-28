package balancer

import (
	"testing"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

func TestNew(t *testing.T) {
	pool := backend.NewPool()
	bal := New(pool)

	if bal == nil {
		t.Fatal("balancer should not be nil")
	}

	if bal.GetPool() != pool {
		t.Error("balancer pool mismatch")
	}
}

func TestNext_NoBackends(t *testing.T) {
	pool := backend.NewPool()
	bal := New(pool)

	selected := bal.Next()
	if selected != nil {
		t.Error("should return nil when no backends available")
	}
}

func TestNext_AllUnhealthy(t *testing.T) {
	pool := backend.NewPool()
	b1 := backend.New("127.0.0.1:1080")
	b1.SetHealthy(false)
	pool.Add(b1)

	bal := New(pool)
	selected := bal.Next()

	if selected != nil {
		t.Error("should return nil when all backends unhealthy")
	}
}

func TestNext_SingleBackend(t *testing.T) {
	pool := backend.NewPool()
	b1 := backend.New("127.0.0.1:1080")
	b1.SetHealthy(true)
	pool.Add(b1)

	bal := New(pool)

	// Multiple calls should return same backend
	for i := 0; i < 5; i++ {
		selected := bal.Next()
		if selected != b1 {
			t.Errorf("iteration %d: expected backend %v, got %v", i, b1.Address(), selected.Address())
		}
	}
}

func TestNext_RoundRobin(t *testing.T) {
	pool := backend.NewPool()

	// Create 3 backends with same latency
	b1 := backend.New("127.0.0.1:1080")
	b2 := backend.New("127.0.0.1:1081")
	b3 := backend.New("127.0.0.1:1082")

	b1.SetHealthy(true)
	b2.SetHealthy(true)
	b3.SetHealthy(true)

	pool.Add(b1)
	pool.Add(b2)
	pool.Add(b3)

	bal := New(pool)

	// Should cycle through backends in round-robin
	selections := make(map[string]int)
	iterations := 9 // 3 full cycles

	for i := 0; i < iterations; i++ {
		selected := bal.Next()
		if selected == nil {
			t.Fatalf("iteration %d: got nil backend", i)
		}
		selections[selected.Address()]++
	}

	// Each backend should be selected 3 times
	for addr, count := range selections {
		if count != 3 {
			t.Errorf("backend %s selected %d times, expected 3", addr, count)
		}
	}
}

func TestNext_LatencySorting(t *testing.T) {
	pool := backend.NewPool()

	// Create backends with different latencies
	b1 := backend.New("127.0.0.1:1080") // High latency
	b2 := backend.New("127.0.0.1:1081") // Low latency
	b3 := backend.New("127.0.0.1:1082") // Medium latency

	b1.SetHealthy(true)
	b2.SetHealthy(true)
	b3.SetHealthy(true)

	b1.UpdateLatency(100 * time.Millisecond)
	b2.UpdateLatency(10 * time.Millisecond)  // Lowest
	b3.UpdateLatency(50 * time.Millisecond)

	pool.Add(b1)
	pool.Add(b2)
	pool.Add(b3)

	bal := New(pool)

	// First selection should be lowest latency (b2)
	first := bal.Next()
	if first.Address() != b2.Address() {
		t.Errorf("first selection should be lowest latency backend %s, got %s", b2.Address(), first.Address())
	}

	// Second should be medium latency (b3)
	second := bal.Next()
	if second.Address() != b3.Address() {
		t.Errorf("second selection should be medium latency backend %s, got %s", b3.Address(), second.Address())
	}

	// Third should be high latency (b1)
	third := bal.Next()
	if third.Address() != b1.Address() {
		t.Errorf("third selection should be high latency backend %s, got %s", b1.Address(), third.Address())
	}

	// Fourth cycles back to b2
	fourth := bal.Next()
	if fourth.Address() != b2.Address() {
		t.Errorf("fourth selection should cycle to %s, got %s", b2.Address(), fourth.Address())
	}
}

func TestNext_MixedHealth(t *testing.T) {
	pool := backend.NewPool()

	b1 := backend.New("127.0.0.1:1080")
	b2 := backend.New("127.0.0.1:1081")
	b3 := backend.New("127.0.0.1:1082")

	b1.SetHealthy(true)
	b2.SetHealthy(false) // Unhealthy
	b3.SetHealthy(true)

	pool.Add(b1)
	pool.Add(b2)
	pool.Add(b3)

	bal := New(pool)

	// Should only select healthy backends (b1 and b3)
	selections := make(map[string]int)
	for i := 0; i < 10; i++ {
		selected := bal.Next()
		if selected == nil {
			t.Fatalf("iteration %d: got nil backend", i)
		}
		if selected.Address() == b2.Address() {
			t.Errorf("iteration %d: selected unhealthy backend %s", i, b2.Address())
		}
		selections[selected.Address()]++
	}

	// b2 should never be selected
	if count, exists := selections[b2.Address()]; exists && count > 0 {
		t.Errorf("unhealthy backend %s was selected %d times", b2.Address(), count)
	}

	// b1 and b3 should be selected
	if selections[b1.Address()] == 0 {
		t.Error("healthy backend b1 was never selected")
	}
	if selections[b3.Address()] == 0 {
		t.Error("healthy backend b3 was never selected")
	}
}

func TestNext_ConcurrentAccess(t *testing.T) {
	pool := backend.NewPool()

	// Add multiple backends
	for i := 0; i < 5; i++ {
		b := backend.New("127.0.0.1:" + string(rune(1080+i)))
		b.SetHealthy(true)
		pool.Add(b)
	}

	bal := New(pool)

	// Concurrent selections
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				selected := bal.Next()
				if selected == nil {
					t.Error("concurrent selection returned nil")
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
