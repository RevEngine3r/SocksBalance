package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

// BackendConfig represents a single backend server or port range
type BackendConfig struct {
	Address string `yaml:"address"` // Supports "host:port" or "host:port-port" for ranges
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
	Algorithm         string        `yaml:"algorithm"`
	MaxLatency        time.Duration `yaml:"max_latency"`         // Only use backends with latency <= this value (0 = no limit)
	StickySessionTTL  time.Duration `yaml:"sticky_session_ttl"`  // How long to keep client -> backend mapping (0 = disabled)
	MaxActiveBackends int           `yaml:"max_active_backends"` // Maximum number of backends to use concurrently (0 = use all)
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

		// Validate address format (supports ranges)
		if _, err := ParseAddress(b.Address); err != nil {
			return fmt.Errorf("backend %d (%s): invalid address: %w", i, b.Name, err)
		}
	}

	if c.Balancer.MaxActiveBackends < 0 {
		return fmt.Errorf("max_active_backends cannot be negative")
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
	if c.Balancer.MaxLatency == 0 {
		c.Balancer.MaxLatency = 0 // 0 = no limit (use all backends)
	}
	if c.Balancer.StickySessionTTL == 0 {
		c.Balancer.StickySessionTTL = 5 * time.Minute // Default 5 minutes
	}
	// MaxActiveBackends defaults to 0 (use all backends)

	// Log defaults
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "text"
	}
}

// ParseAddress parses a backend address and returns individual addresses
// Supports:
//   - "host:port" (returns single address)
//   - "host:port-port" (returns range of addresses)
//   - "[ipv6]:port" or "[ipv6]:port-port"
func ParseAddress(addr string) ([]string, error) {
	// Handle IPv6 addresses
	var host string
	var portPart string

	if strings.HasPrefix(addr, "[") {
		// IPv6 format: [host]:port or [host]:port-port
		closingBracket := strings.Index(addr, "]")
		if closingBracket == -1 {
			return nil, fmt.Errorf("invalid IPv6 format: missing closing bracket")
		}
		host = addr[1:closingBracket]
		if len(addr) <= closingBracket+1 || addr[closingBracket+1] != ':' {
			return nil, fmt.Errorf("invalid IPv6 format: missing port")
		}
		portPart = addr[closingBracket+2:]
	} else {
		// IPv4 or hostname: host:port or host:port-port
		lastColon := strings.LastIndex(addr, ":")
		if lastColon == -1 {
			return nil, fmt.Errorf("invalid address format: missing port")
		}
		host = addr[:lastColon]
		portPart = addr[lastColon+1:]
	}

	// Check if port part contains a range
	if strings.Contains(portPart, "-") {
		// Port range: start-end
		parts := strings.SplitN(portPart, "-", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port range format")
		}

		startPort, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start port: %w", err)
		}

		endPort, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end port: %w", err)
		}

		if startPort < 1 || startPort > 65535 {
			return nil, fmt.Errorf("start port out of range: %d", startPort)
		}

		if endPort < 1 || endPort > 65535 {
			return nil, fmt.Errorf("end port out of range: %d", endPort)
		}

		if startPort > endPort {
			return nil, fmt.Errorf("start port (%d) greater than end port (%d)", startPort, endPort)
		}

		if endPort-startPort > 1000 {
			return nil, fmt.Errorf("port range too large (max 1000): %d-%d", startPort, endPort)
		}

		// Expand range
		addresses := make([]string, 0, endPort-startPort+1)
		for port := startPort; port <= endPort; port++ {
			if strings.Contains(host, ":") {
				// IPv6
				addresses = append(addresses, fmt.Sprintf("[%s]:%d", host, port))
			} else {
				// IPv4 or hostname
				addresses = append(addresses, fmt.Sprintf("%s:%d", host, port))
			}
		}
		return addresses, nil
	} else {
		// Single port
		port, err := strconv.Atoi(strings.TrimSpace(portPart))
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}

		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("port out of range: %d", port)
		}

		return []string{addr}, nil
	}
}

// ExpandBackends expands port ranges in backend configurations
func (c *Config) ExpandBackends() []BackendConfig {
	expanded := make([]BackendConfig, 0)

	for _, backend := range c.Backends {
		addresses, err := ParseAddress(backend.Address)
		if err != nil {
			// Should not happen as validation already passed
			continue
		}

		if len(addresses) == 1 {
			// Single address, keep as is
			expanded = append(expanded, backend)
		} else {
			// Expanded range, create individual backends
			for i, addr := range addresses {
				name := backend.Name
				if name != "" && len(addresses) > 1 {
					// Append port number to name for ranges
					name = fmt.Sprintf("%s#%d", backend.Name, i+1)
				}
				expanded = append(expanded, BackendConfig{
					Address: addr,
					Name:    name,
				})
			}
		}
	}

	return expanded
}
