# SocksBalance

> High-performance SOCKS5 load balancer with health checking, latency-based routing, GFW evasion, and real-time web dashboard

## Overview

SocksBalance is a smart SOCKS5 proxy load balancer that distributes client connections across multiple backend SOCKS5 servers. It performs continuous health checks and latency measurements to ensure optimal routing performance while evading detection. **NEW: Real-time web dashboard** for monitoring backend status!

### Key Features

- **Real-time web dashboard**: Monitor all backends with live health status and latencies
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

### 2. Configure for GFW Evasion + Web Dashboard

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

web:
  enabled: true               # Enable web dashboard
  listen: "127.0.0.1:8080"    # Dashboard URL
  refresh_interval: 2         # Auto-refresh every 2 seconds
```

### 3. Run

```bash
./socksbalance
```

Output:
```
SocksBalance v0.6.0
[INFO] Backends (total after expansion): 20
[INFO] Max Latency Filter: 2s (only use backends faster than this)
[INFO] Sticky Sessions: 15m (same client → same backend)
[INFO] Max Active Backends: 3 (only use top 3 fastest backends)
[INFO] Web Dashboard: enabled on 127.0.0.1:8080 (refresh: 2s)
[INFO] Anti-detection mode: Only top 3 fastest backends will be used concurrently
[INFO] Access dashboard at: http://127.0.0.1:8080
```

### 4. Open Web Dashboard

Navigate to: **http://127.0.0.1:8080**

![Dashboard Preview](docs/dashboard-preview.png)

## Web Dashboard

### Features

- **Real-time monitoring**: Auto-refresh every 2 seconds
- **Color-coded latency**:
  - 🟢 Green: < 100ms (fast)
  - 🟡 Yellow: 100-500ms (medium)
  - 🔴 Red: ≥ 500ms (slow)
- **Health status badges**: ✓ Healthy / ✗ Unhealthy
- **Sorted by speed**: Fastest backends first
- **Summary stats**: Total, Healthy, Unhealthy counts
- **Responsive design**: Works on desktop, tablet, mobile
- **Modern dark theme**: Easy on the eyes

### Configuration

```yaml
web:
  # Enable web dashboard
  enabled: true
  
  # Dashboard listen address
  # SECURITY: Bind to 127.0.0.1 (localhost) by default
  # Only change to 0.0.0.0 if you need remote access
  listen: "127.0.0.1:8080"
  
  # Frontend auto-refresh interval in seconds
  refresh_interval: 2
```

### Security Notes

⚠️ **Important Security Considerations**:

1. **Disabled by default**: Must explicitly enable in config
2. **Localhost only**: Binds to `127.0.0.1` by default (not accessible from network)
3. **No authentication**: v0.6.0 has no login system (for localhost use)
4. **Read-only**: Dashboard cannot modify backends or configuration

**For remote access** (not recommended for production):
```yaml
web:
  enabled: true
  listen: "0.0.0.0:8080"  # ⚠️ Accessible from network!
```

**Better alternative**: Use SSH tunnel instead:
```bash
ssh -L 8080:localhost:8080 user@server
# Then access http://localhost:8080 locally
```

### Dashboard API

The dashboard uses a JSON API:

**Endpoint**: `GET /api/stats`

**Response**:
```json
{
  "timestamp": "2026-01-28T21:30:00+03:30",
  "total_backends": 20,
  "healthy_backends": 18,
  "backends": [
    {
      "address": "127.0.0.1:9070",
      "name": "Tor#1",
      "healthy": true,
      "latency_ms": 45,
      "last_checked": "2026-01-28T21:29:55+03:30"
    }
  ]
}
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
5. **Web dashboard**: Monitor the rotation in real-time!

**Example**:
```
Available: 20 backends
Latency sorted: [50ms, 80ms, 100ms, 150ms, 200ms, ...]
Active (top 3): Backend#1 (50ms), Backend#2 (80ms), Backend#3 (100ms)

Backend#2 fails → Immediately replaced with Backend#4 (150ms)
New active: Backend#1 (50ms), Backend#3 (100ms), Backend#4 (150ms)

👁️ Watch this happen live in the web dashboard!
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

### Layer 4: Real-time Monitoring
```yaml
web:
  enabled: true  # Watch backend health in real-time
```
- Visual confirmation of backend status
- Monitor latencies and health changes
- See GFW evasion in action

### Complete Anti-GFW Setup

```yaml
backends:
  - address: "127.0.0.1:9070-9089"  # 20 Tor circuits
    name: "Tor"

balancer:
  max_latency: 3000ms         # Tor can be slow, allow 3s
  sticky_session_ttl: 30m     # Long sessions for stability
  max_active_backends: 3      # Only 3 circuits exposed to GFW

web:
  enabled: true               # Monitor in real-time
  listen: "127.0.0.1:8080"
```

**Result**: 
- ✅ Stable connections (sticky sessions)
- ✅ Fast performance (latency filtering)
- ✅ GFW evasion (limited exposure)
- ✅ Automatic recovery (17 backup circuits)
- ✅ **Real-time visibility (web dashboard)**

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

web:
  enabled: true                # Web dashboard
  listen: "127.0.0.1:8080"     # Localhost only
  refresh_interval: 2          # Update every 2s

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

### Web Dashboard Options

| Option | Description | Default |
|--------|-------------|---------||
| `enabled` | Enable/disable dashboard | `false` (disabled) |
| `listen` | Dashboard listen address | `127.0.0.1:8080` |
| `refresh_interval` | Auto-refresh interval (seconds) | `2` |

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

web:
  enabled: true  # Monitor circuit rotation
```

**Benefits**:
- 20 circuits available, only 3 exposed to GFW
- If GFW blocks 3, auto-switch to next 3 fastest
- 17 circuits remain as backup
- **Watch rotation happen live in dashboard**

### 2. Large Proxy Farm (Optimize Performance)

```yaml
backends:
  - address: "proxy.example.com:10000-10099"  # 100 proxies
    name: "Farm"

balancer:
  max_latency: 500ms
  sticky_session_ttl: 10m
  max_active_backends: 5  # Only use 5 fastest

web:
  enabled: true  # See which proxies are fastest
```

**Benefits**:
- Always using 5 fastest out of 100
- 95 proxies as backup reserve
- Automatic rotation on failure
- **Dashboard shows speed rankings**

### 3. Maximum Performance (No GFW)

```yaml
balancer:
  max_latency: 1000ms
  sticky_session_ttl: 5m
  max_active_backends: 0  # Use all backends (unlimited)

web:
  enabled: true  # Monitor all backends
```

## Performance

- **Throughput**: 10,000+ concurrent connections
- **Latency**: < 0.1ms routing overhead (transparent mode)
- **Memory**: ~50MB base + ~5KB per connection
- **CPU**: Minimal (< 1% for moderate load)
- **Scalability**: Tested with 1000+ backends
- **Dashboard**: Negligible overhead (~0.01% CPU)

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

### Dashboard Not Accessible

**Problem**: Cannot access web dashboard  
**Solutions**:

1. **Check if enabled**:
   ```yaml
   web:
     enabled: true  # Must be true
   ```

2. **Check listen address**:
   - `127.0.0.1:8080` - Only localhost (secure)
   - `0.0.0.0:8080` - All interfaces (use with caution)

3. **Check firewall**: Ensure port 8080 is not blocked

4. **Use SSH tunnel for remote access**:
   ```bash
   ssh -L 8080:localhost:8080 user@server
   ```

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for more details.

## Roadmap

- [x] Transparent mode (zero-copy)
- [x] Port range expansion
- [x] Latency filtering
- [x] Sticky sessions
- [x] GFW evasion (max active backends)
- [x] **Web UI dashboard** ✨ NEW
- [ ] Prometheus metrics
- [ ] Hot reload configuration
- [ ] Docker image
- [ ] Authentication for web dashboard

## Version History

- **v0.1.0** - SOCKS5 protocol handling
- **v0.2.0** - Transparent mode
- **v0.3.0** - Port range expansion
- **v0.4.0** - Latency filtering + Sticky sessions
- **v0.5.0** - GFW evasion (max active backends)
- **v0.6.0** - **Web UI dashboard** 🎉

## License

MIT License

---

**Made with ❤️ by [RevEngine3r](https://github.com/RevEngine3r)**
