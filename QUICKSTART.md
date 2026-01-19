# Quick Start Guide - DNS DDoS Defense

## Setup (5 minutes)

### 1. Navigate to project directory
```bash
cd dns-ddos-defense
```

### 2. Install dependencies
```bash
make deps
# or manually:
go mod download
go mod tidy
```

### 3. Create logs directory
```bash
make setup
# or manually:
mkdir -p logs
```

### 4. Build the application
```bash
make build
```

## Running the Server

### Option 1: Run on non-privileged port (easiest)
```bash
make run
# Server will run on port 5353
```

Test it:
```bash
dig @localhost -p 5353 google.com
```

### Option 2: Run on standard DNS port 53 (requires sudo)
```bash
make run-sudo
```

Test it:
```bash
dig @localhost google.com
```

## Testing DDoS Protection

### Test 1: Trigger High Request Rate
```bash
# Send 150 requests (rate limit is 100/min)
for i in {1..150}; do 
  dig @localhost -p 5353 test$i.example.com +short
done

# Check logs to see detection and blocking
tail -f logs/dns-defense.log
```

### Test 2: Trigger Repeated Queries
```bash
# Query same domain 40 times
for i in {1..40}; do 
  dig @localhost -p 5353 same-domain.com +short
done
```

### Test 3: Trigger Random Subdomain Attack
```bash
# Generate random subdomains
for i in {1..35}; do 
  RANDOM_SUB=$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 12 | head -n 1)
  dig @localhost -p 5353 $RANDOM_SUB.example.com +short
done
```

## Monitoring

### View live logs
```bash
tail -f logs/dns-defense.log
```

### Filter for attacks
```bash
grep "DDoS Pattern Detected" logs/dns-defense.log
```

### Filter for blocked IPs
```bash
grep "IP Blocked" logs/dns-defense.log
```

### Filter for specific IP
```bash
grep "192.168.1.100" logs/dns-defense.log
```

## Configuration Options

Run with custom settings:

```bash
# Use Cloudflare DNS as upstream
./dns-defense-server -port 5353 -upstream 1.1.1.1:53

# Stricter rate limiting
./dns-defense-server -port 5353 -rate-limit 50

# Longer block duration (10 minutes)
./dns-defense-server -port 5353 -block-time 600

# All together
./dns-defense-server \
  -port 5353 \
  -upstream 1.1.1.1:53 \
  -rate-limit 75 \
  -block-time 600 \
  -log logs/custom.log
```

## Understanding the Logs

### Normal DNS Query
```json
{
  "level": "info",
  "timestamp": "2026-01-19T10:30:45Z",
  "msg": "DNS Query",
  "client_ip": "192.168.1.100",
  "domain": "example.com",
  "query_type": "A",
  "event": "dns_query"
}
```

### DDoS Detection
```json
{
  "level": "warn",
  "timestamp": "2026-01-19T10:31:20Z",
  "msg": "DDoS Pattern Detected",
  "client_ip": "192.168.1.100",
  "reason": "high request rate",
  "request_count": 150,
  "event": "ddos_detected"
}
```

### IP Blocked
```json
{
  "level": "warn",
  "timestamp": "2026-01-19T10:31:21Z",
  "msg": "IP Blocked",
  "client_ip": "192.168.1.100",
  "reason": "high_request_rate",
  "block_duration_seconds": 300,
  "event": "ip_blocked",
  "action": "block"
}
```

## Common Issues

### Port 53 already in use
```bash
# Check what's using it
sudo lsof -i :53

# If it's systemd-resolved
sudo systemctl stop systemd-resolved

# Or just use port 5353
./dns-defense-server -port 5353
```

### Permission denied on port 53
```bash
# Option 1: Run with sudo
sudo ./dns-defense-server

# Option 2: Give binary permission
sudo setcap 'cap_net_bind_service=+ep' ./dns-defense-server
./dns-defense-server
```

### No upstream DNS response
```bash
# Check connectivity
dig @8.8.8.8 google.com

# Try different upstream
./dns-defense-server -upstream 1.1.1.1:53
```

## Next Steps

1. Run unit tests: `make test`
2. Check test coverage: `make test-coverage`
3. Review README.md for production deployment
4. Configure as system service (see README)
5. Set up log rotation
6. Configure firewall rules

## Development

### Run tests
```bash
make test
```

### Format code
```bash
make fmt
```

### Clean build artifacts
```bash
make clean
```

### Rebuild
```bash
make clean build
```

## Architecture Quick Reference

```
Client Request
    ↓
[IP Blocked?] → Yes → Refuse
    ↓ No
[IP Rate Limited?] → Yes → Add delay
    ↓ No/Continue
[Record Traffic]
    ↓
[Analyze for DDoS Patterns]
    ↓
[High Rate?] [Repeated?] [Random Subdomains?] [Burst?]
    ↓ If detected
[Apply Mitigation: Block or Rate Limit]
    ↓
[Forward to Upstream DNS]
    ↓
Return Response
```

## Support

- Check logs first: `tail -f logs/dns-defense.log`
- Review README.md for detailed documentation
- Test with: `dig @localhost -p 5353 example.com`
