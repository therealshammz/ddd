# Troubleshooting Guide

## Issue: "cannot find module providing package"

**Problem:** Import paths don't match your module name.

**Solution:**
The project is configured with module name `ddd`. All files should use:
```go
import "ddd/internal/logger"
```

If you renamed the directory, update `go.mod`:
```
module your-new-name

go 1.21
```

Then update all import statements in:
- `cmd/server/main.go`
- `internal/dns/server.go`
- `internal/detector/ddos.go`
- `internal/blocker/ratelimit.go`
- `test/detector_test.go`

## Issue: Port 53 Already in Use

**Problem:**
```
bind: address already in use
```

**Solutions:**

1. **Use a different port:**
```bash
./dns-defense-server -port 5353
```

2. **Check what's using port 53:**
```bash
sudo lsof -i :53
```

3. **Stop systemd-resolved (Ubuntu/Debian):**
```bash
sudo systemctl stop systemd-resolved
sudo systemctl disable systemd-resolved
```

4. **Configure systemd-resolved to not bind to port 53:**
Edit `/etc/systemd/resolved.conf`:
```ini
[Resolve]
DNSStubListener=no
```
Then restart:
```bash
sudo systemctl restart systemd-resolved
```

## Issue: Permission Denied on Port 53

**Problem:**
```
listen udp :53: bind: permission denied
```

**Solutions:**

1. **Run with sudo:**
```bash
sudo ./dns-defense-server
```

2. **Give binary permission to bind to privileged ports:**
```bash
sudo setcap 'cap_net_bind_service=+ep' ./dns-defense-server
./dns-defense-server
```

3. **Use a non-privileged port (>1024):**
```bash
./dns-defense-server -port 5353
```

## Issue: No Response from Upstream DNS

**Problem:**
```
Error querying upstream DNS: i/o timeout
```

**Solutions:**

1. **Check internet connectivity:**
```bash
ping 8.8.8.8
```

2. **Try a different upstream DNS:**
```bash
# Cloudflare
./dns-defense-server -upstream 1.1.1.1:53

# Quad9
./dns-defense-server -upstream 9.9.9.9:53

# OpenDNS
./dns-defense-server -upstream 208.67.222.222:53
```

3. **Check firewall rules:**
```bash
sudo iptables -L -n | grep 53
```

## Issue: Build Fails

**Problem:**
```
package X is not in GOROOT
```

**Solution:**
```bash
# Clean and rebuild
make clean
go mod download
go mod tidy
make build
```

## Issue: Tests Fail

**Problem:**
```
panic: runtime error: invalid memory address
```

**Solution:**
```bash
# Ensure logs directory exists
mkdir -p logs

# Run tests with verbose output
go test -v ./test/
```

## Issue: High Memory Usage

**Problem:** Server consuming too much memory.

**Solutions:**

1. **The traffic monitor keeps last 100 queries per IP** - this is normal
2. **Cleanup runs every 5 minutes** to remove old data
3. **For production, consider adding limits on total IPs tracked**

## Issue: Logs Not Being Written

**Problem:** Log file is empty.

**Solutions:**

1. **Check log directory exists:**
```bash
mkdir -p logs
```

2. **Check permissions:**
```bash
ls -la logs/
chmod 755 logs
```

3. **Specify absolute path:**
```bash
./dns-defense-server -log /var/log/dns-defense.log
```

## Issue: DNS Queries Not Being Answered

**Problem:** Client gets no response.

**Debug steps:**

1. **Check if server is running:**
```bash
ps aux | grep dns-defense
```

2. **Check if listening on correct port:**
```bash
sudo netstat -ulnp | grep 53
# or
sudo ss -ulnp | grep 53
```

3. **Test with verbose dig:**
```bash
dig @localhost -p 5353 google.com +trace
```

4. **Check logs for errors:**
```bash
tail -f logs/dns-defense.log
```

## Issue: Too Many False Positives

**Problem:** Legitimate users being blocked.

**Solutions:**

1. **Increase rate limit:**
```bash
./dns-defense-server -rate-limit 200
```

2. **Decrease block time:**
```bash
./dns-defense-server -block-time 60
```

3. **Tune detection thresholds** in `internal/detector/ddos.go`:
- Increase minimum queries for detection
- Adjust percentage thresholds
- Modify random subdomain detection sensitivity

## Issue: Module Download Fails

**Problem:**
```
go: github.com/miekg/dns@...: Get "https://proxy.golang.org/...": dial tcp: lookup proxy.golang.org: no such host
```

**Solutions:**

1. **Check internet connection**

2. **Set Go proxy:**
```bash
export GOPROXY=https://proxy.golang.org,direct
go mod download
```

3. **Use direct mode:**
```bash
export GOPROXY=direct
go mod download
```

## Getting Help

If you're still having issues:

1. **Check the logs:**
```bash
tail -100 logs/dns-defense.log
```

2. **Run with verbose output:**
```bash
./dns-defense-server -port 5353 2>&1 | tee debug.log
```

3. **Test DNS resolution manually:**
```bash
# Test upstream directly
dig @8.8.8.8 google.com

# Test your server
dig @localhost -p 5353 google.com
```

4. **Check Go environment:**
```bash
go env
```

## Common Workflow Issues

### "I can't build the project"
```bash
# Full reset and rebuild
make clean
rm -rf go.sum
go mod tidy
go mod download
make build
```

### "I want to start fresh"
```bash
# Clean everything
make clean
rm -rf logs/*
rm go.sum

# Rebuild
./setup.sh
```

### "I need to change the module name"

1. Edit `go.mod` and change the module line
2. Find and replace all imports in:
   - `cmd/server/main.go`
   - All files in `internal/*/`
3. Run `go mod tidy`

Example for module name `myproject`:
```bash
# Update go.mod
sed -i 's/module ddd/module myproject/' go.mod

# Update imports (Linux)
find . -name "*.go" -exec sed -i 's|ddd/internal|myproject/internal|g' {} +

# Update imports (macOS)
find . -name "*.go" -exec sed -i '' 's|ddd/internal|myproject/internal|g' {} +
```
