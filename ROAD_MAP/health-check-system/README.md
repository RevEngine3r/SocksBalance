# Feature: Health Check System

## Overview
Implement continuous health monitoring with connection tests and latency measurement.

## Steps
1. **STEP1**: Basic connection health check (TCP dial test)
2. **STEP2**: URL latency test (HTTP request through SOCKS5)
3. **STEP3**: Failure threshold and recovery logic
4. **STEP4**: Background checker with 10s interval
5. **STEP5**: Integration with backend pool

## Acceptance Criteria
- [ ] Backends marked healthy/unhealthy based on connectivity
- [ ] Real latency measured via URL test every 10s
- [ ] Failed backends removed after threshold
- [ ] Recovered backends re-added to pool
- [ ] Health status logged clearly
