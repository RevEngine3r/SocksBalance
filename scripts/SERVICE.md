# SocksBalance Systemd Service Deployment

Complete guide for deploying SocksBalance as a systemd service on Linux.

## Quick Installation

### Automated Installation (Recommended)

```bash
# 1. Build the binary
go build -o socksbalance ./cmd/socksbalance

# 2. Run the installation script
sudo ./scripts/install-service.sh

# 3. Edit configuration
sudo nano /etc/socksbalance/config.yaml

# 4. Start the service
sudo systemctl enable socksbalance
sudo systemctl start socksbalance

# 5. Check status
sudo systemctl status socksbalance
```

---

## Manual Installation

If you prefer manual installation or the automated script doesn't work:

### 1. Create System User

```bash
sudo useradd --system --no-create-home --shell /bin/false socksbalance
```

### 2. Create Directories

```bash
sudo mkdir -p /opt/socksbalance
sudo mkdir -p /etc/socksbalance
sudo mkdir -p /var/log/socksbalance
sudo chown -R socksbalance:socksbalance /opt/socksbalance
sudo chown -R socksbalance:socksbalance /var/log/socksbalance
```

### 3. Install Binary

```bash
# Build the binary
go build -o socksbalance ./cmd/socksbalance

# Copy to installation directory
sudo cp socksbalance /opt/socksbalance/
sudo chmod +x /opt/socksbalance/socksbalance
sudo chown socksbalance:socksbalance /opt/socksbalance/socksbalance
```

### 4. Install Configuration

```bash
sudo cp config.example.yaml /etc/socksbalance/config.yaml
sudo chown socksbalance:socksbalance /etc/socksbalance/config.yaml
sudo chmod 640 /etc/socksbalance/config.yaml

# Edit configuration
sudo nano /etc/socksbalance/config.yaml
```

### 5. Install Service File

```bash
sudo cp scripts/socksbalance.service /etc/systemd/system/
sudo systemctl daemon-reload
```

### 6. Enable and Start Service

```bash
# Enable service to start on boot
sudo systemctl enable socksbalance

# Start the service now
sudo systemctl start socksbalance

# Check status
sudo systemctl status socksbalance
```

---

## Service Management

### Basic Commands

```bash
# Start service
sudo systemctl start socksbalance

# Stop service
sudo systemctl stop socksbalance

# Restart service
sudo systemctl restart socksbalance

# Reload configuration (if supported)
sudo systemctl reload socksbalance

# Check status
sudo systemctl status socksbalance

# Enable auto-start on boot
sudo systemctl enable socksbalance

# Disable auto-start on boot
sudo systemctl disable socksbalance
```

### View Logs

```bash
# View all logs
sudo journalctl -u socksbalance

# Follow logs in real-time
sudo journalctl -u socksbalance -f

# View logs since boot
sudo journalctl -u socksbalance -b

# View last 100 lines
sudo journalctl -u socksbalance -n 100

# View logs from last hour
sudo journalctl -u socksbalance --since "1 hour ago"

# View logs with timestamps
sudo journalctl -u socksbalance -o short-precise
```

### Check Service Health

```bash
# Check if service is running
systemctl is-active socksbalance

# Check if service is enabled
systemctl is-enabled socksbalance

# Check if service failed
systemctl is-failed socksbalance

# View service dependencies
systemctl list-dependencies socksbalance
```

---

## Configuration

### File Locations

- **Binary**: `/opt/socksbalance/socksbalance`
- **Configuration**: `/etc/socksbalance/config.yaml`
- **Service file**: `/etc/systemd/system/socksbalance.service`
- **Logs**: View via `journalctl` or optionally `/var/log/socksbalance/`

### Update Configuration

```bash
# Edit config
sudo nano /etc/socksbalance/config.yaml

# Restart service to apply changes
sudo systemctl restart socksbalance

# Verify configuration is working
sudo systemctl status socksbalance
```

### Important Configuration Notes

1. **Listen Address**: By default `0.0.0.0:1080` - accessible from network
2. **Web Dashboard**: Default `127.0.0.1:8080` - localhost only for security
3. **Backend Servers**: Must be configured before starting service
4. **Circuit Breaker**: Enabled by default with 3 failure threshold

---

## Updating the Service

```bash
# 1. Stop the service
sudo systemctl stop socksbalance

# 2. Build new binary
go build -o socksbalance ./cmd/socksbalance

# 3. Replace binary
sudo cp socksbalance /opt/socksbalance/
sudo chmod +x /opt/socksbalance/socksbalance
sudo chown socksbalance:socksbalance /opt/socksbalance/socksbalance

# 4. Update service file if needed
sudo cp scripts/socksbalance.service /etc/systemd/system/
sudo systemctl daemon-reload

# 5. Start the service
sudo systemctl start socksbalance

# 6. Check logs for any issues
sudo journalctl -u socksbalance -f
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check detailed status
sudo systemctl status socksbalance -l

# Check logs
sudo journalctl -u socksbalance -n 50

# Check if binary is executable
ls -la /opt/socksbalance/socksbalance

# Check if config file exists and is readable
ls -la /etc/socksbalance/config.yaml

# Try running manually to see errors
sudo -u socksbalance /opt/socksbalance/socksbalance -config /etc/socksbalance/config.yaml
```

### Permission Issues

```bash
# Fix ownership
sudo chown -R socksbalance:socksbalance /opt/socksbalance
sudo chown -R socksbalance:socksbalance /var/log/socksbalance
sudo chown socksbalance:socksbalance /etc/socksbalance/config.yaml

# Fix permissions
sudo chmod 755 /opt/socksbalance
sudo chmod 755 /opt/socksbalance/socksbalance
sudo chmod 640 /etc/socksbalance/config.yaml
```

### Port Already in Use

```bash
# Check what's using port 1080
sudo netstat -tlnp | grep 1080
# or
sudo lsof -i :1080

# Change port in config if needed
sudo nano /etc/socksbalance/config.yaml
```

### Service Crashes/Restarts

```bash
# Check crash logs
sudo journalctl -u socksbalance -p err

# Check if it's restarting
sudo systemctl status socksbalance

# Disable auto-restart temporarily to debug
sudo systemctl edit socksbalance
# Add: [Service]
#      Restart=no

# Or run in foreground mode
sudo -u socksbalance /opt/socksbalance/socksbalance -config /etc/socksbalance/config.yaml
```

### Backend Connection Issues

```bash
# Test backend connectivity
curl -x socks5://127.0.0.1:1080 https://www.google.com

# Check backend health in logs
sudo journalctl -u socksbalance -f | grep -i "backend\|health\|circuit"

# Access web dashboard (if enabled)
curl http://127.0.0.1:8080/api/stats
```

---

## Uninstallation

### Automated Uninstallation (Recommended)

```bash
sudo ./scripts/uninstall-service.sh
```

This will:
- Stop and disable the service
- Remove the service file
- Remove the binary
- Optionally remove config, logs, and user

### Manual Uninstallation

```bash
# 1. Stop and disable service
sudo systemctl stop socksbalance
sudo systemctl disable socksbalance

# 2. Remove service file
sudo rm /etc/systemd/system/socksbalance.service
sudo systemctl daemon-reload

# 3. Remove files
sudo rm -rf /opt/socksbalance
sudo rm -rf /etc/socksbalance  # Optional: keeps config
sudo rm -rf /var/log/socksbalance  # Optional: keeps logs

# 4. Remove user
sudo userdel socksbalance  # Optional
```

---

## Security Hardening

The service file includes several security features:

- **Non-root user**: Runs as `socksbalance` user
- **NoNewPrivileges**: Prevents privilege escalation
- **PrivateTmp**: Isolated /tmp directory
- **SystemCallFilter**: Restricts system calls
- **Resource limits**: LimitNOFILE=65536, LimitNPROC=512

### Additional Security (Optional)

Edit `/etc/systemd/system/socksbalance.service` and add:

```ini
[Service]
# Read-only root filesystem
ReadOnlyPaths=/
ReadWritePaths=/var/log/socksbalance

# Restrict network access
RestrictAddressFamilies=AF_INET AF_INET6

# No access to home directories
ProtectHome=true

# No access to kernel logs
ProtectKernelLogs=true
```

Then reload:
```bash
sudo systemctl daemon-reload
sudo systemctl restart socksbalance
```

---

## Monitoring

### Health Checks

```bash
# Check if service is running
systemctl is-active socksbalance

# Test SOCKS5 connection
curl -x socks5://127.0.0.1:1080 https://ifconfig.me

# Check web dashboard
curl http://127.0.0.1:8080/api/stats | jq
```

### Performance Monitoring

```bash
# CPU and memory usage
systemctl status socksbalance

# Detailed resource usage
sudo systemd-cgtop /system.slice/socksbalance.service

# Network connections
sudo netstat -anp | grep socksbalance
```

### Alerts

Create a monitoring script:

```bash
#!/bin/bash
# /usr/local/bin/check-socksbalance.sh

if ! systemctl is-active --quiet socksbalance; then
    echo "SocksBalance service is DOWN!"
    # Send alert (email, Slack, etc.)
fi
```

Add to crontab:
```bash
*/5 * * * * /usr/local/bin/check-socksbalance.sh
```

---

## Advanced Configuration

### Multiple Instances

Run multiple instances on different ports:

```bash
# Copy service file
sudo cp /etc/systemd/system/socksbalance.service \
        /etc/systemd/system/socksbalance-alt.service

# Edit the new service file
sudo nano /etc/systemd/system/socksbalance-alt.service
# Change: ExecStart=/opt/socksbalance/socksbalance -config /etc/socksbalance/config-alt.yaml

# Create alternate config
sudo cp /etc/socksbalance/config.yaml /etc/socksbalance/config-alt.yaml
sudo nano /etc/socksbalance/config-alt.yaml
# Change listen port to 1081

# Start second instance
sudo systemctl start socksbalance-alt
sudo systemctl enable socksbalance-alt
```

### Systemd Socket Activation (Advanced)

For on-demand service activation - see systemd.socket documentation.

---

## Support

For issues:
- Check [TROUBLESHOOTING.md](../TROUBLESHOOTING.md)
- View logs: `sudo journalctl -u socksbalance -f`
- GitHub: https://github.com/RevEngine3r/SocksBalance/issues
