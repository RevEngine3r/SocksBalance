# STEP2: Configuration System

## Goal
Implement YAML configuration loading with validation and type-safe structs.

## Tasks
1. Create `internal/config/config.go` with configuration structs
2. Implement `Load(path string)` function using `gopkg.in/yaml.v3`
3. Add validation for required fields and value ranges
4. Implement default values for optional fields
5. Write unit tests for valid and invalid configs

## Files to Create
- `internal/config/config.go`
- `internal/config/config_test.go`

## Test Cases
- Valid configuration loads successfully
- Missing required fields return errors
- Invalid durations return errors
- Default values applied correctly
