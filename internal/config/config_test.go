package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadValidConfig(t *testing.T) {
	yaml := `
listen: "127.0.0.1:2080"
backends:
  - address: "proxy1.example.com:1080"
    name: "Proxy 1"
  - address: "proxy2.example.com:1080"
    name: "Proxy 2"
health:
  connect_timeout: 3s
  test_url: "https://www.example.com"
  check_interval: 15s
  request_timeout: 8s
  failure_threshold: 2
balancer:
  algorithm: "roundrobin"
  sort_by_latency: true
  latency_tolerance: 30
log:
  level: "debug"
  format: "json"
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Listen != "127.0.0.1:2080" {
		t.Errorf("Expected listen 127.0.0.1:2080, got %s", cfg.Listen)
	}
	if len(cfg.Backends) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(cfg.Backends))
	}
	if cfg.Backends[0].Address != "proxy1.example.com:1080" {
		t.Errorf("Expected proxy1.example.com:1080, got %s", cfg.Backends[0].Address)
	}
	if cfg.Backends[0].Name != "Proxy 1" {
		t.Errorf("Expected 'Proxy 1', got %s", cfg.Backends[0].Name)
	}
	if cfg.Health.ConnectTimeout != 3*time.Second {
		t.Errorf("Expected 3s timeout, got %v", cfg.Health.ConnectTimeout)
	}
	if cfg.Health.TestURL != "https://www.example.com" {
		t.Errorf("Expected test URL, got %s", cfg.Health.TestURL)
	}
	if cfg.Health.CheckInterval != 15*time.Second {
		t.Errorf("Expected 15s interval, got %v", cfg.Health.CheckInterval)
	}
	if cfg.Health.FailureThreshold != 2 {
		t.Errorf("Expected failure threshold 2, got %d", cfg.Health.FailureThreshold)
	}
	if cfg.Balancer.Algorithm != "roundrobin" {
		t.Errorf("Expected roundrobin, got %s", cfg.Balancer.Algorithm)
	}
	if !cfg.Balancer.SortByLatency {
		t.Error("Expected sort_by_latency true")
	}
	if cfg.Balancer.LatencyTolerance != 30 {
		t.Errorf("Expected latency tolerance 30, got %d", cfg.Balancer.LatencyTolerance)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected debug level, got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Expected json format, got %s", cfg.Log.Format)
	}
}

func TestLoadDefaults(t *testing.T) {
	yaml := `
backends:
  - address: "proxy1.example.com:1080"
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Listen != "0.0.0.0:1080" {
		t.Errorf("Expected default listen, got %s", cfg.Listen)
	}
	if cfg.Health.ConnectTimeout != 5*time.Second {
		t.Errorf("Expected default 5s connect timeout, got %v", cfg.Health.ConnectTimeout)
	}
	if cfg.Health.CheckInterval != 10*time.Second {
		t.Errorf("Expected default 10s interval, got %v", cfg.Health.CheckInterval)
	}
	if cfg.Health.RequestTimeout != 10*time.Second {
		t.Errorf("Expected default 10s request timeout, got %v", cfg.Health.RequestTimeout)
	}
	if cfg.Health.FailureThreshold != 3 {
		t.Errorf("Expected default failure threshold 3, got %d", cfg.Health.FailureThreshold)
	}
	if cfg.Balancer.Algorithm != "roundrobin" {
		t.Errorf("Expected default roundrobin, got %s", cfg.Balancer.Algorithm)
	}
	if cfg.Balancer.LatencyTolerance != 50 {
		t.Errorf("Expected default latency tolerance 50, got %d", cfg.Balancer.LatencyTolerance)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Expected default info level, got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "text" {
		t.Errorf("Expected default text format, got %s", cfg.Log.Format)
	}
}

func TestLoadMissingBackends(t *testing.T) {
	yaml := `
listen: "0.0.0.0:1080"
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing backends")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	yaml := `
backends:
  - address: "proxy1.example.com:1080"
log:
  level: "trace"
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid log level")
	}
}

func TestLoadInvalidLogFormat(t *testing.T) {
	yaml := `
backends:
  - address: "proxy1.example.com:1080"
log:
  format: "xml"
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid log format")
	}
}

func TestLoadInvalidFailureThreshold(t *testing.T) {
	yaml := `
backends:
  - address: "proxy1.example.com:1080"
health:
  failure_threshold: 0
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid failure_threshold")
	}
}

func TestLoadNegativeTimeout(t *testing.T) {
	yests := []struct {
		name string
		yaml string
	}{
		{
			name: "negative connect_timeout",
			yaml: `
backends:
  - address: "proxy1.example.com:1080"
health:
  connect_timeout: -5s
`,
		},
		{
			name: "negative check_interval",
			yaml: `
backends:
  - address: "proxy1.example.com:1080"
health:
  check_interval: -10s
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := writeTemp(t, tt.yaml)
			defer os.Remove(tmpFile)

			_, err := Load(tmpFile)
			if err == nil {
				t.Fatal("Expected error for negative timeout")
			}
		})
	}
}

func TestLoadMissingBackendAddress(t *testing.T) {
	yaml := `
backends:
  - name: "Proxy 1"
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing backend address")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	yaml := `
backends:
  - address: "proxy1.example.com:1080"
    name: "Proxy 1
`
	tmpFile := writeTemp(t, yaml)
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}
}

func writeTemp(t *testing.T, content string) string {
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	return tmpFile
}
