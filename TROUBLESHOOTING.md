# SocksBalance Troubleshooting Guide

## Common Issues

### 1. Twitter/Social Media Not Loading Images

**Symptoms**:
- Twitter feeds load but images don't appear
- Instagram posts visible but images broken
- Multi-request web apps fail

**Root Cause**:
Different backends used for HTML vs images. Social media uses cookies/session tokens that are IP-specific.

**Solution**:
Enable sticky sessions to keep same client on same backend:

```yaml
balancer:
  sticky_session_ttl: 15m  # Increase for longer sessions
```

**Verification**:
- Check web dashboard - same client should use same backend
- Clear browser cache and retry
- Try `sticky_session_ttl: 30m` for longer stability

---

### 2. All Backends Getting Blocked (GFW)

**Symptoms**:
- All connections suddenly fail
- All backends marked unhealthy at once
- Complete service outage

**Root Cause**:
GFW detected traffic pattern across all backends simultaneously.

**Solution**:
Limit concurrent backend exposure:

```yaml
balancer:
  max_active_backends: 3  # Only expose 3 backends at a time
```

**How it helps**:
- Only 3 backends used concurrently
- If GFW blocks 3, auto-switch to next 3
- Remaining backends stay as backup

**Verification**:
- Enable web dashboard to watch rotation
- Should see only top N backends active
- When backend fails, next fastest automatically used

---

### 3. Slow Connection Performance

**Symptoms**:
- High latency
- Timeouts
- Sluggish browsing

**Root Cause**:
Using slow or overloaded backends.

**Solution**:
Enable latency filtering:

```yaml
balancer:
  max_latency: 1000ms  # Only use backends faster than 1s
```

**Tuning**:
- Commercial proxies: `500ms-1000ms`
- Tor circuits: `2000ms-3000ms`
- Local SOCKS5: `100ms-500ms`

**Verification**:
- Check web dashboard for backend latencies
- Slow backends should be excluded
- Only green/yellow latencies should be active

---

### 4. Connection Refused Errors

**Symptoms**:
```
connection refused
dial tcp connect: connection refused
```

**Possible Causes**:

**A. Backend not running**
- Check backend SOCKS5 servers are started
- Verify ports are correct: `netstat -tuln | grep <port>`

**B. Wrong listen address**
```yaml
listen: "0.0.0.0:1080"  # Should match your client config
```

**C. Firewall blocking**
- Linux: `sudo iptables -L`
- Allow port: `sudo ufw allow 1080`

---

### 5. Health Checks Failing

**Symptoms**:
- All backends marked unhealthy
- Dashboard shows all red
- Logs show health check errors

**Solution A: Adjust timeouts**
```yaml
health:
  connect_timeout: 10s     # Increase for slow backends
  request_timeout: 15s     # Increase for slow networks
  failure_threshold: 5     # More forgiving
```

**Solution B: Change test URL**
```yaml
health:
  test_url: "https://httpbin.org/ip"  # Alternative test endpoint
```

**Solution C: Check backend connectivity**
```bash
# Test backend manually
curl -x socks5://127.0.0.1:9070 https://www.google.com
```

---

### 6. Web Dashboard Not Accessible

**Symptoms**:
- Cannot open http://127.0.0.1:8080
- Connection refused
- Page not found

**Solution A: Check if enabled**
```yaml
web:
  enabled: true  # Must be explicitly enabled
```

**Solution B: Check listen address**
```yaml
web:
  listen: "127.0.0.1:8080"  # Localhost only
  # OR
  listen: "0.0.0.0:8080"    # All interfaces (less secure)
```

**Solution C: Port conflict**
Another service using port 8080:
```bash
# Check what's using port 8080
sudo lsof -i :8080

# Change dashboard port
web:
  listen: "127.0.0.1:9090"  # Use different port
```

**Solution D: Remote access via SSH tunnel**
```bash
# From your local machine
ssh -L 8080:localhost:8080 user@remote-server

# Then access http://localhost:8080 locally
```

**Verification**:
- Console should show: `[INFO] Web dashboard started successfully`
- Check API directly: `curl http://127.0.0.1:8080/api/stats`
- Check health: `curl http://127.0.0.1:8080/health`

---

### 7. Port Range Expansion Not Working

**Symptoms**:
- Only 1 backend created instead of range
- Unexpected backend count

**Solution**:
Check address format:

```yaml
# Correct formats
backends:
  - address: "127.0.0.1:9070-9089"   # ✅ Creates 20 backends
  - address: "[::1]:8080-8082"       # ✅ IPv6 range

# Incorrect formats
  - address: "127.0.0.1:9070:9089"   # ❌ Wrong separator
  - address: "127.0.0.1:9070 - 9089" # ❌ Spaces
```

**Verification**:
- Console shows: `Backends (total after expansion): 20`
- Web dashboard lists all expanded backends

---

### 8. High Memory Usage

**Symptoms**:
- Memory constantly increasing
- OOM (Out of Memory) errors

**Possible Causes**:

**A. Too many backends**
- Each backend: ~5KB
- 1000 backends = ~5MB
- Solution: Use `max_active_backends` to limit concurrent usage

**B. Goroutine leak**
- Check with: `GODEBUG=gctrace=1 ./socksbalance`
- Update to latest version

**C. Long-lived connections**
- Each connection: ~5KB
- 10,000 connections = ~50MB
- Normal for high-traffic scenarios

---

### 9. Dashboard Shows Old Data

**Symptoms**:
- Backend statuses not updating
- Latencies frozen
- "Last checked" timestamp old

**Solution A: Check health checker**
```yaml
health:
  check_interval: 10s  # Should be running
```

**Solution B: Check browser**
- Hard refresh: `Ctrl+Shift+R` (Windows/Linux) or `Cmd+Shift+R` (Mac)
- Clear cache
- Try incognito mode

**Solution C: Check API directly**
```bash
curl http://127.0.0.1:8080/api/stats
```

**Solution D: Increase refresh interval**
```yaml
web:
  refresh_interval: 1  # Update every second (was 2)
```

---

### 10. Transparent Mode Not Working

**Symptoms**:
- Connections fail
- "SOCKS5 protocol error" in logs

**Solution**:
Try SOCKS5 mode instead:

```yaml
mode: "socks5"  # Full protocol handling
```

**When to use each mode**:
- **Transparent** (default): Fastest, works with most backends
- **SOCKS5**: Better compatibility, slight overhead

---

## Debug Mode

Enable debug logging for detailed troubleshooting:

```yaml
log:
  level: "debug"  # Shows all internal operations
```

Output:
```
[DEBUG] Backend 127.0.0.1:9070 latency: 45ms
[DEBUG] Health check passed: 127.0.0.1:9070
[DEBUG] Selected backend: 127.0.0.1:9070 (latency: 45ms)
[DEBUG] Client 192.168.1.100 → Backend 127.0.0.1:9070 (cached)
```

---

## Performance Tuning

### For Maximum Speed
```yaml
mode: "transparent"              # Zero-copy mode
balancer:
  max_latency: 500ms             # Only fast backends
  sticky_session_ttl: 5m         # Short sessions
  max_active_backends: 0         # Use all (no limit)
health:
  check_interval: 30s            # Less frequent checks
```

### For Maximum Stability
```yaml
mode: "socks5"                   # Full protocol
balancer:
  max_latency: 3000ms            # Forgiving threshold
  sticky_session_ttl: 30m        # Long sessions
  max_active_backends: 5         # Moderate rotation
health:
  check_interval: 10s            # Frequent checks
  failure_threshold: 5           # Forgiving
```

### For GFW Evasion
```yaml
mode: "transparent"
balancer:
  max_latency: 2000ms
  sticky_session_ttl: 20m
  max_active_backends: 3         # Critical: limit exposure
web:
  enabled: true                  # Monitor rotation
```

---

## Getting Help

1. **Check logs**: Set `log.level: "debug"` for detailed output
2. **Check dashboard**: Visual confirmation of backend status
3. **Test manually**: Use `curl -x socks5://...` to test backends
4. **Check issues**: [GitHub Issues](https://github.com/RevEngine3r/SocksBalance/issues)
5. **Report bugs**: Include config, logs, and dashboard screenshot

---

## Useful Commands

### Test Backend Connectivity
```bash
# Test single backend
curl -x socks5://127.0.0.1:9070 https://api.ipify.org

# Test SocksBalance
curl -x socks5://127.0.0.1:1080 https://api.ipify.org

# Check backend latency
time curl -x socks5://127.0.0.1:9070 https://www.google.com > /dev/null
```

### Monitor in Real-Time
```bash
# Watch logs
tail -f /var/log/socksbalance.log

# Watch dashboard API
watch -n 2 'curl -s http://127.0.0.1:8080/api/stats | jq .'

# Monitor connections
watch -n 1 'netstat -an | grep 1080'
```

### Check Ports
```bash
# What's listening
sudo netstat -tulpn | grep socksbalance

# What's using port 8080
sudo lsof -i :8080

# Kill process on port
sudo kill $(sudo lsof -t -i:8080)
```

---

**Still having issues?** Open an issue on [GitHub](https://github.com/RevEngine3r/SocksBalance/issues) with:
- Your `config.yaml`
- Logs with `log.level: "debug"`
- Screenshot of web dashboard
- Steps to reproduce
