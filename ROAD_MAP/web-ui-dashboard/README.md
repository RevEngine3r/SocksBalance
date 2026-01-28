# Web UI Dashboard

## Overview
Real-time web dashboard to monitor SOCKS5 backend servers with health status, latencies, and automatic AJAX updates.

## Goal
Provide a modern, responsive web interface that displays:
- All backend servers in a sorted table
- Real-time health status (healthy/unhealthy)
- Current latency measurements
- Auto-refresh using AJAX (no page reload)
- Clean, professional UI with modern design

## Architecture
```
Browser
  ↓
  HTTP GET /          → Serves index.html (dashboard UI)
  HTTP GET /api/stats → Returns JSON with backends status
  ↓
HTTP Server (internal/web/)
  ↓
Backend Pool (existing)
```

## Technology Stack
- **Backend**: Go net/http (built-in HTTP server)
- **Frontend**: Vanilla JavaScript (no frameworks)
- **Styling**: Modern CSS with Tailwind-inspired utility classes
- **Data Format**: JSON API
- **Updates**: AJAX polling (configurable interval)

## Features
1. **HTTP Server**: Lightweight HTTP server on configurable port
2. **JSON API Endpoint**: `/api/stats` returns backend status
3. **Dashboard UI**: Single-page HTML with embedded CSS/JS
4. **Real-time Updates**: Auto-refresh every 2 seconds via AJAX
5. **Sorted Display**: Backends sorted by latency (fastest first)
6. **Health Indicators**: Visual status badges (✅ healthy, ❌ unhealthy)
7. **Responsive Design**: Works on desktop and mobile
8. **Dark Mode**: Modern dark theme by default

## Steps

### STEP1: HTTP Server Foundation
**File**: `internal/web/server.go`
- Create HTTP server struct
- Accept backend pool reference
- Implement Start/Stop methods
- Listen on configurable address (e.g., `:8080`)
- Serve static content and API endpoint
- Unit tests for server lifecycle

### STEP2: JSON API Endpoint
**File**: `internal/web/handler.go`
- Implement `/api/stats` handler
- Fetch all backends from pool
- Sort by latency (ascending)
- Serialize to JSON format:
  ```json
  {
    "timestamp": "2026-01-28T21:00:00Z",
    "total_backends": 20,
    "healthy_backends": 18,
    "backends": [
      {
        "address": "127.0.0.1:9070",
        "name": "Tor-1",
        "healthy": true,
        "latency_ms": 45,
        "last_checked": "2026-01-28T20:59:55Z"
      }
    ]
  }
  ```
- Add CORS headers for local development
- Unit tests for JSON serialization

### STEP3: Dashboard HTML/CSS
**File**: `internal/web/static/index.html`
- Single-page HTML structure
- Header with title and stats summary
- Table with columns: Status, Name, Address, Latency, Last Check
- Modern CSS styling:
  - Dark theme (background: #1a1a2e, text: #eee)
  - Card-based layout with shadows
  - Color-coded latency (green < 100ms, yellow < 500ms, red >= 500ms)
  - Smooth transitions and hover effects
- Loading indicator for initial load
- Responsive design (mobile-friendly)

### STEP4: AJAX Auto-Update
**File**: `internal/web/static/index.html` (inline JavaScript)
- Fetch `/api/stats` on page load
- Parse JSON and populate table
- Sort backends by latency (frontend confirmation)
- Update table rows without page reload
- Auto-refresh every 2 seconds using `setInterval`
- Show "Last Updated" timestamp
- Handle errors gracefully (retry logic)

### STEP5: Integration & Configuration
**File**: `cmd/socksbalance/main.go`, `internal/config/config.go`
- Add web dashboard config to YAML:
  ```yaml
  web:
    enabled: true
    listen: "0.0.0.0:8080"
    refresh_interval: 2  # seconds
  ```
- Initialize web server in main.go
- Start web server in separate goroutine
- Add graceful shutdown for web server
- Update README with web UI usage
- Integration test: start server + verify API response

### STEP6: Polish & Documentation
- Add visual enhancements:
  - Backend count badges
  - Latency trend indicators (↑↓)
  - Active backend highlighting
- Update `config.example.yaml` with web section
- Add screenshots to README
- Document web dashboard in TROUBLESHOOTING.md
- Add security note (bind to localhost in production)

## Completion Criteria
- ✅ HTTP server serves dashboard on `:8080`
- ✅ `/api/stats` returns accurate JSON data
- ✅ Dashboard displays all backends sorted by latency
- ✅ AJAX updates table every 2 seconds
- ✅ Health status visually distinct (colors/icons)
- ✅ Responsive design works on mobile
- ✅ Configuration option to enable/disable web UI
- ✅ All unit tests pass (web package)
- ✅ Integration test verifies end-to-end flow
- ✅ Documentation updated

## Security Considerations
- Bind to `127.0.0.1` by default (localhost only)
- No authentication in v1 (add in future if needed)
- Read-only API (no POST/PUT/DELETE)
- CORS headers only for development

## Future Enhancements (Not in Scope)
- Historical latency graphs
- WebSocket for push updates
- Backend control panel (enable/disable backends)
- Authentication/authorization
- TLS support
