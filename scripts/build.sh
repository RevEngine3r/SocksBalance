#!/bin/bash
# SocksBalance Cross-Platform Build Script (Linux/macOS)

APP_NAME="socksbalance"
MAIN_PATH="cmd/socksbalance/main.go"
OUT_DIR="bin"

# Supported OS/Arch combinations
platforms=(
    "linux/386"
    "linux/amd64"
    "linux/arm"
    "linux/arm64"
    "linux/mips"
    "linux/mipsle"
    "linux/mips64"
    "linux/mips64le"
    "linux/riscv64"
    "windows/386"
    "windows/amd64"
    "windows/arm"
    "windows/arm64"
)

echo "🚀 Starting build for $APP_NAME..."
mkdir -p $OUT_DIR

for platform in "${platforms[@]}"; do
    # Split OS and Arch
    IFS="/" read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"
    
    # Handle extension for Windows
    EXT=""
    if [ "$GOOS" == "windows" ]; then
        EXT=".exe"
    fi
    
    OUTPUT_NAME="${OUT_DIR}/${APP_NAME}-${GOOS}-${GOARCH}${EXT}"
    
    echo "📦 Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT_NAME" "$MAIN_PATH"
    
    if [ $? -ne 0 ]; then
        echo "❌ Failed to build for $GOOS/$GOARCH"
    fi
done

echo "✅ All builds completed! Binaries are in the '$OUT_DIR' directory."
