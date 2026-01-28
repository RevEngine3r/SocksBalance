# STEP3: Backend Representation

## Goal
Create backend server structs and pool manager for tracking SOCKS5 servers.

## Tasks
1. Create `internal/backend/backend.go` with `Backend` struct
2. Add fields: address, name, healthy status, latency, failure count
3. Implement thread-safe getters/setters
4. Create `internal/backend/pool.go` for managing backend list
5. Implement methods: Add, Remove, GetHealthy, GetAll, UpdateLatency
6. Write unit tests

## Files to Create
- `internal/backend/backend.go`
- `internal/backend/pool.go`
- `internal/backend/backend_test.go`
- `internal/backend/pool_test.go`

## Test Cases
- Backend status updates thread-safely
- Pool returns only healthy backends
- Latency updates reflected immediately
