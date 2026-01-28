# STEP1: GitHub Actions Workflow

## Goal
Create workflow that builds cross-platform binaries and creates releases on tag push.

## Tasks
1. Create `.github/workflows/release.yml`
2. Trigger on tags matching `release` or `release-*`
3. Use Go 1.22+ for builds
4. Build matrix for multiple OS/arch combinations
5. Generate archives (tar.gz for Unix, zip for Windows)
6. Extract version from git tag
7. Create GitHub Release with artifacts
8. Add checksums file (SHA256)

## Files to Create
- `.github/workflows/release.yml`

## Build Targets
- linux/amd64
- linux/arm64
- linux/arm (armv7)
- darwin/amd64
- darwin/arm64
- windows/amd64

## Test Plan
- Push test tag and verify builds
- Download artifacts and test execution
- Verify checksums match
