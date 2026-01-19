package detector

import (
	"strings"
	"time"

	"ddd/internal/logger"
	"ddd/internal/monitor"
)

// DDoSDetector detects various DDoS attack patterns
type DDoSDetector struct {
	rateLimit int // Max requests per minute
	log       *logger.Logger
}

// NewDDoSDetector creates a new DDoS detector
func NewDDoSDetector(rateLimit int, log *logger.Logger) *DDoSDetector {
	return &DDoSDetector{
		rateLimit: rateLimit,
		log:       log,
	}
}

// DetectionResult holds the result of DDoS detection
type DetectionResult struct {
	IsAttack    bool
	AttackType  string
	Severity    string // "low", "medium", "high"
	Description string
	ShouldBlock bool
}

// AnalyzeTraffic analyzes traffic from an IP and detects DDoS patterns
func (d *DDoSDetector) AnalyzeTraffic(ip string, trafficMonitor *monitor.TrafficMonitor) *DetectionResult {
	result := &DetectionResult{
		IsAttack:    false,
		ShouldBlock: false,
	}

	// Check 1: High request rate
	recentCount := trafficMonitor.GetRecentRequestCount(ip, 1*time.Minute)
	if recentCount > d.rateLimit {
		result.IsAttack = true
		result.AttackType = "high_request_rate"
		result.Severity = d.calculateSeverity(recentCount, d.rateLimit)
		result.Description = "Excessive request rate detected"
		result.ShouldBlock = recentCount > d.rateLimit*2
		
		d.log.LogDDoSDetected(ip, "high request rate", recentCount)
		return result
	}

	// Check 2: Repeated queries (same domain queried many times)
	queries := trafficMonitor.GetRecentQueries(ip, 1*time.Minute)
	if repeatedQueriesDetected := d.checkRepeatedQueries(queries); repeatedQueriesDetected {
		result.IsAttack = true
		result.AttackType = "repeated_queries"
		result.Severity = "medium"
		result.Description = "Repeated queries to same domain detected"
		result.ShouldBlock = true
		
		d.log.LogDDoSDetected(ip, "repeated queries", len(queries))
		return result
	}

	// Check 3: Random subdomain attack
	if randomSubdomainAttack := d.checkRandomSubdomains(queries); randomSubdomainAttack {
		result.IsAttack = true
		result.AttackType = "random_subdomain"
		result.Severity = "high"
		result.Description = "Random subdomain attack detected"
		result.ShouldBlock = true
		
		d.log.LogDDoSDetected(ip, "random subdomain attack", len(queries))
		return result
	}

	// Check 4: Query burst (many queries in very short time)
	if burstDetected := d.checkQueryBurst(queries); burstDetected {
		result.IsAttack = true
		result.AttackType = "query_burst"
		result.Severity = "medium"
		result.Description = "Query burst detected"
		result.ShouldBlock = false // Rate limit instead of block
		
		d.log.LogDDoSDetected(ip, "query burst", len(queries))
		return result
	}

	return result
}

// checkRepeatedQueries detects if the same domain is queried repeatedly
func (d *DDoSDetector) checkRepeatedQueries(queries []monitor.QueryInfo) bool {
	if len(queries) < 20 {
		return false
	}

	domainCounts := make(map[string]int)
	for _, q := range queries {
		domainCounts[q.Domain]++
	}

	// If any domain is queried more than 50% of the time, it's suspicious
	for _, count := range domainCounts {
		if float64(count)/float64(len(queries)) > 0.5 && count > 10 {
			return true
		}
	}

	return false
}

// checkRandomSubdomains detects random subdomain attacks
func (d *DDoSDetector) checkRandomSubdomains(queries []monitor.QueryInfo) bool {
	if len(queries) < 30 {
		return false
	}

	// Extract base domains and subdomains
	baseDomains := make(map[string][]string)
	
	for _, q := range queries {
		parts := strings.Split(q.Domain, ".")
		if len(parts) >= 2 {
			// Get base domain (last two parts)
			baseDomain := strings.Join(parts[len(parts)-2:], ".")
			subdomain := strings.Join(parts[:len(parts)-2], ".")
			
			if subdomain != "" {
				baseDomains[baseDomain] = append(baseDomains[baseDomain], subdomain)
			}
		}
	}

	// Check if many unique subdomains for same base domain
	for _, subdomains := range baseDomains {
		uniqueSubdomains := make(map[string]bool)
		for _, sub := range subdomains {
			uniqueSubdomains[sub] = true
		}
		
		// If more than 20 unique subdomains, likely random subdomain attack
		if len(uniqueSubdomains) > 20 {
			return true
		}
		
		// Check if subdomains look random (contain many numbers/random chars)
		randomCount := 0
		for sub := range uniqueSubdomains {
			if d.looksRandom(sub) {
				randomCount++
			}
		}
		
		if randomCount > 10 {
			return true
		}
	}

	return false
}

// checkQueryBurst detects sudden bursts of queries
func (d *DDoSDetector) checkQueryBurst(queries []monitor.QueryInfo) bool {
	if len(queries) < 10 {
		return false
	}

	// Check if more than 50 queries in last 10 seconds
	cutoff := time.Now().Add(-10 * time.Second)
	recentCount := 0
	
	for _, q := range queries {
		if q.Timestamp.After(cutoff) {
			recentCount++
		}
	}

	return recentCount > 50
}

// looksRandom checks if a string looks randomly generated
func (d *DDoSDetector) looksRandom(s string) bool {
	if len(s) < 8 {
		return false
	}

	digitCount := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			digitCount++
		}
	}

	// If more than 50% digits, likely random
	return float64(digitCount)/float64(len(s)) > 0.5
}

// calculateSeverity calculates attack severity based on request count
func (d *DDoSDetector) calculateSeverity(requestCount, limit int) string {
	ratio := float64(requestCount) / float64(limit)
	
	if ratio > 5 {
		return "high"
	} else if ratio > 2 {
		return "medium"
	}
	return "low"
}
