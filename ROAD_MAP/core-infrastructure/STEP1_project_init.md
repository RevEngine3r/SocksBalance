# STEP1: Project Initialization

## Goal
Initialize Go module, define project structure, and create basic main.go entry point.

## Tasks
1. Create `go.mod` with module name `github.com/RevEngine3r/SocksBalance`
2. Create directory structure: `cmd/socksbalance/`, `internal/{config,proxy,backend,health,balancer}/`
3. Create `cmd/socksbalance/main.go` with basic CLI flag parsing
4. Add example config file `config.example.yaml`
5. Create README.md with project description and usage

## Files to Create
- `go.mod`
- `cmd/socksbalance/main.go`
- `config.example.yaml`
- `README.md`
- `LICENSE` (MIT)

## Dependencies
- `gopkg.in/yaml.v3` for config parsing

## Test Plan
- Verify `go mod tidy` runs successfully
- Verify `go build ./cmd/socksbalance` compiles
