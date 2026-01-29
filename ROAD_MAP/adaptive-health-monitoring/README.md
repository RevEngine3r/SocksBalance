# Feature: Adaptive Health Monitoring

## Overview
Intelligent health monitoring that uses real connection outcomes for active backends instead of redundant URL tests, with automatic circuit-breaking and recovery.

## Problem Statement
Current implementation wastes resources by periodically testing backends that are already handling real traffic successfully. Additionally, it takes up to 10 seconds to detect failures (until next health check cycle).

## Solution
**Passive Health Monitoring**: Track actual connection success/failure for active backends, with circuit-breaker pattern for automatic failover.

## Architecture
```
Client Request → Proxy Handler → Backend Connection
                                       ↓
                                  [Outcome?]
                                   ↙     ↘
                           Success      Failure/Timeout
                               ↓              ↓
                        MarkSuccess()    MarkFailure()
                               ↓              ↓
                        Stay Active     Failure Counter++
                                              ↓
                                    [Threshold exceeded?]
                                              ↓
                                        Circuit OPEN
                                        (Remove from pool)
                                              ↓
                                    Background Recovery
                                    (Periodic URL test)
                                              ↓
                                        [Success?]
                                              ↓
                                        Circuit CLOSED
                                        (Add back to pool)
```

## Roadmap

### STEP 1: Connection Outcome Tracking
**Goal**: Capture success/failure/timeout for every backend connection attempt.

**Tasks**:
- Add `ConnectionMetrics` struct to track per-backend statistics
- Modify proxy handler to report connection outcomes
- Add `RecordSuccess()` and `RecordFailure()` methods to Backend
- Add sliding window for failure rate calculation (last N attempts)

**Files to modify**:
- `internal/backend/backend.go` - Add metrics fields
- `internal/proxy/handler.go` - Report outcomes after connection attempt

**Test cases**:
- Successful connection increments success counter
- Failed connection increments failure counter
- Timeout is treated as failure
- Metrics are thread-safe

---

### STEP 2: Circuit Breaker Implementation
**Goal**: Automatically remove backends from rotation after consecutive failures.

**Tasks**:
- Add `CircuitBreaker` with states: CLOSED (healthy), OPEN (failed), HALF_OPEN (recovering)
- Implement failure threshold (default: 3 consecutive failures)
- Implement automatic state transitions
- Add cooldown period before retry (exponential backoff: 10s → 30s → 60s)
- Emit events when circuit state changes (for logging/monitoring)

**Files to create**:
- `internal/health/circuit.go` - Circuit breaker logic
- `internal/health/circuit_test.go` - Unit tests

**Files to modify**:
- `internal/backend/backend.go` - Integrate circuit breaker
- `internal/backend/pool.go` - Filter out backends with OPEN circuits

**Test cases**:
- Circuit opens after N failures
- Circuit stays closed on success
- HALF_OPEN allows single probe request
- Exponential backoff works correctly

---

### STEP 3: Dual-Mode Health Checking
**Goal**: Use passive monitoring for active backends, periodic URL tests for failed/idle backends.

**Tasks**:
- Refactor `Checker` to support two modes:
  - **Passive mode**: No URL tests, relies on connection outcomes
  - **Active mode**: Periodic URL tests for recovery detection
- Add logic to switch backends between modes based on circuit state
- Implement recovery prober for OPEN circuits
- Add configuration: `passive_monitoring: true` (default)

**Files to modify**:
- `internal/health/checker.go` - Add mode switching logic
- `internal/config/config.go` - Add passive monitoring config
- `config.example.yaml` - Document new option

**Test cases**:
- Active backends skip URL tests
- Failed backends receive periodic probes
- Recovery probe success closes circuit
- Config toggle works correctly

---

### STEP 4: Integration & Configuration
**Goal**: Wire everything together and provide user-facing configuration.

**Tasks**:
- Update main.go to initialize adaptive monitoring
- Add metrics exposure (success rate, failure rate, circuit state)
- Update web dashboard to show circuit breaker states
- Add configuration options:
  ```yaml
  health:
    passive_monitoring: true        # Use connection outcomes
    failure_threshold: 3             # Failures before circuit opens
    recovery_interval: 30s           # Time between recovery probes
    failure_window: 10               # Track last N attempts
  ```
- Update README and TROUBLESHOOTING docs

**Files to modify**:
- `cmd/socksbalance/main.go` - Initialize adaptive monitoring
- `internal/web/stats.go` - Add circuit state to API
- `internal/web/dashboard.go` - Display circuit state (🟢 CLOSED, 🔴 OPEN, 🟡 HALF_OPEN)
- `config.example.yaml` - Add new health config
- `README.md` - Document adaptive monitoring
- `TROUBLESHOOTING.md` - Add debugging tips

**Test cases**:
- End-to-end: Failed connection triggers circuit, recovery closes it
- Dashboard shows correct circuit states
- Config validation works
- All unit tests pass

---

### STEP 5: Performance Testing & Polish
**Goal**: Validate performance improvements and edge case handling.

**Tasks**:
- Benchmark: Compare passive vs. active monitoring overhead
- Load test: Verify failover under high connection rates (1000+ req/s)
- Edge cases:
  - All backends fail simultaneously
  - Single backend flapping (up/down repeatedly)
  - Recovery during high load
- Add metrics logging: "Backend X failed 3 times, circuit opened"
- Add prometheus-style metrics (optional future enhancement)

**Files to create**:
- `test/integration_adaptive_test.go` - Integration tests
- `test/benchmark_monitoring_test.go` - Performance benchmarks

**Test cases**:
- Passive monitoring has <1ms overhead
- Failover happens within 100ms of failure
- Recovery completes within configured interval
- System survives all-backend failure gracefully

---

## Completion Criteria

- [x] Connection outcomes tracked for all backends
- [x] Circuit breaker opens after N failures
- [x] Circuit breaker closes after successful recovery
- [x] Active backends skip redundant URL tests
- [x] Failed backends receive periodic recovery probes
- [x] Web dashboard shows circuit states
- [x] Configuration options documented
- [x] All unit tests pass (20+ new tests)
- [x] Integration tests validate end-to-end flow
- [x] Performance benchmark shows <1ms overhead
- [x] Documentation updated (README, TROUBLESHOOTING)

## Expected Outcomes

**Performance**:
- 90% reduction in health check HTTP requests (only probe failed backends)
- <100ms failover time (vs. up to 10s with periodic checks)
- <1ms passive monitoring overhead per connection

**Reliability**:
- Immediate detection of connection failures
- Automatic failover without waiting for health check cycle
- Graceful recovery with exponential backoff
- No thundering herd (staggered recovery probes)

**Observability**:
- Real-time circuit state in web dashboard (🟢🟡🔴)
- Detailed failure metrics per backend
- Event logging for all circuit state changes
