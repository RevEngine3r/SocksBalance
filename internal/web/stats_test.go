package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RevEngine3r/SocksBalance/internal/backend"
)

func TestHandleStatsEmpty(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":8080", pool)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStatsReal(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.TotalBackends != 0 {
		t.Errorf("Expected 0 backends, got %d", stats.TotalBackends)
	}

	if stats.HealthyBackends != 0 {
		t.Errorf("Expected 0 healthy backends, got %d", stats.HealthyBackends)
	}

	if len(stats.Backends) != 0 {
		t.Errorf("Expected empty backends array, got %d items", len(stats.Backends))
	}
}

func TestHandleStatsSingleBackend(t *testing.T) {
	pool := backend.NewPool()
	b := backend.New("127.0.0.1:9050", "TestBackend")
	b.SetHealthy(true)
	b.SetLatency(100 * time.Millisecond)
	pool.Add(b)

	server := NewServer(":8080", pool)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStatsReal(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.TotalBackends != 1 {
		t.Errorf("Expected 1 backend, got %d", stats.TotalBackends)
	}

	if stats.HealthyBackends != 1 {
		t.Errorf("Expected 1 healthy backend, got %d", stats.HealthyBackends)
	}

	if len(stats.Backends) != 1 {
		t.Fatalf("Expected 1 backend in array, got %d", len(stats.Backends))
	}

	backendStat := stats.Backends[0]
	if backendStat.Address != "127.0.0.1:9050" {
		t.Errorf("Expected address 127.0.0.1:9050, got %s", backendStat.Address)
	}

	if backendStat.Name != "TestBackend" {
		t.Errorf("Expected name TestBackend, got %s", backendStat.Name)
	}

	if !backendStat.Healthy {
		t.Error("Expected backend to be healthy")
	}

	if backendStat.LatencyMs != 100 {
		t.Errorf("Expected latency 100ms, got %d", backendStat.LatencyMs)
	}
}

func TestHandleStatsMultipleBackends(t *testing.T) {
	pool := backend.NewPool()

	// Add backends with different latencies
	b1 := backend.New("127.0.0.1:9051", "Fast")
	b1.SetHealthy(true)
	b1.SetLatency(50 * time.Millisecond)
	pool.Add(b1)

	b2 := backend.New("127.0.0.1:9052", "Medium")
	b2.SetHealthy(true)
	b2.SetLatency(200 * time.Millisecond)
	pool.Add(b2)

	b3 := backend.New("127.0.0.1:9053", "Slow")
	b3.SetHealthy(true)
	b3.SetLatency(500 * time.Millisecond)
	pool.Add(b3)

	server := NewServer(":8080", pool)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStatsReal(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.TotalBackends != 3 {
		t.Errorf("Expected 3 backends, got %d", stats.TotalBackends)
	}

	if stats.HealthyBackends != 3 {
		t.Errorf("Expected 3 healthy backends, got %d", stats.HealthyBackends)
	}

	// Verify sorting (fastest first)
	if stats.Backends[0].Name != "Fast" {
		t.Errorf("Expected first backend to be 'Fast', got '%s'", stats.Backends[0].Name)
	}

	if stats.Backends[1].Name != "Medium" {
		t.Errorf("Expected second backend to be 'Medium', got '%s'", stats.Backends[1].Name)
	}

	if stats.Backends[2].Name != "Slow" {
		t.Errorf("Expected third backend to be 'Slow', got '%s'", stats.Backends[2].Name)
	}
}

func TestHandleStatsUnhealthyBackends(t *testing.T) {
	pool := backend.NewPool()

	// Healthy backend
	b1 := backend.New("127.0.0.1:9051", "Healthy")
	b1.SetHealthy(true)
	b1.SetLatency(100 * time.Millisecond)
	pool.Add(b1)

	// Unhealthy backend
	b2 := backend.New("127.0.0.1:9052", "Unhealthy")
	b2.SetHealthy(false)
	b2.SetLatency(50 * time.Millisecond)
	pool.Add(b2)

	server := NewServer(":8080", pool)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStatsReal(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.TotalBackends != 2 {
		t.Errorf("Expected 2 backends, got %d", stats.TotalBackends)
	}

	if stats.HealthyBackends != 1 {
		t.Errorf("Expected 1 healthy backend, got %d", stats.HealthyBackends)
	}

	// Healthy backends should come first
	if !stats.Backends[0].Healthy {
		t.Error("Expected first backend to be healthy")
	}

	if stats.Backends[1].Healthy {
		t.Error("Expected second backend to be unhealthy")
	}
}

func TestHandleStatsCORSHeaders(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":8080", pool)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStatsReal(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if cors := resp.Header.Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Expected CORS header *, got %s", cors)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestHandleStatsOptionsRequest(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":8080", pool)

	req := httptest.NewRequest("OPTIONS", "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStatsReal(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", resp.StatusCode)
	}
}

func TestStatsResponseTimestamp(t *testing.T) {
	pool := backend.NewPool()
	server := NewServer(":8080", pool)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	beforeTime := time.Now()
	server.handleStatsReal(w, req)
	afterTime := time.Now()

	resp := w.Result()
	defer resp.Body.Close()

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}

	// Parse timestamp
	ts, err := time.Parse(time.RFC3339, stats.Timestamp)
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	// Verify timestamp is within reasonable range
	if ts.Before(beforeTime) || ts.After(afterTime.Add(time.Second)) {
		t.Errorf("Timestamp %v is outside expected range [%v, %v]", ts, beforeTime, afterTime)
	}
}
