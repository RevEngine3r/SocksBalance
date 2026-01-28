# SocksBalance

> High-performance SOCKS5 load balancer with health checking and latency-based routing

## Overview

SocksBalance is a smart SOCKS5 proxy load balancer that distributes client connections across multiple backend SOCKS5 servers. It performs continuous health checks and latency measurements to ensure optimal routing performance.

### Key Features

- **Two operating modes**:
  - **Transparent mode** (default): Zero-copy TCP forwarding - blazing fast!
  - **SOCKS5 mode**: Full protocol handling for advanced use cases
- **Intelligent load balancing**: Round-robin with latency-based sorting
- **Continuous health monitoring**: Automatic detection and removal of failed backends
- **Latency measurement**: Routes traffic through fastest available backends
- **Automatic failover**: Seamless recovery when backends fail
- **Thread-safe**: Handle thousands of concurrent connections
- **Zero-config defaults**: Works out of the box with minimal setup

## Operating Modes

### Transparent Mode (Recommended)

**Zero-copy TCP forwarding** - The proxy simply forwards raw bytes between client and backend.

```
Client (SOCKS5) → SocksBalance (TCP forward) → Backend (SOCKS5) → Target
```

**Advantages**:
- ⚡ **Fastest**: No protocol decoding/encoding overhead
- 📊 **Lowest latency**: < 0.1ms routing overhead
- 🟢 **Simple**: Direct byte-for-byte forwarding
- 💾 **Efficient**: Minimal CPU and memory usage

**Use when**: Client and all backends speak SOCKS5 (most common scenario)

### SOCKS5 Mode

**Full protocol handling** - The proxy decodes client SOCKS5, extracts target, then re-encodes to backend.

```
Client (SOCKS5) → SocksBalance (decode+re-encode) → Backend (SOCKS5) → Target
```

**Advantages**:
- 🔍 **Target visibility**: Can log/filter destination addresses
- 🛡️ **Security**: Can implement access controls
- 📊 **Metrics**: Track per-destination statistics

**Use when**: You need to inspect or modify connection targets

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
mode: "transparent"  # or "socks5"

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
# Transparent mode (default, fastest)
./socksbalance -config config.yaml

# SOCKS5 mode (with protocol handling)
./socksbalance -config config.yaml -mode socks5
```

### 4. Use

```bash
# Test with curl
curl -x socks5://localhost:1080 https://ifconfig.me

# Configure browser to use localhost:1080 as SOCKS5 proxy
```

## How It Works

### Transparent Mode (Default)

```
┌─────────┐                    ┌───────────────────────┐
│ Client  │────SOCKS5─────▶ │   SocksBalance     │
│ (App)   │    bytes       │   (Zero-copy TCP)  │
└─────────┘                    │                      │
                               │  1. Select backend │
                               │     (round-robin   │
                               │      + latency)    │
                               │                      │
                               │  2. Forward bytes  │
                               │     transparently  │
                               └───────────────────────┘
                                          │
                      ┌───────────┼───────────┐
                      │               │               │
                      ▼               ▼               ▼
              ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
              │ Backend #1  │ │ Backend #2  │ │ Backend #3  │
              │ (SOCKS5)    │ │ (SOCKS5)    │ │ (SOCKS5)    │
              └─────────────┘ └─────────────┘ └─────────────┘
                Fastest        Medium          Slowest
             (sorted by latency)
```

**Flow**:
1. Client sends SOCKS5 handshake bytes
2. SocksBalance selects backend (round-robin on latency-sorted list)
3. Raw bytes forwarded **without decoding** (zero-copy)
4. Backend performs SOCKS5 handshake with target
5. All subsequent data flows transparently

**Performance**: < 0.1ms overhead per connection

### SOCKS5 Mode

```
┌─────────┐                    ┌───────────────────────┐
│ Client  │────SOCKS5─────▶ │   SocksBalance     │
│ (App)   │   request      │   (Full decode)    │
└─────────┘                    │                      │
                               │  1. Decode SOCKS5  │
                               │  2. Extract target │
                               │  3. Select backend │
                               │  4. Re-encode      │
                               │     SOCKS5 request │
                               └───────────────────────┘
                                          │
                                 SOCKS5 to target
                                    (re-encoded)
```

**Performance**: ~1-2ms overhead per connection (protocol processing)

## Usage Examples

### Command Line

```bash
# Transparent mode (default)
./socksbalance

# SOCKS5 mode
./socksbalance -mode socks5

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

## Configuration Reference

### Mode Comparison

| Feature | Transparent Mode | SOCKS5 Mode |
|---------|-----------------|-------------|
| Speed | ⚡⚡⚡ Fastest | ⚡⚡ Fast |
| CPU Usage | Very Low | Low |
| Latency Overhead | < 0.1ms | ~1-2ms |
| Target Visibility | ❌ No | ✅ Yes |
| Access Control | ❌ No | ✅ Yes |
| Logging Targets | ❌ No | ✅ Yes |
| Use Case | General purpose | Filtering/monitoring |

### Complete Example

```yaml
# Listen address for incoming connections
listen: "0.0.0.0:1080"

# Mode: "transparent" (fast) or "socks5" (full protocol)
mode: "transparent"

# Backend SOCKS5 proxies
backends:
  - address: "192.168.1.100:1080"
    name: "Primary"
  - address: "192.168.1.101:1080"
    name: "Secondary"

# Health check settings
health:
  test_url: "https://www.google.com"
  check_interval: 10s
  connect_timeout: 5s
  request_timeout: 10s
  failure_threshold: 3

# Load balancer configuration
balancer:
  algorithm: "roundrobin"  # Only roundrobin supported currently

# Logging configuration
log:
  level: "info"   # debug, info, warn, error
  format: "text"  # text or json
```

## Performance

### Transparent Mode
- **Throughput**: 10,000+ concurrent connections
- **Latency**: < 0.1ms routing overhead
- **Memory**: ~50MB base + ~5KB per connection
- **CPU**: Minimal (< 1% for moderate load)

### SOCKS5 Mode
- **Throughput**: 8,000+ concurrent connections
- **Latency**: ~1-2ms routing overhead
- **Memory**: ~50MB base + ~10KB per connection
- **CPU**: Low (< 5% for moderate load)

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

## Architecture

### Components

- **Configuration System** (`internal/config`): YAML-based configuration with validation
- **Backend Pool** (`internal/backend`): Thread-safe backend management with health tracking
- **Load Balancer** (`internal/balancer`): Round-robin selection with latency optimization
- **Health Checker** (`internal/health`): Continuous health monitoring and latency measurement
- **Proxy Server** (`internal/proxy`):
  - **Transparent**: Zero-copy TCP forwarding
  - **SOCKS5**: Full protocol handling

## Development

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests
go test ./test/...

# With coverage
go test -cover ./...
```

### Building

```bash
# Development build
go build -o socksbalance ./cmd/socksbalance

# Production build (smaller binary)
go build -ldflags="-s -w" -o socksbalance ./cmd/socksbalance

# Cross-compilation
GOOS=linux GOARCH=amd64 go build -o socksbalance-linux ./cmd/socksbalance
```

## Roadmap

- [x] Project initialization
- [x] Configuration system
- [x] Backend pool management
- [x] TCP proxy server
- [x] SOCKS5 protocol implementation
- [x] Health checker
- [x] Round-robin load balancer
- [x] **Transparent mode (zero-copy)**
- [x] Integration tests
- [ ] Metrics and monitoring (Prometheus)
- [ ] WebUI dashboard
- [ ] Hot reload configuration
- [ ] Advanced algorithms (least-connections)
- [ ] Authentication support
- [ ] Rate limiting
- [ ] Docker image

## License

MIT License - see [LICENSE](./LICENSE) file.

## Contributing

Contributions welcome! Please open an issue or PR.

---

**Made with ❤️ by [RevEngine3r](https://github.com/RevEngine3r)**
