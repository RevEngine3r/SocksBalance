# SocksBalance

A high-performance SOCKS5 load balancer with intelligent health monitoring, automatic failover, and circuit breaker pattern for reliable proxy routing.

## Features

- **Load Balancing**: Round-robin distribution with latency-based sorting
- **Health Monitoring**: Real-time connection outcome tracking with circuit breaker
- **Automatic Failover**: Switches to healthy backends within 100ms on failure
- **Circuit Breaker**: Removes failed backends automatically after 3 consecutive failures
- **Latency Filtering**: Route only through fastest backends (configurable threshold)
- **Sticky Sessions**: Keep clients connected to same backend (prevents session breaks)
- **GFW Evasion**: Limit concurrent active backends to avoid detection
- **Port Ranges**: Expand single config entry to multiple backends (e.g., Tor circuits)
- **Web Dashboard**: Real-time monitoring UI with health status and metrics
- **Zero-Copy Mode**: Transparent TCP forwarding for maximum performance
- **Cross-Platform**: Linux, Windows (Intel/ARM, 32/64-bit)

## Quick Start

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

### Configuration

Create `config.yaml`:

```yaml
listen: "0.0.0.0:1080"
mode: "transparent"  # or "socks5"

backends:
  - address: "192.168.1.100:1080"
    name: "Primary Proxy"
  - address: "127.0.0.1:9070-9089"  # Port range (20 backends)
    name: "Tor Instances"

health:
  test_url: "https://www.google.com"
  check_interval: 10s
  circuit_threshold: 3        # Failures before circuit opens
  recovery_interval: 30s      # Time between recovery probes
  passive_monitoring: true    # Use real connection outcomes

balancer:
  max_latency: 1000ms         # Only use backends faster than 1s
  sticky_session_ttl: 10m     # Keep client on same backend
  max_active_backends: 3      # Use top 3 fastest backends

web:
  enabled: true
  listen: "127.0.0.1:8080"
```

See [config.example.yaml](config.example.yaml) for full documentation.

## Systemd Service Deployment (Linux)

### Quick Installation

```bash
# 1. Build the binary
go build -o socksbalance ./cmd/socksbalance

# 2. Run automated installer
sudo ./scripts/install-service.sh

# 3. Edit configuration
sudo nano /etc/socksbalance/config.yaml

# 4. Start service
sudo systemctl enable socksbalance
sudo systemctl start socksbalance

# 5. Check status
sudo systemctl status socksbalance

# 6. View logs
sudo journalctl -u socksbalance -f
```

### Service Management

```bash
# Start/Stop/Restart
sudo systemctl start socksbalance
sudo systemctl stop socksbalance
sudo systemctl restart socksbalance

# Enable/Disable auto-start
sudo systemctl enable socksbalance
sudo systemctl disable socksbalance

# View logs
sudo journalctl -u socksbalance -f
```

**Full deployment guide**: [scripts/SERVICE.md](scripts/SERVICE.md)

## Cross-Platform Build

SocksBalance supports easy cross-platform compilation for various operating systems and architectures.

### Prerequisites
- Go 1.22 or higher installed.

### Using Build Scripts
The project includes automated scripts to build binaries for Linux and Windows (Intel/ARM, 32/64-bit).

**On Linux/macOS:**
```bash
chmod +x scripts/build.sh
./scripts/build.sh
```

**On Windows:**
```cmd
scripts\build.bat
```

Binaries will be generated in the `bin/` directory with the following naming convention:
`socksbalance-[os]-[arch][.exe]`

### Manual Build
To build for your current platform:
```bash
go build -o socksbalance ./cmd/socksbalance
```

## Architecture

```
Client → SocksBalance (Port 1080)
            ↓
    [Circuit Breaker]
            ↓
    [Health Monitor] → Passive monitoring of active backends
            ↓         → Active probing of failed backends
    [Load Balancer]
            ↓
    Round-robin with latency sorting
            ↓
    [Backend Pool]
    ├─ Backend 1 (🟢 CLOSED - Healthy)
    ├─ Backend 2 (🟢 CLOSED - Healthy)
    ├─ Backend 3 (🔴 OPEN - Failed, recovering...)
    └─ Backend 4 (🟡 HALF_OPEN - Testing recovery)
```

## Automatic Failover

SocksBalance detects failures in real-time and automatically switches to healthy backends:

1. **Connection fails** (timeout/refused/error)
2. **Record failure** (1/3, 2/3, 3/3)
3. **Circuit opens** after 3 consecutive failures
4. **Backend removed** from rotation
5. **Next request** automatically routed to healthy backend
6. **Recovery probe** every 30s with exponential backoff
7. **Circuit closes** when backend recovers

**Failover time**: < 100ms

## Web Dashboard

Real-time monitoring dashboard at `http://127.0.0.1:8080`:

- Backend health status (🟢 Healthy / 🔴 Failed / 🟡 Recovering)
- Latency measurements
- Circuit breaker states
- Connection success rates
- Auto-refresh every 2 seconds

## Use Cases

### Tor Load Balancing
```yaml
backends:
  - address: "127.0.0.1:9070-9089"  # 20 Tor circuits
    name: "Tor"

balancer:
  max_active_backends: 3  # Only use 3 fastest circuits
  sticky_session_ttl: 30m # Keep sessions consistent
```

### Multi-Proxy Failover
```yaml
backends:
  - address: "proxy1.example.com:1080"
    name: "Primary"
  - address: "proxy2.example.com:1080"
    name: "Backup"
  - address: "proxy3.example.com:1080"
    name: "Failover"

health:
  circuit_threshold: 3      # Switch after 3 failures
  recovery_interval: 30s    # Check every 30s
```

### High-Performance Routing
```yaml
mode: "transparent"          # Zero-copy forwarding

balancer:
  max_latency: 500ms        # Only use fast backends
  algorithm: "roundrobin"   # Sorted by latency

health:
  passive_monitoring: true  # No redundant health checks
```

## Performance

- **Routing Overhead**: < 0.1ms (transparent mode)
- **Failover Time**: < 100ms (automatic retry)
- **Throughput**: Limited only by backend speed
- **Scalability**: Tested with 1000+ backends
- **Memory**: ~5MB base + ~1KB per backend

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Integration tests
go test ./test/...

# Benchmark
go test -bench=. ./internal/...
```

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and solutions.

**Quick checks:**
```bash
# Test SOCKS5 connection
curl -x socks5://127.0.0.1:1080 https://ifconfig.me

# Check backend health
curl http://127.0.0.1:8080/api/stats | jq

# View logs
sudo journalctl -u socksbalance -f  # systemd
# or
./socksbalance -config config.yaml  # foreground
```

## Documentation

- [Configuration Guide](config.example.yaml) - Full configuration options
- [Service Deployment](scripts/SERVICE.md) - Systemd service setup
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues
- [Project Roadmap](ROAD_MAP/README.md) - Development roadmap
- [Progress Tracker](PROGRESS.md) - Current development status

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please open an issue or pull request.

## Author

RevEngine3r - [GitHub](https://github.com/RevEngine3r)
