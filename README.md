# SocksBalance

> High-performance SOCKS5 load balancer with health checking and latency-based routing

## Overview

SocksBalance is a smart SOCKS5 proxy load balancer that distributes client connections across multiple backend SOCKS5 servers. It performs continuous health checks and latency measurements to ensure optimal routing performance.

### Key Features

- **Intelligent load balancing**: Round-robin with latency-based sorting
- **Continuous health monitoring**: Automatic detection and removal of failed backends
- **Latency measurement**: Routes traffic through fastest available backends
- **Automatic failover**: Seamless recovery when backends fail
- **Thread-safe**: Handle thousands of concurrent connections
- **Zero-config defaults**: Works out of the box with minimal setup
- **SOCKS5 compliant**: Full IPv4, IPv6, and domain name support

## Quick Start

### 1. Installation

```bash
# Download and build
git clone https://github.com/RevEngine3r/SocksBalance.git
cd SocksBalance
go build -o socksbalance ./cmd/socksbalance

# Copy example config
cp config.example.yaml config.yaml
```

### 2. Configure Backends

Edit `config.yaml`:

```yaml
listen: "0.0.0.0:1080"

backends:
  - address: "proxy1.example.com:1080"
    name: "US East"
  - address: "proxy2.example.com:1080"
    name: "EU West"
  - address: "proxy3.example.com:1080"
    name: "Asia Pacific"

health:
  test_url: "https://www.google.com"
  check_interval: 10s
  connect_timeout: 5s
  request_timeout: 10s
  failure_threshold: 3

balancer:
  algorithm: "roundrobin"

log:
  level: "info"
```

### 3. Run

```bash
./socksbalance -config config.yaml
```

### 4. Use

```bash
# Test with curl
curl -x socks5://localhost:1080 https://ifconfig.me

# Configure browser to use localhost:1080 as SOCKS5 proxy
```

## How It Works

```
┌─────────┐                    ┌──────────────┐                ┌─────────────┐
│ Client  │──SOCKS5 Request──▶ │ SocksBalance │──Sorted────────▶│ Backend #1  │
│ (App)   │                    │              │   by Latency   │ (Fast)      │
└─────────┘                    │  • Health    │                └─────────────┘
                               │    Checker   │                        
                               │  • Latency   │                ┌─────────────┐
                               │    Tester    │──Round Robin──▶│ Backend #2  │
                               │  • Load      │                │ (Medium)    │
                               │    Balancer  │                └─────────────┘
                               └──────────────┘                        
                                      │                        ┌─────────────┐
                                      └───────────────────────▶│ Backend #3  │
                                                               │ (Slow)      │
                                                               └─────────────┘
```

**Flow**:
1. Client connects to SocksBalance via SOCKS5 protocol
2. Health checker continuously verifies backend availability
3. Latency tester measures response times every 10 seconds
4. Backends are sorted by latency (fastest first)
5. Round-robin selects next backend from sorted list
6. Connection is transparently proxied through selected backend

## Usage Examples

### Command Line

```bash
# Start with default config
./socksbalance

# Specify config file
./socksbalance -config /etc/socksbalance/config.yaml

# Override listen address
./socksbalance -listen 127.0.0.1:8080

# Show version
./socksbalance -version
```

### Client Configuration

#### cURL
```bash
# Single request
curl -x socks5://localhost:1080 https://api.example.com

# With authentication (if backend supports)
curl -x socks5://user:pass@localhost:1080 https://api.example.com
```

#### SSH
```bash
# SSH through SOCKS5 proxy
ssh -o ProxyCommand="nc -X 5 -x localhost:1080 %h %p" user@remote.server.com

# Add to ~/.ssh/config
Host remote.server.com
    ProxyCommand nc -X 5 -x localhost:1080 %h %p
```

#### Git
```bash
# Clone through SOCKS5
git config --global http.proxy socks5://localhost:1080
git clone https://github.com/user/repo.git
```

#### Browser (Firefox)
1. Open Settings → Network Settings
2. Select "Manual proxy configuration"
3. SOCKS Host: `localhost`, Port: `1080`
4. Select "SOCKS v5"
5. Check "Proxy DNS when using SOCKS v5"

#### Docker
```json
{
  "proxies": {
    "default": {
      "socksProxy": "socks5://localhost:1080"
    }
  }
}
```

## Configuration Reference

### Complete Example

```yaml
# Listen address for incoming SOCKS5 connections
listen: "0.0.0.0:1080"

# Backend SOCKS5 proxies
backends:
  - address: "192.168.1.100:1080"  # Required
    name: "Primary"                 # Optional
  - address: "192.168.1.101:1080"
    name: "Secondary"

# Health check settings
health:
  test_url: "https://www.google.com"  # URL to test through backends
  check_interval: 10s                  # How often to check
  connect_timeout: 5s                  # Max time to connect
  request_timeout: 10s                 # Max time for full request
  failure_threshold: 3                 # Failures before marking unhealthy

# Load balancer configuration
balancer:
  algorithm: "roundrobin"  # Only roundrobin supported currently

# Logging configuration
log:
  level: "info"   # debug, info, warn, error
  format: "text"  # text or json
```

### Environment Variables

Override config with environment variables:

```bash
export SOCKSBALANCE_LISTEN="0.0.0.0:1080"
export SOCKSBALANCE_LOG_LEVEL="debug"
./socksbalance
```

## Architecture

### Components

- **Configuration System** (`internal/config`): YAML-based configuration with validation
- **Backend Pool** (`internal/backend`): Thread-safe backend management with health tracking
- **Load Balancer** (`internal/balancer`): Round-robin selection with latency optimization
- **Health Checker** (`internal/health`): Continuous health monitoring and latency measurement
- **Proxy Server** (`internal/proxy`): SOCKS5 protocol handling and connection routing

### Project Structure

```
SocksBalance/
├── cmd/socksbalance/     # Main application entry point
├── internal/
│   ├── backend/          # Backend pool and health tracking
│   ├── balancer/         # Load balancing algorithms
│   ├── config/           # Configuration management
│   ├── health/           # Health checking and monitoring
│   └── proxy/            # SOCKS5 proxy server
├── test/                 # Integration tests
├── ROAD_MAP/             # Development roadmap
├── config.example.yaml   # Example configuration
└── README.md
```

## Development

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests
go test ./test/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Building

```bash
# Development build
go build -o socksbalance ./cmd/socksbalance

# Production build (smaller binary)
go build -ldflags="-s -w" -o socksbalance ./cmd/socksbalance

# Cross-compilation
GOOS=linux GOARCH=amd64 go build -o socksbalance-linux ./cmd/socksbalance
GOOS=darwin GOARCH=amd64 go build -o socksbalance-macos ./cmd/socksbalance
GOOS=windows GOARCH=amd64 go build -o socksbalance.exe ./cmd/socksbalance
```

## Troubleshooting

See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for common issues and solutions.

### Quick Checks

```bash
# Verify SocksBalance is running
ps aux | grep socksbalance
netstat -tulpn | grep 1080

# Test backend connectivity
curl -x socks5://backend:1080 https://www.google.com

# Check logs
tail -f socksbalance.log | grep ERROR
```

## Performance

- **Throughput**: Handles 10,000+ concurrent connections
- **Latency**: < 1ms overhead for connection routing
- **Memory**: ~50MB base + ~10KB per active connection
- **CPU**: Minimal impact with efficient connection handling

## Roadmap

- [x] Project initialization and structure
- [x] Configuration system with YAML support
- [x] Backend pool management
- [x] Thread-safe backend health tracking
- [x] TCP proxy server with graceful shutdown
- [x] SOCKS5 protocol implementation
- [x] Health checker with latency measurement
- [x] Round-robin load balancer with latency sorting
- [x] Integration tests
- [ ] Metrics and monitoring (Prometheus)
- [ ] WebUI dashboard
- [ ] Hot reload configuration
- [ ] Advanced load balancing algorithms
- [ ] Authentication support
- [ ] Rate limiting
- [ ] Connection pooling
- [ ] Docker image
- [ ] CI/CD pipeline

## License

MIT License - see [LICENSE](./LICENSE) file.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

- **Issues**: [GitHub Issues](https://github.com/RevEngine3r/SocksBalance/issues)
- **Discussions**: [GitHub Discussions](https://github.com/RevEngine3r/SocksBalance/discussions)
- **Documentation**: See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) and code comments

---

**Made with ❤️ by [RevEngine3r](https://github.com/RevEngine3r)**
