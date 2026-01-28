package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Listen   string            `yaml:"listen"`
	Mode     string            `yaml:"mode"` // "transparent" or "socks5"
	Backends []BackendConfig  `yaml:"backends"`
	Health   HealthConfig     `yaml:"health"`
	Balancer BalancerConfig   `yaml:"balancer"`
	Log      LogConfig        `yaml:"log"`
}

// BackendConfig represents a single backend server
type BackendConfig struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}

// HealthConfig represents health check settings
type HealthConfig struct {
	TestURL          string        `yaml:"test_url"`
	CheckInterval    time.Duration `yaml:"check_interval"`
	ConnectTimeout   time.Duration `yaml:"connect_timeout"`
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	FailureThreshold int           `yaml:"failure_threshold"`
}

// BalancerConfig represents load balancer settings
type BalancerConfig struct {
	Algorithm string `yaml:"algorithm"`
}

// LogConfig represents logging settings
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	cfg.SetDefaults()

	return &cfg, nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen address is required")
	}

	if len(c.Backends) == 0 {
		return fmt.Errorf("at least one backend is required")
	}

	for i, b := range c.Backends {
		if b.Address == "" {
			return fmt.Errorf("backend %d: address is required", i)
		}
	}

	return nil
}

// SetDefaults sets default values for optional fields
func (c *Config) SetDefaults() {
	// Mode defaults
	if c.Mode == "" {
		c.Mode = "transparent" // Default to transparent (zero-copy)
	}

	// Health check defaults
	if c.Health.CheckInterval == 0 {
		c.Health.CheckInterval = 10 * time.Second
	}
	if c.Health.ConnectTimeout == 0 {
		c.Health.ConnectTimeout = 5 * time.Second
	}
	if c.Health.RequestTimeout == 0 {
		c.Health.RequestTimeout = 10 * time.Second
	}
	if c.Health.FailureThreshold == 0 {
		c.Health.FailureThreshold = 3
	}
	if c.Health.TestURL == "" {
		c.Health.TestURL = "https://www.google.com"
	}

	// Balancer defaults
	if c.Balancer.Algorithm == "" {
		c.Balancer.Algorithm = "roundrobin"
	}

	// Log defaults
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "text"
	}
}
