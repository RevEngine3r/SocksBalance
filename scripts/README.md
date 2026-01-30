# SocksBalance Scripts

Utility scripts for building, deploying, and managing SocksBalance.

## Build Scripts

### `build.sh` (Linux/macOS)
Cross-compile binaries for multiple platforms.

```bash
chmod +x build.sh
./build.sh
```

Output: `bin/socksbalance-{os}-{arch}`

### `build.bat` (Windows)
Cross-compile binaries on Windows.

```cmd
build.bat
```

### `release.sh` / `release.bat`
Build release binaries with version info and optimizations.

```bash
./release.sh v1.0.0
```

---

## Service Management (Linux)

### `socksbalance.service`
Systemd service file for running SocksBalance as a Linux service.

**Features:**
- Automatic restart on failure
- Security hardening (non-root user, restricted syscalls)
- Resource limits (65536 file descriptors, 512 processes)
- Logging to journald

### `install-service.sh`
Automated installation script for systemd service.

**What it does:**
1. Creates `socksbalance` user and group
2. Creates directories: `/opt/socksbalance`, `/etc/socksbalance`, `/var/log/socksbalance`
3. Installs binary to `/opt/socksbalance/socksbalance`
4. Copies config to `/etc/socksbalance/config.yaml`
5. Installs systemd service file
6. Reloads systemd

**Usage:**
```bash
chmod +x install-service.sh
sudo ./install-service.sh
```

**After installation:**
```bash
# Edit config
sudo nano /etc/socksbalance/config.yaml

# Start service
sudo systemctl enable socksbalance
sudo systemctl start socksbalance

# Check status
sudo systemctl status socksbalance

# View logs
sudo journalctl -u socksbalance -f
```

### `uninstall-service.sh`
Removes SocksBalance systemd service.

**What it does:**
1. Stops and disables service
2. Removes service file
3. Removes binary directory
4. Optionally removes config and logs
5. Optionally removes system user

**Usage:**
```bash
chmod +x uninstall-service.sh
sudo ./uninstall-service.sh
```

---

## Documentation

### `SERVICE.md`
Comprehensive guide for systemd service deployment and management.

**Topics covered:**
- Installation (automated and manual)
- Service management commands
- Configuration updates
- Troubleshooting
- Security hardening
- Monitoring and alerts
- Multiple instances
- Uninstallation

**View:** [SERVICE.md](SERVICE.md)

---

## Quick Reference

### Build for Production
```bash
# Linux/macOS
./scripts/build.sh

# Windows
scripts\build.bat

# Binaries in: bin/
```

### Deploy as Service (Linux)
```bash
# 1. Build
go build -o socksbalance ./cmd/socksbalance

# 2. Install
sudo ./scripts/install-service.sh

# 3. Configure
sudo nano /etc/socksbalance/config.yaml

# 4. Start
sudo systemctl enable socksbalance
sudo systemctl start socksbalance

# 5. Monitor
sudo journalctl -u socksbalance -f
```

### Common Tasks
```bash
# Check service status
sudo systemctl status socksbalance

# Restart after config change
sudo systemctl restart socksbalance

# View live logs
sudo journalctl -u socksbalance -f

# View errors only
sudo journalctl -u socksbalance -p err

# Update binary
sudo systemctl stop socksbalance
sudo cp socksbalance /opt/socksbalance/
sudo systemctl start socksbalance
```

### Troubleshooting
```bash
# Check if service is running
systemctl is-active socksbalance

# View recent logs
sudo journalctl -u socksbalance -n 50

# Test manually (foreground)
sudo -u socksbalance /opt/socksbalance/socksbalance -config /etc/socksbalance/config.yaml

# Check ports
sudo netstat -tlnp | grep socksbalance
sudo lsof -i :1080

# Check permissions
ls -la /opt/socksbalance/
ls -la /etc/socksbalance/config.yaml
```

---

## File Locations (Service Installation)

| Item | Location |
|------|----------|
| Binary | `/opt/socksbalance/socksbalance` |
| Configuration | `/etc/socksbalance/config.yaml` |
| Service File | `/etc/systemd/system/socksbalance.service` |
| Logs | `journalctl -u socksbalance` |
| User/Group | `socksbalance:socksbalance` |

---

## Security Notes

The systemd service runs with security hardening:

- **Non-root user**: Runs as `socksbalance` user
- **NoNewPrivileges**: Cannot escalate privileges
- **PrivateTmp**: Isolated temporary directory
- **SystemCallFilter**: Restricted system calls
- **Resource limits**: Controlled file descriptors and processes

For additional hardening options, see [SERVICE.md](SERVICE.md#security-hardening).

---

## Support

- **Full Documentation**: [SERVICE.md](SERVICE.md)
- **Configuration Guide**: [../config.example.yaml](../config.example.yaml)
- **Troubleshooting**: [../TROUBLESHOOTING.md](../TROUBLESHOOTING.md)
- **Issues**: https://github.com/RevEngine3r/SocksBalance/issues
