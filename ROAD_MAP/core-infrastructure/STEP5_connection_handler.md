# STEP5: Connection Handler

## Goal
Complete basic routing from clients to backends without protocol decoding.

## Tasks
1. Update `handler.go` to select backend from pool
2. Establish connection to selected backend
3. Implement bidirectional TCP copy (io.Copy in goroutines)
4. Add connection timeout handling
5. Log connection events
6. Update `main.go` to wire everything together
7. Write end-to-end tests

## Files to Update
- `internal/proxy/handler.go`
- `cmd/socksbalance/main.go`

## Test Cases
- Client request reaches backend
- Backend response reaches client
- Connection closes properly on errors
- Multiple concurrent connections work
