# SocksBalance

> High-performance SOCKS5 load balancer with health checking, latency-based routing, and GFW evasion

## Overview

SocksBalance is a smart SOCKS5 proxy load balancer that distributes client connections across multiple backend SOCKS5 servers. It performs continuous health checks and latency measurements to ensure optimal routing performance while evading detection.

### Key Features

- **Two operating modes**:
  - **Transparent mode** (default): Zero-copy TCP forwarding - blazing fast!
  - **SOCKS5 mode**: Full protocol handling for advanced use cases
- **Port range expansion**: Single config entry creates multiple backends (e.g., `127.0.0.1:9070-9089`)
- **Intelligent load balancing**: Round-robin with latency-based sorting
- **Latency filtering**: Only use fast backends (configurable threshold)
- **Sticky sessions**: Keep clients on same backend for stable connections
- **GFW evasion**: Limit concurrent backend usage to avoid mass blocking
- **Continuous health monitoring**: Automatic detection and removal of failed backends
- **Automatic failover**: Seamless recovery when backends fail
- **Thread-safe**: Handle thousands of concurrent connections

## Quick Start

### 1. Installation

```bash
git clone https://github.com/RevEngine3r/SocksBalance.git
cd SocksBalance
go build -o socksbalance ./cmd/socksbalance
```

### 2. Configure for GFW Evasion

Edit `config.yaml`:

```yaml
listen: "0.0.0.0:1080"
mode: "transparent"

backends:
  # 20 Tor circuits
  - address: "127.0.0.1:9070-9089"
    name: "Tor"

balancer:
  max_latency: 2000ms         # Only use backends faster than 2s
  sticky_session_ttl: 15m     # Same client → same backend for 15min
  max_active_backends: 3      # Only use top 3 fastest (GFW evasion!)
```

### 3. Run

```bash
./socksbalance
```

Output:
```
SocksBalance v0.5.0
[INFO] Backends (total after expansion): 20
[INFO] Max Latency Filter: 2s (only use backends faster than this)
[INFO] Sticky Sessions: 15m (same client → same backend)
[INFO] Max Active Backends: 3 (only use top 3 fastest backends)
[INFO] Anti-detection mode: Only top 3 fastest backends will be used concurrently
[INFO] GFW Evasion: Rotating through top 3 fastest backends only
```

## GFW Evasion Feature

### The Problem

When using many backends simultaneously, GFW can detect patterns and block **all** backends at once:

```
❌ Without max_active_backends:
20 Tor circuits → All used → GFW detects → All 20 blocked = Total failure
```

### The Solution

Limit concurrent backend usage to avoid mass detection:

```
✅ With max_active_backends: 3
20 Tor circuits → Only top 3 used → GFW blocks 3 → Auto-switch to next 3 → 17 circuits remain!
```

### Configuration

```yaml
balancer:
  # Only use top N fastest backends concurrently
  # 0 = unlimited (use all backends)
  # Recommended: 3-5 for GFW evasion
  max_active_backends: 3
```

### How It Works

1. **All backends monitored**: Health checker tests all 20 backends
2. **Sorted by latency**: Backends ranked by speed (fastest first)
3. **Top N selected**: Only use top 3 fastest backends
4. **Auto-rotation**: If backend fails, immediately use next fastest

**Example**:
```
Available: 20 backends
Latency sorted: [50ms, 80ms, 100ms, 150ms, 200ms, ...]
Active (top 3): Backend#1 (50ms), Backend#2 (80ms), Backend#3 (100ms)

Backend#2 fails → Immediately replaced with Backend#4 (150ms)
New active: Backend#1 (50ms), Backend#3 (100ms), Backend#4 (150ms)
```

## Anti-GFW Configuration Stack

### Layer 1: Latency Filtering
```yaml
max_latency: 2000ms  # Exclude slow/dead backends
```
- Filters out backends slower than threshold
- Improves overall performance
- Reduces timeout issues

### Layer 2: Sticky Sessions
```yaml
sticky_session_ttl: 15m  # Same client → same backend
```
- Prevents connection switching mid-session
- Critical for Twitter, Instagram, etc.
- Avoids triggering anti-bot measures

### Layer 3: Limited Exposure (GFW Evasion)
```yaml
max_active_backends: 3  # Only expose 3 backends to GFW
```
- Not all backends used simultaneously
- If 3 get blocked, 17 remain as backup
- Automatic rotation when backends fail

### Complete Anti-GFW Setup

```yaml
backends:
  - address: "127.0.0.1:9070-9089"  # 20 Tor circuits
    name: "Tor"

balancer:
  max_latency: 3000ms         # Tor can be slow, allow 3s
  sticky_session_ttl: 30m     # Long sessions for stability
  max_active_backends: 3      # Only 3 circuits exposed to GFW
```

**Result**: 
- ✅ Stable connections (sticky sessions)
- ✅ Fast performance (latency filtering)
- ✅ GFW evasion (limited exposure)
- ✅ Automatic recovery (17 backup circuits)

## Configuration Reference

### Complete Example

```yaml
listen: "0.0.0.0:1080"
mode: "transparent"

backends:
  - address: "127.0.0.1:9070-9089"  # 20 backends
    name: "Tor"

health:
  test_url: "https://www.google.com"
  check_interval: 10s
  connect_timeout: 5s
  request_timeout: 10s
  failure_threshold: 3

balancer:
  algorithm: "roundrobin"
  max_latency: 2000ms          # Only fast backends
  sticky_session_ttl: 15m      # Stable connections
  max_active_backends: 3       # GFW evasion

log:
  level: "info"
  format: "text"
```

### Balancer Options

| Option | Description | Recommended |
|--------|-------------|-------------|
| `max_latency` | Only use backends faster than this | `1000ms-3000ms` |
| `sticky_session_ttl` | Keep client on same backend | `10m-30m` |
| `max_active_backends` | Limit concurrent backend usage | `3-5` for GFW, `0` for max speed |

## Use Cases

### 1. Tor Multi-Circuit (GFW Evasion)

```yaml
backends:
  - address: "127.0.0.1:9070-9089"  # 20 Tor circuits
    name: "Tor"

balancer:
  max_latency: 3000ms
  sticky_session_ttl: 30m
  max_active_backends: 3  # Only expose 3 circuits
```

**Benefits**:
- 20 circuits available, only 3 exposed to GFW
- If GFW blocks 3, auto-switch to next 3 fastest
- 17 circuits remain as backup

### 2. Large Proxy Farm (Optimize Performance)

```yaml
backends:
  - address: "proxy.example.com:10000-10099"  # 100 proxies
    name: "Farm"

balancer:
  max_latency: 500ms
  sticky_session_ttl: 10m
  max_active_backends: 5  # Only use 5 fastest
```

**Benefits**:
- Always using 5 fastest out of 100
- 95 proxies as backup reserve
- Automatic rotation on failure

### 3. Maximum Performance (No GFW)

```yaml
balancer:
  max_latency: 1000ms
  sticky_session_ttl: 5m
  max_active_backends: 0  # Use all backends (unlimited)
```

## Performance

- **Throughput**: 10,000+ concurrent connections
- **Latency**: < 0.1ms routing overhead (transparent mode)
- **Memory**: ~50MB base + ~5KB per connection
- **CPU**: Minimal (< 1% for moderate load)
- **Scalability**: Tested with 1000+ backends

## Troubleshooting

### Twitter/Images Not Loading

**Problem**: Different backends for each request  
**Solution**: Enable sticky sessions

```yaml
balancer:
  sticky_session_ttl: 15m  # Keep client on same backend
```

### All Backends Getting Blocked

**Problem**: GFW detecting all backends  
**Solution**: Limit concurrent exposure

```yaml
balancer:
  max_active_backends: 3  # Only expose 3 at a time
```

### Slow Connections

**Problem**: Using slow backends  
**Solution**: Filter by latency

```yaml
balancer:
  max_latency: 1000ms  # Only use fast backends
```

## Roadmap

- [x] Transparent mode (zero-copy)
- [x] Port range expansion
- [x] Latency filtering
- [x] Sticky sessions
- [x] **GFW evasion (max active backends)**
- [ ] Prometheus metrics
- [ ] WebUI dashboard
- [ ] Hot reload configuration
- [ ] Docker image

## Version History

- **v0.1.0** - SOCKS5 protocol handling
- **v0.2.0** - Transparent mode
- **v0.3.0** - Port range expansion
- **v0.4.0** - Latency filtering + Sticky sessions
- **v0.5.0** - **GFW evasion** (max active backends)

## License

MIT License

---

**Made with ❤️ by [RevEngine3r](https://github.com/RevEngine3r)**
