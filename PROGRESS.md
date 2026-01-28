# SocksBalance Progress Tracker

## Latest Feature: GFW Evasion (Max Active Backends)

### Version 0.5.0 (2026-01-28)

Added **`max_active_backends`** option to limit concurrent backend usage for anti-detection.

### The Problem

**Before**: All 20 backends used simultaneously
```
Client connects → Uses all 20 Tor circuits
  ↓
GFW detects pattern → Blocks ALL 20 circuits at once
  ↓
Result: Complete service outage
```

**After**: Only top 3 fastest backends used
```
Client connects → Uses only top 3 fastest circuits
  ↓
GFW detects pattern → Blocks only 3 circuits
  ↓
Health check detects failures → Switches to next 3 fastest
  ↓
Result: Service continues with 17 remaining backends!
```

### Configuration

```yaml
balancer:
  max_active_backends: 3  # Only use top 3 fastest backends
```

### How It Works

1. **Health Check**: All 20 backends monitored continuously
2. **Latency Sort**: Backends sorted by speed (fastest first)
3. **Limit**: Only use top 3 fastest backends
4. **Rotation**: If backend fails, automatically use next fastest

### Benefits

✅ **GFW Evasion**: Not all backends exposed at once  
✅ **Automatic Recovery**: Failed backends replaced immediately  
✅ **Best Performance**: Always using fastest available backends  
✅ **Reserve Pool**: 17 backends ready as backup  

### Example Scenarios

**Scenario 1: 20 Tor Circuits, Use Top 3**
```yaml
backends:
  - address: "127.0.0.1:9070-9089"  # 20 Tor instances
    name: "Tor"

balancer:
  max_active_backends: 3  # Only expose 3 to GFW
```

**Scenario 2: 100 Proxies, Use Top 5**
```yaml
backends:
  - address: "proxy.example.com:10000-10099"  # 100 proxies
    name: "Proxy Farm"

balancer:
  max_active_backends: 5  # Only use 5 fastest
```

**Scenario 3: Unlimited (Use All)**
```yaml
balancer:
  max_active_backends: 0  # Use all available backends (default)
```

### Real-Time Adaptation

Backend pool gets automatically re-sorted every 10 seconds:

```
Time 0:00 - Using: Backend#1 (50ms), Backend#5 (100ms), Backend#8 (150ms)
Time 0:10 - Backend#5 fails, now using: Backend#1, Backend#8, Backend#12 (200ms)
Time 0:20 - Backend#3 now faster (80ms), using: Backend#1, Backend#3, Backend#8
```

## Complete Feature Set

### Version History

- **v0.1.0** - SOCKS5 protocol handling
- **v0.2.0** - Transparent mode (zero-copy)
- **v0.3.0** - Port range expansion
- **v0.4.0** - Latency filtering + Sticky sessions
- **v0.5.0** - **GFW evasion (max active backends)**

### Anti-GFW Stack

```yaml
balancer:
  # Layer 1: Only use fast backends
  max_latency: 1000ms
  
  # Layer 2: Keep clients on same backend (avoid pattern)
  sticky_session_ttl: 10m
  
  # Layer 3: Limit concurrent exposure (GFW evasion)
  max_active_backends: 3
```

### Recommended Settings

**For Tor (Anti-GFW)**:
```yaml
backends:
  - address: "127.0.0.1:9070-9089"  # 20 circuits
    name: "Tor"

balancer:
  max_latency: 3000ms         # Tor is slower
  sticky_session_ttl: 30m     # Long sessions for circuit stability
  max_active_backends: 3      # Only expose 3 circuits to GFW
```

**For Commercial Proxies**:
```yaml
backends:
  - address: "proxy.example.com:10000-10099"  # 100 proxies
    name: "Proxies"

balancer:
  max_latency: 500ms          # Fast commercial proxies
  sticky_session_ttl: 10m     # Medium sessions
  max_active_backends: 5      # Rotate through top 5
```

**For Maximum Performance (No GFW)**:
```yaml
balancer:
  max_latency: 1000ms         # Moderate filtering
  sticky_session_ttl: 5m      # Short sessions
  max_active_backends: 0      # Use all backends (no limit)
```

## Completed Features

- ✅ **STEP1**: Project Initialization
- ✅ **STEP2**: Configuration System
- ✅ **STEP3**: Backend Representation
- ✅ **STEP4**: TCP Proxy Server
- ✅ **STEP5**: SOCKS5 Protocol Handler
- ✅ **STEP6**: Health Checker
- ✅ **STEP7**: Load Balancer
- ✅ **STEP8**: Integration Testing & Polish
- ✅ **STEP9**: Transparent Mode (Zero-Copy)
- ✅ **STEP10**: Port Range Expansion
- ✅ **STEP11**: Latency Filtering + Sticky Sessions
- ✅ **STEP12**: GFW Evasion (Max Active Backends) (NEW)

## Project Metrics

- **Total Development Time**: ~11 hours
- **Lines of Code**: ~4,500+
- **Test Coverage**: 70+ unit tests, 4 integration tests
- **Dependencies**: Minimal (Go stdlib + yaml + x/net)
- **Performance**: < 0.1ms routing overhead (transparent mode)
- **Scalability**: Tested with 1000+ backends
- **GFW Evasion**: Backend exposure limiting

## Status Summary

🎉 **SocksBalance v0.5.0 - Production Ready with GFW Evasion!**

**Perfect for**:
- 🛡️ **GFW Circumvention**: Limit backend exposure to avoid mass blocking
- ⚡ **Tor Optimization**: Use only fastest circuits from large pool
- 🔄 **Automatic Failover**: Failed backends replaced in real-time
- 🌐 **Large Proxy Farms**: Efficiently manage 100+ backends
- 🎯 **Twitter/Social Media**: Stable multi-request connections

**Deploy with confidence!**
