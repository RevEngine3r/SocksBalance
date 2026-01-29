# STEP 1: Connection Outcome Tracking

## Goal
Capture and store success/failure/timeout outcomes for every backend connection attempt to enable passive health monitoring.

## Tasks

### 1. Add ConnectionMetrics struct
**File**: `internal/backend/backend.go`

```go
type ConnectionMetrics struct {
    mu              sync.RWMutex
    totalAttempts   uint64
    successCount    uint64
    failureCount    uint64
    consecutiveFails int
    lastAttempt     time.Time
    recentOutcomes  []bool  // Sliding window (true=success, false=failure)
    windowSize      int     // Size of sliding window
}
```

### 2. Add methods to Backend struct

```go
// RecordSuccess records a successful connection
func (b *Backend) RecordSuccess() {
    // Update metrics
    // Reset consecutive failure counter
    // Add true to sliding window
}

// RecordFailure records a failed connection
func (b *Backend) RecordFailure() {
    // Update metrics
    // Increment consecutive failure counter
    // Add false to sliding window
}

// GetSuccessRate returns success rate from sliding window
func (b *Backend) GetSuccessRate() float64 {
    // Calculate percentage of successful attempts in window
}

// GetConsecutiveFailures returns current consecutive failure count
func (b *Backend) GetConsecutiveFailures() int
```

### 3. Modify proxy handler
**File**: `internal/proxy/handler.go`

After each connection attempt:
```go
backendConn, err := backend.Dial()
if err != nil {
    backend.RecordFailure()  // ← Add this
    return err
}
backend.RecordSuccess()  // ← Add this
```

### 4. Add sliding window logic
- Keep last N outcomes (default: 10)
- Efficiently calculate success rate
- Thread-safe updates

## Implementation Notes

- Use `sync.RWMutex` for concurrent access
- Sliding window as circular buffer for efficiency
- Timestamp last attempt for staleness detection
- Expose metrics via getter methods (for dashboard)

## Test Cases

1. **Successful connection**:
   - Success counter increments
   - Consecutive failures reset to 0
   - Sliding window updated

2. **Failed connection**:
   - Failure counter increments
   - Consecutive failures increments
   - Sliding window updated

3. **Success rate calculation**:
   - 10 successes → 100%
   - 7 success, 3 fail → 70%
   - Empty window → 0%

4. **Thread safety**:
   - Concurrent RecordSuccess/RecordFailure calls
   - No data races (verify with `-race` flag)

5. **Window overflow**:
   - Window size = 10
   - After 15 attempts, only last 10 counted

## Acceptance Criteria

- [ ] ConnectionMetrics struct implemented
- [ ] RecordSuccess/RecordFailure methods work correctly
- [ ] Sliding window maintains last N outcomes
- [ ] Success rate calculated accurately
- [ ] Proxy handler reports all outcomes
- [ ] All tests pass
- [ ] No data races detected
