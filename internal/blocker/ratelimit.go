package blocker

import (
	"context"
	"sync"
	"time"

	"ddd/internal/logger"
)

// BlockedIP holds information about a blocked IP
type BlockedIP struct {
	IP          string
	BlockedAt   time.Time
	BlockUntil  time.Time
	Reason      string
	BlockCount  int
}

// IPBlocker handles IP blocking and rate limiting
type IPBlocker struct {
	mu               sync.RWMutex
	blockedIPs       map[string]*BlockedIP
	rateLimitedIPs   map[string]time.Time
	blockDuration    int // in seconds
	rateLimitWindow  time.Duration
	log              *logger.Logger
}

// NewIPBlocker creates a new IP blocker
func NewIPBlocker(blockDuration int, log *logger.Logger) *IPBlocker {
	return &IPBlocker{
		blockedIPs:      make(map[string]*BlockedIP),
		rateLimitedIPs:  make(map[string]time.Time),
		blockDuration:   blockDuration,
		rateLimitWindow: 30 * time.Second,
		log:             log,
	}
}

// IsBlocked checks if an IP is currently blocked
func (b *IPBlocker) IsBlocked(ip string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if blocked, exists := b.blockedIPs[ip]; exists {
		// Check if block has expired
		if time.Now().Before(blocked.BlockUntil) {
			return true
		}
		// Block expired, remove it
		delete(b.blockedIPs, ip)
	}

	return false
}

// IsRateLimited checks if an IP is currently rate limited
func (b *IPBlocker) IsRateLimited(ip string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limitedUntil, exists := b.rateLimitedIPs[ip]; exists {
		if time.Now().Before(limitedUntil) {
			return true
		}
		// Rate limit expired
		delete(b.rateLimitedIPs, ip)
	}

	return false
}

// BlockIP blocks an IP address for the configured duration
func (b *IPBlocker) BlockIP(ip, reason string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	blockUntil := time.Now().Add(time.Duration(b.blockDuration) * time.Second)

	if blocked, exists := b.blockedIPs[ip]; exists {
		// IP already blocked, extend block and increment count
		blocked.BlockUntil = blockUntil
		blocked.BlockCount++
		blocked.Reason = reason
	} else {
		// New block
		b.blockedIPs[ip] = &BlockedIP{
			IP:         ip,
			BlockedAt:  time.Now(),
			BlockUntil: blockUntil,
			Reason:     reason,
			BlockCount: 1,
		}
	}

	b.log.LogIPBlocked(ip, reason, b.blockDuration)
	b.log.LogMitigationAction(ip, "block", reason)
}

// RateLimitIP applies rate limiting to an IP
func (b *IPBlocker) RateLimitIP(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	limitUntil := time.Now().Add(b.rateLimitWindow)
	b.rateLimitedIPs[ip] = limitUntil

	b.log.LogIPRateLimited(ip)
	b.log.LogMitigationAction(ip, "rate_limit", "temporary rate limiting applied")
}

// UnblockIP manually unblocks an IP address
func (b *IPBlocker) UnblockIP(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.blockedIPs, ip)
	delete(b.rateLimitedIPs, ip)

	b.log.LogMitigationAction(ip, "unblock", "manually unblocked")
}

// GetBlockedIP returns information about a blocked IP
func (b *IPBlocker) GetBlockedIP(ip string) *BlockedIP {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if blocked, exists := b.blockedIPs[ip]; exists {
		// Return a copy
		return &BlockedIP{
			IP:         blocked.IP,
			BlockedAt:  blocked.BlockedAt,
			BlockUntil: blocked.BlockUntil,
			Reason:     blocked.Reason,
			BlockCount: blocked.BlockCount,
		}
	}

	return nil
}

// GetAllBlockedIPs returns all currently blocked IPs
func (b *IPBlocker) GetAllBlockedIPs() []*BlockedIP {
	b.mu.RLock()
	defer b.mu.RUnlock()

	blocked := make([]*BlockedIP, 0, len(b.blockedIPs))
	now := time.Now()

	for _, ip := range b.blockedIPs {
		if now.Before(ip.BlockUntil) {
			blocked = append(blocked, &BlockedIP{
				IP:         ip.IP,
				BlockedAt:  ip.BlockedAt,
				BlockUntil: ip.BlockUntil,
				Reason:     ip.Reason,
				BlockCount: ip.BlockCount,
			})
		}
	}

	return blocked
}

// StartCleanup periodically cleans up expired blocks and rate limits
func (b *IPBlocker) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.cleanup()
		}
	}
}

// cleanup removes expired blocks and rate limits
func (b *IPBlocker) cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	// Clean up expired blocks
	for ip, blocked := range b.blockedIPs {
		if now.After(blocked.BlockUntil) {
			delete(b.blockedIPs, ip)
		}
	}

	// Clean up expired rate limits
	for ip, limitUntil := range b.rateLimitedIPs {
		if now.After(limitUntil) {
			delete(b.rateLimitedIPs, ip)
		}
	}
}

// GetBlockStats returns statistics about blocking
func (b *IPBlocker) GetBlockStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_blocked"] = len(b.blockedIPs)
	stats["total_rate_limited"] = len(b.rateLimitedIPs)

	return stats
}
