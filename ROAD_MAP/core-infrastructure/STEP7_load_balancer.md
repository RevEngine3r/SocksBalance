# STEP7: Load Balancer

## Objective
Implement intelligent load balancing to distribute connections across healthy backends using round-robin with latency-based optimization.

## Tasks
1. Create `internal/balancer/balancer.go` with round-robin algorithm
2. Integrate latency-based backend sorting from pool
3. Thread-safe backend selection with atomic counter
4. Replace hardcoded `backends[0]` in server with balancer
5. Comprehensive unit tests for selection logic

## Components
- **Balancer**: Round-robin selector with latency awareness
- **Thread Safety**: Atomic operations for concurrent access
- **Integration**: Seamless server integration

## Deliverables
- `internal/balancer/balancer.go` - Load balancer implementation
- `internal/balancer/balancer_test.go` - Unit tests (8+ test cases)
- Updated `internal/proxy/server.go` - Use balancer for backend selection
- Updated `cmd/socksbalance/main.go` - Initialize balancer

## Success Criteria
- Round-robin distributes requests evenly
- Latency-sorted backends preferred
- Thread-safe concurrent access
- All tests pass
