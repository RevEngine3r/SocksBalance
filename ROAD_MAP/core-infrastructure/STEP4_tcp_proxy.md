# STEP4: TCP Proxy Server

## Goal
Implement TCP listener that accepts client connections and basic routing.

## Tasks
1. Create `internal/proxy/server.go` with TCP listener
2. Implement `Start(address string)` method
3. Accept connections in goroutines
4. Create `internal/proxy/handler.go` for connection handling
5. Implement basic TCP forwarding (client ↔ backend)
6. Add graceful shutdown support
7. Write integration tests

## Files to Create
- `internal/proxy/server.go`
- `internal/proxy/handler.go`
- `internal/proxy/proxy_test.go`

## Test Cases
- Server listens on configured port
- Connections accepted and forwarded
- Bidirectional data flow works
- Graceful shutdown closes connections
