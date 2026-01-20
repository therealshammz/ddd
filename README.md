# DNS DDoS Defense System

A Go-based DNS server with built-in DDoS protection capabilities including traffic monitoring, attack detection, rate limiting, and IP blocking.

## Features

- **DNS Server**: Responds to DNS requests and forwards to upstream DNS
- **Traffic Monitoring**: Tracks requests per IP address
- **DDoS Detection**: Detects multiple attack patterns:
  - High request rate
  - Repeated queries
  - Random subdomain attacks
  - Query bursts
- **Mitigation**: Automatically blocks or rate limits attackers
- **Logging**: Comprehensive activity and mitigation logging

## Prerequisites

- Go 1.21 or higher
- Root/sudo privileges (for binding to port 53)
- Linux/Unix environment recommended

## Installation

1. **Clone the project:**
```bash
git clone https://github.com/therealshammz/ddd.git
```

2. **Initialize and download dependencies:**
```bash
cd ddd
 # Use install script
./setup.sh
```

## Running

### Basic Usage

```bash
# Run with default settings (requires root for port 53)
sudo ./dns-defense-server

# Or run on non-privileged port
./dns-defense-server -port 5353
```

### Command Line Options

```bash
./dns-defense-server [options]

Options:
  -port int
        DNS server port (default 53)
  -upstream string
        Upstream DNS server (default "8.8.8.8:53")
  -log string
        Log file path (default "logs/dns-defense.log")
  -rate-limit int
        Max requests per IP per minute (default 100)
  -block-time int
        Block duration in seconds (default 300)
```

### Example Configurations

```bash
# Run on port 5353 with Cloudflare DNS and strict rate limiting
./dns-defense-server -port 5353 -upstream 1.1.1.1:53 -rate-limit 50

# Run with longer block time
sudo ./dns-defense-server -block-time 600 -rate-limit 75

# Run with custom log location
sudo ./dns-defense-server -log /var/log/dns-defense.log
```

## Testing

### Test DNS Resolution

```bash
# Query the DNS server
dig @localhost -p 5353 example.com

# Or with nslookup
nslookup example.com localhost
```

### Test Rate Limiting

```bash
# Send many requests quickly (will trigger rate limiting)
for i in {1..150}; do dig @localhost -p 5353 test$i.example.com; done
```

### Test Random Subdomain Detection

```bash
# Simulate random subdomain attack
for i in {1..50}; do 
  dig @localhost -p 5353 $(openssl rand -hex 8).example.com
done
```

## Monitoring

### View Logs

```bash
# Tail logs in real-time
tail -f logs/dns-defense.log

# Search for blocked IPs
grep "IP Blocked" logs/dns-defense.log

# View DDoS detections
grep "DDoS Pattern Detected" logs/dns-defense.log
```

### Log Format

Logs are in JSON format for easy parsing:

```json
{
  "level": "warn",
  "timestamp": "2026-01-19T10:30:45.123Z",
  "msg": "DDoS Pattern Detected",
  "client_ip": "192.168.1.100",
  "reason": "high request rate",
  "request_count": 150,
  "event": "ddos_detected"
}
```

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ DNS Query
       ▼
┌─────────────────────────────────────┐
│        DNS Server (Port 53)         │
├─────────────────────────────────────┤
│  1. Check if IP is blocked          │
│  2. Check if IP is rate limited     │
│  3. Record request in monitor       │
│  4. Analyze for DDoS patterns       │
│  5. Apply mitigation if needed      │
│  6. Forward to upstream DNS         │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────┐
│  Upstream   │
│ DNS Server  │
└─────────────┘
```

## Attack Detection Logic

### High Request Rate
- Triggers when requests exceed configured limit per minute
- Default: 100 requests/minute
- Severity based on how much limit is exceeded

### Repeated Queries
- Detects when same domain is queried >50% of the time
- Minimum 20 queries required for detection
- Often indicates DNS amplification attacks

### Random Subdomain Attack
- Detects >20 unique subdomains for same base domain
- Identifies randomly generated subdomain patterns
- Common in DNS water torture attacks

### Query Burst
- Detects >50 queries in 10-second window
- Indicates sudden attack spike
- Applies rate limiting rather than blocking

## Mitigation Actions

### Rate Limiting
- Applied for less severe patterns
- 30-second rate limit window
- Adds 500ms delay to requests

### IP Blocking
- Applied for severe attack patterns
- Default block duration: 5 minutes (300 seconds)
- Repeated attacks extend block duration

## Project Structure

```
ddd/
.
├── cmd
│   └── server
│       └── main.go
├── configs
├── dns-defense-server
├── go.mod
├── go.sum
├── internal
│   ├── blocker
│   │   └── ratelimit.go
│   ├── detector
│   │   └── ddos.go
│   ├── dns
│   │   └── server.go
│   ├── logger
│   │   └── logger.go
│   └── monitor
│       └── traffic.go
├── logs
├── Makefile
├── output.txt
├── QUICKSTART.md
├── README.md
├── SDLC_WORKFLOW.mermaid
├── setup.sh
├── test
│   └── detector_test.go
└── TROUBLESHOOTING.md

```

## Production Deployment

### System Service (systemd)

Create `/etc/systemd/system/dns-defense.service`:

```ini
[Unit]
Description=DNS DDoS Defense Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/dns-defense-server -port 53 -log /var/log/dns-defense.log
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable dns-defense.service
sudo systemctl start dns-defense.service
sudo systemctl status dns-defense.service
```

### Security Considerations

1. **Run with minimal privileges**: Consider using capabilities instead of root
2. **Firewall rules**: Restrict access to DNS port
3. **Log rotation**: Set up logrotate for log files
4. **Monitoring**: Integrate with monitoring systems
5. **Backup DNS**: Have fallback DNS servers configured

## Troubleshooting

### Port 53 Already in Use

```bash
# Check what's using port 53
sudo lsof -i :53

# Stop systemd-resolved if it's blocking
sudo systemctl stop systemd-resolved
```

### Permission Denied

```bash
# Run with sudo for privileged ports
sudo ./dns-defense-server

# Or use setcap to allow binding to port 53
sudo setcap 'cap_net_bind_service=+ep' ./dns-defense-server
```

## License

MIT License - feel free to use and modify for your needs.

## Contributing

Contributions welcome! Please test thoroughly before submitting pull requests.
