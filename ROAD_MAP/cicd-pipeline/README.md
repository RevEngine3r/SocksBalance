# Feature: CI/CD Pipeline

## Overview
Automated build and release using GitHub Actions.

## Steps
1. **STEP1**: GitHub Actions workflow for releases
2. **STEP2**: Cross-platform builds (Linux, macOS, Windows)
3. **STEP3**: ARM support (arm64, armv7)
4. **STEP4**: Release artifact generation
5. **STEP5**: Version tagging automation

## Acceptance Criteria
- [ ] Workflow triggers on `release` tag
- [ ] Builds for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [ ] Archives created with binaries and config example
- [ ] GitHub Release created automatically
- [ ] Release notes generated from commits
