# STEP1: HTTP Server Foundation

## Objective
Create a lightweight HTTP server that will host the web dashboard and API endpoints.

## Files to Create
- `internal/web/server.go`
- `internal/web/server_test.go`

## Implementation Details

### Server Struct
```go
type Server struct {
    addr     string
    pool     *backend.Pool
    httpSrv  *http.Server
    mu       sync.Mutex
}
```

### Methods
1. **NewServer(addr string, pool *backend.Pool) *Server**
   - Initialize server with address and backend pool
   - Create http.Server instance

2. **Start(ctx context.Context) error**
   - Register HTTP handlers (routes)
   - Start HTTP server in goroutine
   - Return immediately (non-blocking)

3. **Stop() error**
   - Graceful shutdown with timeout
   - Close all connections

### Routes (Initial)
- `/` - Will serve dashboard (STEP3)
- `/api/stats` - Will serve JSON stats (STEP2)
- `/health` - Simple health check endpoint

## Testing
- Test server start/stop lifecycle
- Test health endpoint returns 200 OK
- Test multiple start calls are idempotent
- Test graceful shutdown

## Pseudocode
```
func NewServer(addr, pool):
    server = Server{
        addr: addr,
        pool: pool,
        httpSrv: nil,
    }
    return server

func (s *Server) Start(ctx):
    mux = http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    
    s.httpSrv = &http.Server{
        Addr: s.addr,
        Handler: mux,
    }
    
    go s.httpSrv.ListenAndServe()
    return nil

func (s *Server) Stop():
    if s.httpSrv == nil:
        return nil
    
    ctx = context.WithTimeout(5s)
    err = s.httpSrv.Shutdown(ctx)
    return err
```

## Acceptance Criteria
- ✅ Server starts on specified address
- ✅ `/health` returns 200 OK with `{"status":"ok"}`
- ✅ Server stops gracefully
- ✅ All unit tests pass
- ✅ No goroutine leaks
