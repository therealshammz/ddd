package dns

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"ddd/internal/blocker"
	"ddd/internal/detector"
	"ddd/internal/logger"
	"ddd/internal/monitor"
)

// Server is the DNS server with DDoS protection
type Server struct {
	port            int
	upstreamDNS     string
	server          *dns.Server
	trafficMonitor  *monitor.TrafficMonitor
	ddosDetector    *detector.DDoSDetector
	ipBlocker       *blocker.IPBlocker
	log             *logger.Logger
	upstreamClient  *dns.Client
}

// NewServer creates a new DNS server
func NewServer(
	port int,
	upstreamDNS string,
	trafficMonitor *monitor.TrafficMonitor,
	ddosDetector *detector.DDoSDetector,
	ipBlocker *blocker.IPBlocker,
	log *logger.Logger,
) *Server {
	s := &Server{
		port:           port,
		upstreamDNS:    upstreamDNS,
		trafficMonitor: trafficMonitor,
		ddosDetector:   ddosDetector,
		ipBlocker:      ipBlocker,
		log:            log,
		upstreamClient: &dns.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Create DNS server
	s.server = &dns.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Net:     "udp",
		Handler: dns.HandlerFunc(s.handleDNSRequest),
	}

	return s
}

// Start starts the DNS server
func (s *Server) Start() error {
	s.log.Info("DNS server listening", "port", s.port)
	return s.server.ListenAndServe()
}

// Stop stops the DNS server
func (s *Server) Stop() error {
	return s.server.Shutdown()
}

// handleDNSRequest handles incoming DNS requests
func (s *Server) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Extract client IP
	clientIP := s.extractClientIP(w.RemoteAddr())

	// Check if IP is blocked
	if s.ipBlocker.IsBlocked(clientIP) {
		s.log.Info("Blocked IP attempted request", "ip", clientIP)
		s.sendRefused(w, r)
		return
	}

	// Check if IP is rate limited
	if s.ipBlocker.IsRateLimited(clientIP) {
		s.log.Info("Rate limited IP request", "ip", clientIP)
		// Still process but with delay
		time.Sleep(500 * time.Millisecond)
	}

	// Extract query information
	if len(r.Question) == 0 {
		s.sendRefused(w, r)
		return
	}

	question := r.Question[0]
	domain := strings.TrimSuffix(question.Name, ".")
	qtype := dns.TypeToString[question.Qtype]

	// Record the request
	s.trafficMonitor.RecordRequest(clientIP, domain, qtype)
	s.log.LogDNSQuery(clientIP, domain, qtype)

	// Analyze traffic for DDoS patterns
	detectionResult := s.ddosDetector.AnalyzeTraffic(clientIP, s.trafficMonitor)

	if detectionResult.IsAttack {
		s.log.Warnw("Attack detected",
			"ip", clientIP,
			"attack_type", detectionResult.AttackType,
			"severity", detectionResult.Severity,
		)

		// Apply mitigation
		if detectionResult.ShouldBlock {
			s.ipBlocker.BlockIP(clientIP, detectionResult.AttackType)
			s.sendRefused(w, r)
			return
		} else {
			// Just rate limit
			s.ipBlocker.RateLimitIP(clientIP)
		}
	}

	// Forward request to upstream DNS server
	s.forwardRequest(w, r)
}

// forwardRequest forwards the DNS request to upstream server
func (s *Server) forwardRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Query upstream DNS
	resp, _, err := s.upstreamClient.Exchange(r, s.upstreamDNS)
	if err != nil {
		s.log.Errorw("Error querying upstream DNS",
			"error", err,
			"upstream", s.upstreamDNS,
		)
		s.sendServerFailure(w, r)
		return
	}

	// Send response back to client
	if err != nil {
    s.log. Errorw("Error writing response", "error", err)
    return  // Exit to prevent sending multiple responses
	}
}

// sendRefused sends a REFUSED response
func (s *Server) sendRefused(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeRefused
	w.WriteMsg(m)
}

// sendServerFailure sends a SERVFAIL response
func (s *Server) sendServerFailure(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeServerFailure
	w.WriteMsg(m)
}

// extractClientIP extracts IP address from remote address
func (s *Server) extractClientIP(addr net.Addr) string {
	switch v := addr.(type) {
	case *net.UDPAddr:
		return v.IP.String()
	case *net.TCPAddr:
		return v.IP.String()
	default:
		// Fallback: try to parse string representation
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return addr.String()
		}
		return host
	}
}
