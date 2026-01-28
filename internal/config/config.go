package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen   string         `yaml:"listen"`
	Backends []Backend      `yaml:"backends"`
	Health   HealthConfig   `yaml:"health"`
	Balancer BalancerConfig `yaml:"balancer"`
	Log      LogConfig      `yaml:"log"`
}

type Backend struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}

type HealthConfig struct {
	ConnectTimeout   time.Duration `yaml:"connect_timeout"`
	TestURL          string        `yaml:"test_url"`
	CheckInterval    time.Duration `yaml:"check_interval"`
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	FailureThreshold int           `yaml:"failure_threshold"`
}

type BalancerConfig struct {
	Algorithm        string `yaml:"algorithm"`
	SortByLatency    bool   `yaml:"sort_by_latency"`
	LatencyTolerance int    `yaml:"latency_tolerance"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load reads and parses a YAML configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	applyDefaults(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Listen == "" {
		cfg.Listen = "0.0.0.0:1080"
	}
	if cfg.Health.ConnectTimeout == 0 {
		cfg.Health.ConnectTimeout = 5 * time.Second
	}
	if cfg.Health.CheckInterval == 0 {
		cfg.Health.CheckInterval = 10 * time.Second
	}
	if cfg.Health.RequestTimeout == 0 {
		cfg.Health.RequestTimeout = 10 * time.Second
	}
	if cfg.Health.FailureThreshold == 0 {
		cfg.Health.FailureThreshold = 3
	}
	if cfg.Balancer.Algorithm == "" {
		cfg.Balancer.Algorithm = "roundrobin"
	}
	if cfg.Balancer.LatencyTolerance == 0 {
		cfg.Balancer.LatencyTolerance = 50
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}
}

func validate(cfg *Config) error {
	if len(cfg.Backends) == 0 {
		return fmt.Errorf("at least one backend required")
	}

	for i, b := range cfg.Backends {
		if b.Address == "" {
			return fmt.Errorf("backend[%d]: address is required", i)
		}
	}

	if cfg.Health.ConnectTimeout < 0 {
		return fmt.Errorf("connect_timeout must be positive")
	}
	if cfg.Health.CheckInterval < 0 {
		return fmt.Errorf("check_interval must be positive")
	}
	if cfg.Health.FailureThreshold < 1 {
		return fmt.Errorf("failure_threshold must be >= 1")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Log.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Log.Level)
	}

	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[cfg.Log.Format] {
		return fmt.Errorf("invalid log format: %s (must be text or json)", cfg.Log.Format)
	}

	return nil
}
