# SocksBalance Troubleshooting Guide

Common issues and solutions for SocksBalance.

## Connection Issues

### "No healthy backends available"

**Problem**: All backends are marked as unhealthy.

**Solutions**:
1. Check backend connectivity:
   ```bash
   curl -x socks5://proxy1.example.com:1080 https://www.google.com
   ```
2. Verify backend addresses in `config.yaml`
3. Check firewall rules allowing outbound connections
4. Review health check configuration:
   - Increase `connect_timeout` if backends are slow
   - Verify `test_url` is accessible from your network
   - Check if backends support HTTP through SOCKS5

**Debug**:
```bash
# Check logs for health check failures
grep "health check" socksbalance.log

# Test backend manually
telnet proxy1.example.com 1080
```

### "Connection timeout" errors

**Problem**: Connections to SocksBalance timeout.

**Solutions**:
1. Verify SocksBalance is running:
   ```bash
   ps aux | grep socksbalance
   netstat -tulpn | grep 1080
   ```
2. Check listen address in config:
   - Use `0.0.0.0:1080` for all interfaces
   - Use `127.0.0.1:1080` for localhost only
3. Verify firewall allows incoming connections:
   ```bash
   sudo iptables -L -n | grep 1080
   ```
4. Test connection:
   ```bash
   nc -zv localhost 1080
   ```

### "SOCKS5 handshake failed"

**Problem**: Client can't complete SOCKS5 negotiation.

**Solutions**:
1. Ensure client supports SOCKS5 (not SOCKS4)
2. Verify no authentication is required (SocksBalance doesn't support auth yet)
3. Check client SOCKS5 configuration
4. Test with curl:
   ```bash
   curl -v -x socks5://localhost:1080 https://ifconfig.me
   ```

## Performance Issues

### Slow connections through proxy

**Problem**: High latency when using SocksBalance.

**Solutions**:
1. Check backend latencies in logs
2. Verify `sort_by_latency: true` in config
3. Reduce `check_interval` for more frequent latency updates:
   ```yaml
   health:
     check_interval: 5s  # From default 10s
   ```
4. Use geographically closer backends
5. Test backend latency directly:
   ```bash
   time curl -x socks5://backend:1080 https://www.google.com
   ```

### High CPU usage

**Problem**: SocksBalance consumes excessive CPU.

**Solutions**:
1. Increase `check_interval` to reduce health check frequency:
   ```yaml
   health:
     check_interval: 30s  # From default 10s
   ```
2. Reduce number of concurrent connections
3. Check for connection leaks in logs
4. Monitor with:
   ```bash
   top -p $(pgrep socksbalance)
   ```

### Memory leaks

**Problem**: Memory usage grows over time.

**Solutions**:
1. Update to latest version (may contain fixes)
2. Monitor goroutines:
   ```bash
   # Add pprof endpoint (future feature)
   # curl http://localhost:6060/debug/pprof/goroutine?debug=1
   ```
3. Report issue with logs and memory profile

## Configuration Issues

### "Failed to load configuration"

**Problem**: SocksBalance can't parse config file.

**Solutions**:
1. Validate YAML syntax:
   ```bash
   yamllint config.yaml
   ```
2. Check file permissions:
   ```bash
   ls -l config.yaml
   chmod 644 config.yaml
   ```
3. Verify required fields are present:
   ```yaml
   listen: "0.0.0.0:1080"  # Required
   backends:                # At least one required
     - address: "..."
   ```
4. Review error message for specific field issues

### Backends not being used

**Problem**: Traffic only goes to some backends.

**Solutions**:
1. Verify all backends are healthy (check logs)
2. Ensure round-robin is working (check logs for distribution)
3. Check if some backends have very high latency
4. Verify all backend addresses are unique
5. Review balancer configuration:
   ```yaml
   balancer:
     algorithm: "roundrobin"
     sort_by_latency: true
   ```

## Health Check Issues

### Backends marked unhealthy incorrectly

**Problem**: Healthy backends show as unhealthy.

**Solutions**:
1. Increase timeouts:
   ```yaml
   health:
     connect_timeout: 10s   # From default 5s
     request_timeout: 20s   # From default 10s
   ```
2. Verify `test_url` is accessible:
   ```bash
   curl -x socks5://backend:1080 https://www.google.com
   ```
3. Check if backends support HTTP CONNECT
4. Reduce `failure_threshold`:
   ```yaml
   health:
     failure_threshold: 5  # From default 3
   ```

### Health checks too frequent

**Problem**: Health checks consume too much bandwidth.

**Solutions**:
1. Increase `check_interval`:
   ```yaml
   health:
     check_interval: 30s  # From default 10s
   ```
2. Use lighter `test_url`:
   ```yaml
   health:
     test_url: "http://captive.apple.com/hotspot-detect.html"
   ```

## Logging and Debugging

### Enable debug logging

```yaml
log:
  level: "debug"  # From "info"
  format: "text"
```

### Common log patterns

**Backend failure**:
```
[ERROR] Failed to connect to backend 192.168.1.100:1080: connection refused
[WARN] Backend 192.168.1.100:1080 marked unhealthy
```

**Health check failure**:
```
[ERROR] Health check failed for 192.168.1.100:1080: timeout
[INFO] Backend 192.168.1.100:1080 failure count: 1/3
```

**Successful routing**:
```
[INFO] New SOCKS5 connection from 192.168.1.50:54321
[INFO] Routing 192.168.1.50:54321 through backend 192.168.1.100:1080 to example.com:443
[INFO] Backend handshake successful, relaying data
```

### Collecting diagnostics

When reporting issues, include:

1. **Configuration**:
   ```bash
   cat config.yaml
   ```

2. **Version**:
   ```bash
   socksbalance -version
   ```

3. **Logs** (last 100 lines):
   ```bash
   tail -n 100 socksbalance.log
   ```

4. **System info**:
   ```bash
   uname -a
   go version
   ```

5. **Network test**:
   ```bash
   # Test backend directly
   curl -x socks5://backend:1080 https://ifconfig.me
   
   # Test through SocksBalance
   curl -x socks5://localhost:1080 https://ifconfig.me
   ```

## Getting Help

If issues persist:

1. Check [Issues](https://github.com/RevEngine3r/SocksBalance/issues) for similar problems
2. Open new issue with:
   - SocksBalance version
   - Configuration file (remove sensitive data)
   - Relevant log output
   - Steps to reproduce
3. Join discussions in [Discussions](https://github.com/RevEngine3r/SocksBalance/discussions)

## Known Issues

### Authentication not supported

SocksBalance currently only supports NO_AUTH SOCKS5. Backend servers requiring username/password authentication are not supported.

**Workaround**: Use backends without authentication or use pre-authenticated tunnels.

### IPv6 support

IPv6 addresses are parsed but may not work correctly in all scenarios.

**Workaround**: Use IPv4 addresses for backends and listen address.

### Hot reload not implemented

Changing configuration requires restart.

**Workaround**: Plan maintenance windows for config updates.
