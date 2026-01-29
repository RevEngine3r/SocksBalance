... (existing content) ...

## Build from Source

SocksBalance supports easy cross-platform compilation for various operating systems and architectures.

### Prerequisites
- Go 1.22 or higher installed.

### Using Build Scripts
The project includes automated scripts to build binaries for Linux and Windows (Intel/ARM, 32/64-bit).

**On Linux/macOS:**
```bash
chmod +x scripts/build.sh
./scripts/build.sh
```

**On Windows:**
```cmd
scripts\build.bat
```

Binaries will be generated in the `bin/` directory with the following naming convention:
`socksbalance-[os]-[arch][.exe]`

### Manual Build
To build for your current platform:
```bash
go build -o socksbalance cmd/socksbalance/main.go
```

... (rest of the file)
