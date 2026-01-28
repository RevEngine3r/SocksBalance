# Feature: Core Infrastructure

## Overview
Bootstrap the Go project with configuration management, backend representation, and TCP proxy server foundation.

## Steps
1. **STEP1**: Project initialization (go.mod, basic structure)
2. **STEP2**: Configuration system (YAML loader, validation)
3. **STEP3**: Backend representation (server struct, pool manager)
4. **STEP4**: TCP proxy server (listener, connection acceptor)
5. **STEP5**: Connection handler (basic routing to backends)

## Acceptance Criteria
- [ ] Go module initialized with dependencies
- [ ] Configuration loads from YAML
- [ ] Backend pool manages server list
- [ ] TCP server accepts connections
- [ ] Connections route to backends (no health checks yet)
