package config

import (
	"os"
	"testing"
	"time"
)

func TestParseAddress_Single(t *testing.T) {
	addrs, err := ParseAddress("127.0.0.1:1080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addrs) != 1 {
		t.Errorf("expected 1 address, got %d", len(addrs))
	}
	if addrs[0] != "127.0.0.1:1080" {
		t.Errorf("expected 127.0.0.1:1080, got %s", addrs[0])
	}
}

func TestParseAddress_Range(t *testing.T) {
	addrs, err := ParseAddress("127.0.0.1:9070-9072")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addrs) != 3 {
		t.Errorf("expected 3 addresses, got %d", len(addrs))
	}

	expected := []string{"127.0.0.1:9070", "127.0.0.1:9071", "127.0.0.1:9072"}
	for i, addr := range addrs {
		if addr != expected[i] {
			t.Errorf("address %d: expected %s, got %s", i, expected[i], addr)
		}
	}
}

func TestParseAddress_IPv6Single(t *testing.T) {
	addrs, err := ParseAddress("[::1]:1080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addrs) != 1 {
		t.Errorf("expected 1 address, got %d", len(addrs))
	}
	if addrs[0] != "[::1]:1080" {
		t.Errorf("expected [::1]:1080, got %s", addrs[0])
	}
}

func TestParseAddress_IPv6Range(t *testing.T) {
	addrs, err := ParseAddress("[::1]:9070-9072")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addrs) != 3 {
		t.Errorf("expected 3 addresses, got %d", len(addrs))
	}

	expected := []string{"[::1]:9070", "[::1]:9071", "[::1]:9072"}
	for i, addr := range addrs {
		if addr != expected[i] {
			t.Errorf("address %d: expected %s, got %s", i, expected[i], addr)
		}
	}
}

func TestParseAddress_InvalidRange(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{"reverse range", "127.0.0.1:9072-9070"},
		{"invalid start port", "127.0.0.1:abc-9072"},
		{"invalid end port", "127.0.0.1:9070-xyz"},
		{"port too high", "127.0.0.1:70000-70001"},
		{"port too low", "127.0.0.1:0-10"},
		{"range too large", "127.0.0.1:1000-3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAddress(tt.addr)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tt.addr)
			}
		})
	}
}

func TestExpandBackends(t *testing.T) {
	cfg := &Config{
		Backends: []BackendConfig{
			{Address: "127.0.0.1:1080", Name: "Single"},
			{Address: "127.0.0.1:9070-9072", Name: "Range"},
		},
	}

	expanded := cfg.ExpandBackends()

	// Should have 1 + 3 = 4 backends
	if len(expanded) != 4 {
		t.Errorf("expected 4 expanded backends, got %d", len(expanded))
	}

	// First backend should remain unchanged
	if expanded[0].Address != "127.0.0.1:1080" {
		t.Errorf("expected first backend 127.0.0.1:1080, got %s", expanded[0].Address)
	}
	if expanded[0].Name != "Single" {
		t.Errorf("expected first backend name 'Single', got %s", expanded[0].Name)
	}

	// Range should be expanded with numbered names
	expectedRange := []struct {
		addr string
		name string
	}{
		{"127.0.0.1:9070", "Range#1"},
		{"127.0.0.1:9071", "Range#2"},
		{"127.0.0.1:9072", "Range#3"},
	}

	for i, expected := range expectedRange {
		idx := i + 1 // Skip first backend
		if expanded[idx].Address != expected.addr {
			t.Errorf("backend %d: expected address %s, got %s", idx, expected.addr, expanded[idx].Address)
		}
		if expanded[idx].Name != expected.name {
			t.Errorf("backend %d: expected name %s, got %s", idx, expected.name, expanded[idx].Name)
		}
	}
}

func TestLoad(t *testing.T) {
	configData := `
listen: "0.0.0.0:1080"
mode: "transparent"
backends:
  - address: "127.0.0.1:1080"
    name: "Test Backend"
health:
  test_url: "https://example.com"
  check_interval: 5s
  connect_timeout: 3s
  request_timeout: 8s
  failure_threshold: 2
balancer:
  algorithm: "roundrobin"
log:
  level: "debug"
  format: "json"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Listen != "0.0.0.0:1080" {
		t.Errorf("expected listen 0.0.0.0:1080, got %s", cfg.Listen)
	}

	if cfg.Mode != "transparent" {
		t.Errorf("expected mode transparent, got %s", cfg.Mode)
	}

	if len(cfg.Backends) != 1 {
		t.Errorf("expected 1 backend, got %d", len(cfg.Backends))
	}

	if cfg.Health.CheckInterval != 5*time.Second {
		t.Errorf("expected check interval 5s, got %v", cfg.Health.CheckInterval)
	}
}

func TestValidate_MissingListen(t *testing.T) {
	cfg := &Config{
		Backends: []BackendConfig{{Address: "127.0.0.1:1080"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing listen address")
	}
}

func TestValidate_NoBackends(t *testing.T) {
	cfg := &Config{
		Listen: "0.0.0.0:1080",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for no backends")
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &Config{
		Listen:   "0.0.0.0:1080",
		Backends: []BackendConfig{{Address: "127.0.0.1:1080"}},
	}

	cfg.SetDefaults()

	if cfg.Mode != "transparent" {
		t.Errorf("expected default mode transparent, got %s", cfg.Mode)
	}

	if cfg.Health.CheckInterval != 10*time.Second {
		t.Errorf("expected default check interval 10s, got %v", cfg.Health.CheckInterval)
	}

	if cfg.Balancer.Algorithm != "roundrobin" {
		t.Errorf("expected default algorithm roundrobin, got %s", cfg.Balancer.Algorithm)
	}
}
