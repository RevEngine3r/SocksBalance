# SocksBalance

> High-performance SOCKS5 load balancer with health checking and latency-based routing

## Overview

SocksBalance is a smart SOCKS5 proxy load balancer that distributes client connections across multiple backend SOCKS5 servers. It performs continuous health checks and latency measurements to ensure optimal routing performance.

### Key Features

- **Zero-copy TCP routing**: Passes SOCKS5 traffic directly without protocol decoding
- **Intelligent health checks**: Verifies backend availability and measures real latency
- **Latency-based sorting**: Routes to fastest available backends first
- **Round-robin distribution**: Evenly distributes load across healthy backends
- **Automatic failover**: Removes unhealthy backends from rotation
- **Hot reload**: Updates backend list without restart (planned)
- **Detailed metrics**: Exports health and performance stats (planned)

## How It Works

```
┌─────────┐                    ┌──────────────┐                ┌─────────────┐
│ Client  │──SOCKS5 Request──▶ │ SocksBalance │──Round Robin──▶│ Backend #1  │
│ (App)   │                    │              │                │ (SOCKS5)    │
└─────────┘                    │  Health      │                └─────────────┘
                               │  Checker     │                        
                               │      +       │                ┌─────────────┐
                               │  Latency     │──────────────▶ │ Backend #2  │
                               │  Sorter      │                │ (SOCKS5)    │
                               └──────────────┘                └─────────────┘
                                      │                                
                                      │                        ┌─────────────┐
                                      └───────────────────────▶│ Backend #3  │
                                                               │ (SOCKS5)    │
                                                               └─────────────┘
```

1. **Client connects** to SocksBalance via SOCKS5
2. **Health checker** continuously verifies backend availability
3. **Latency tester** measures real response times every 10s
4. **Balancer sorts** backends by latency (fastest first)
5. **Round-robin** selects next backend from sorted list
6. **TCP proxy** routes connection transparently

## Installation

### Pre-built Binaries

Download from [Releases](https://github.com/RevEngine3r/SocksBalance/releases) page.

### Build from Source

```bash
# Clone repository
git clone https://github.com/RevEngine3r/SocksBalance.git
cd SocksBalance

# Build
go build -o socksbalance ./cmd/socksbalance

# Run
./socksbalance -config config.yaml
```

**Requirements**: Go 1.22 or later

## Configuration

Create `config.yaml` based on [config.example.yaml](./config.example.yaml):

```yaml
listen: "0.0.0.0:1080"

backends:
  - address: "proxy1.example.com:1080"
    name: "US Proxy"
  - address: "proxy2.example.com:1080"
    name: "EU Proxy"

health:
  test_url: "https://www.google.com"
  check_interval: 10s
  connect_timeout: 5s
  request_timeout: 10s
  failure_threshold: 3

balancer:
  algorithm: "roundrobin"
  sort_by_latency: true
  latency_tolerance: 50

log:
  level: "info"
  format: "text"
```

## Usage

```bash
# Start with default config
socksbalance

# Specify config file
socksbalance -config /path/to/config.yaml

# Override listen address
socksbalance -listen 127.0.0.1:8080

# Show version
socksbalance -version
```

### Client Configuration

Point your SOCKS5 client to SocksBalance:

```bash
# Example: curl through SocksBalance
curl -x socks5://localhost:1080 https://ifconfig.me

# Example: SSH through SocksBalance
ssh -o ProxyCommand="nc -X 5 -x localhost:1080 %h %p" user@remote

# Example: Browser (Firefox)
# Preferences → Network Settings → Manual proxy configuration
# SOCKS Host: localhost, Port: 1080, SOCKS v5
```

## Architecture

See [PROJECT_MAP.md](./PROJECT_MAP.md) for detailed architecture and development progress.

## Development Status

🚧 **Early Development** - See [PROGRESS.md](./PROGRESS.md) for current roadmap.

## Performance

- **Zero overhead**: Direct TCP forwarding without SOCKS5 decoding
- **Concurrent connections**: Handles thousands of simultaneous clients
- **Low latency**: Smart routing to fastest backends
- **Efficient**: Minimal CPU and memory footprint

## License

MIT License - see [LICENSE](./LICENSE) file.

## Contributing

Contributions welcome! Please open an issue or PR.

## Roadmap

- [x] Project initialization
- [ ] Configuration system
- [ ] Backend pool management
- [ ] TCP proxy server
- [ ] Health checking
- [ ] Latency measurement
- [ ] Round-robin with sorting
- [ ] CI/CD pipeline
- [ ] Metrics and monitoring
- [ ] Hot reload
- [ ] WebUI dashboard

---

**Made with ❤️ by [RevEngine3r](https://github.com/RevEngine3r)**
