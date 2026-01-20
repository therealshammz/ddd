package monitor

import (
	"context"
	"sync"
	"time"
)

// IPStats holds statistics for a single IP address
type IPStats struct {
	RequestCount    int
	LastRequestTime time.Time
	Queries         []QueryInfo
	FirstSeen       time.Time
}

// QueryInfo holds information about a DNS query
type QueryInfo struct {
	Domain    string
	QueryType string
	Timestamp time.Time
}

// TrafficMonitor monitors traffic per IP address
type TrafficMonitor struct {
	mu    sync.RWMutex
	stats map[string]*IPStats
}

// NewTrafficMonitor creates a new traffic monitor
func NewTrafficMonitor() *TrafficMonitor {
	return &TrafficMonitor{
		stats: make(map[string]*IPStats),
	}
}

// RecordRequest records a DNS request from an IP
func (tm *TrafficMonitor) RecordRequest(ip, domain, qtype string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.stats[ip]; !exists {
		tm.stats[ip] = &IPStats{
			FirstSeen: time.Now(),
			Queries:   make([]QueryInfo, 0),
		}
	}

	stats := tm.stats[ip]
	stats.RequestCount++
	stats.LastRequestTime = time.Now()
	
	// Keep only last 100 queries per IP to avoid memory issues
	if len(stats.Queries) >= 100 {
		stats.Queries = stats.Queries[1:]
	}
	
	stats.Queries = append(stats.Queries, QueryInfo{
		Domain:    domain,
		QueryType: qtype,
		Timestamp: time.Now(),
	})
}

// GetIPStats returns statistics for a specific IP
func (tm *TrafficMonitor) GetIPStats(ip string) *IPStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if stats, exists := tm.stats[ip]; exists {
		// Return a copy to avoid race conditions
		statsCopy := &IPStats{
			RequestCount:    stats.RequestCount,
			LastRequestTime: stats.LastRequestTime,
			FirstSeen:       stats.FirstSeen,
			Queries:         make([]QueryInfo, len(stats.Queries)),
		}
		copy(statsCopy.Queries, stats.Queries)
		return statsCopy
	}
	return nil
}

// GetRecentRequestCount returns the number of requests in the last minute
func (tm *TrafficMonitor) GetRecentRequestCount(ip string, duration time.Duration) int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats, exists := tm.stats[ip]
	if !exists {
		return 0
	}

	cutoff := time.Now().Add(-duration)
	count := 0
	
	for _, query := range stats.Queries {
		if query.Timestamp.After(cutoff) {
			count++
		}
	}
	
	return count
}

// GetRecentQueries returns queries from an IP in the specified duration
func (tm *TrafficMonitor) GetRecentQueries(ip string, duration time.Duration) []QueryInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats, exists := tm.stats[ip]
	if !exists {
		return nil
	}

	cutoff := time.Now().Add(-duration)
	recent := make([]QueryInfo, 0)
	
	for _, query := range stats.Queries {
		if query.Timestamp.After(cutoff) {
			recent = append(recent, query)
		}
	}
	
	return recent
}

// StartCleanup periodically cleans up old statistics
func (tm *TrafficMonitor) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.cleanup()
		}
	}
}

// cleanup removes old statistics (older than 1 hour)
func (tm *TrafficMonitor) cleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cutoff := time.Now().Add(-30 * time. Minute)
	
	for ip, stats := range tm.stats {
		if stats.LastRequestTime.Before(cutoff) {
			delete(tm.stats, ip)
		}
	}
}

// GetAllStats returns all current statistics (for monitoring/debugging)
func (tm *TrafficMonitor) GetAllStats() map[string]*IPStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	statsCopy := make(map[string]*IPStats)
	for ip, stats := range tm.stats {
		statsCopy[ip] = &IPStats{
			RequestCount:    stats.RequestCount,
			LastRequestTime: stats.LastRequestTime,
			FirstSeen:       stats.FirstSeen,
		}
	}
	
	return statsCopy
}
