# Feature: Cross-Platform Build System

## Overview
Provide automated build scripts for Windows and Linux across all major architectures (Intel/ARM, 32/64-bit) to simplify distribution.

## Roadmap
- [ ] **STEP 1: Cross-Compilation Scripts**
  - Create `scripts/build.sh` for Linux/macOS users.
  - Create `scripts/build.bat` for Windows users.
  - Support: Linux (386, amd64, arm, arm64, mips, riscv64), Windows (386, amd64, arm, arm64).
- [ ] **STEP 2: Build Documentation**
  - Update `README.md` with instructions on how to use the scripts.
  - Update `TROUBLESHOOTING.md` for common build issues.
