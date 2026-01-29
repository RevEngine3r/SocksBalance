@echo off
setlocal enabledelayedexpansion

:: SocksBalance Cross-Platform Build Script (Windows)

set APP_NAME=socksbalance
set MAIN_PATH=cmd/socksbalance/main.go
set OUT_DIR=bin

echo 🚀 Starting build for %APP_NAME%...
if not exist %OUT_DIR% mkdir %OUT_DIR%

:: Define platforms (OS/Arch)
:: Note: Windows batch doesn't have arrays, so we use a space-separated string
set platforms=linux/386 linux/amd64 linux/arm linux/arm64 linux/mips linux/mipsle linux/mips64 linux/mips64le linux/riscv64 windows/386 windows/amd64 windows/arm windows/arm64

for %%p in (%platforms%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        set GOOS=%%a
        set GOARCH=%%b
        
        set EXT=
        if "!GOOS!"=="windows" set EXT=.exe
        
        set OUTPUT_NAME=%OUT_DIR%\%APP_NAME%-!GOOS!-!GOARCH!!EXT!
        
        echo 📦 Building for !GOOS!/!GOARCH!...
        set GOOS=!GOOS!
        set GOARCH=!GOARCH!
        go build -ldflags="-s -w" -o "!OUTPUT_NAME!" "%MAIN_PATH%"
        
        if errorlevel 1 (
            echo ❌ Failed to build for !GOOS!/!GOARCH!
        )
    )
)

echo ✅ All builds completed! Binaries are in the '%OUT_DIR%' directory.
endlocal
