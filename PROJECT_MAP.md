# SocksBalance Project Map

## Overview
High-performance SOCKS5 load balancer with health checking and latency-based routing.

## Technology Stack
- **Language:** Go 1.22+
- **Protocol:** SOCKS5 TCP proxy
- **Load Balancing:** Round-robin with latency-based sorting
- **Health Checks:** Connection test + URL latency test (10s interval)

## Architecture
```
Client (SOCKS5) → SocksBalance → [Server1, Server2, Server3...] (SOCKS5)
                      ↓
                  Health Checker
                      ↓
                  Latency Sorter
                      ↓
                  Round-Robin LB
```

## Project Structure
```
.
├── cmd/
│   └── socksbalance/
│       └── main.go          # Entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration loader
│   ├── proxy/
│   │   ├── server.go        # TCP listener & router
│   │   └── handler.go       # Connection handler
│   ├── backend/
│   │   ├── backend.go       # Backend server representation
│   │   └── pool.go          # Backend pool manager
│   ├── health/
│   │   ├── checker.go       # Health check logic
│   │   └── latency.go       # URL latency test
│   └── balancer/
│       └── roundrobin.go    # Round-robin with latency sorting
├── config.example.yaml      # Example configuration
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── .github/
│   └── workflows/
│       └── release.yml      # Build & release automation
├── ROAD_MAP/
│   └── README.md            # Feature roadmap index
└── PROGRESS.md              # Current progress tracker
```

## Core Features
1. **TCP Proxy**: Accept SOCKS5 client connections
2. **Backend Pool**: Manage multiple SOCKS5 backend servers
3. **Health Checks**: Verify backend availability
4. **Latency Testing**: Measure real-time delay to each backend
5. **Smart Routing**: Sort backends by latency, route using round-robin
6. **Configuration**: YAML-based config for backends and settings
7. **CI/CD**: GitHub Actions for cross-platform builds

## Development Workflow
See [PROGRESS.md](./PROGRESS.md) for current status and [ROAD_MAP/](./ROAD_MAP/) for detailed feature steps.
