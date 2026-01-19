package detector

import (
	"testing"
	"time"

	"ddd/internal/logger"
	"ddd/internal/monitor"
)

func TestHighRequestRate(t *testing.T) {
	// Create test logger (to /dev/null in tests)
	log, _ := logger.NewLogger("/tmp/test.log")
	detector := NewDDoSDetector(100, log)
	trafficMonitor := monitor.NewTrafficMonitor()

	// Simulate 150 requests in one minute
	testIP := "192.168.1.100"
	for i := 0; i < 150; i++ {
		trafficMonitor.RecordRequest(testIP, "example.com", "A")
	}

	result := detector.AnalyzeTraffic(testIP, trafficMonitor)

	if !result.IsAttack {
		t.Error("Expected high request rate to be detected as attack")
	}

	if result.AttackType != "high_request_rate" {
		t.Errorf("Expected attack type 'high_request_rate', got '%s'", result.AttackType)
	}

	if !result.ShouldBlock {
		t.Error("Expected IP to be blocked for exceeding 2x rate limit")
	}
}

func TestRepeatedQueries(t *testing.T) {
	log, _ := logger.NewLogger("/tmp/test.log")
	detector := NewDDoSDetector(100, log)
	trafficMonitor := monitor.NewTrafficMonitor()

	testIP := "192.168.1.101"
	
	// Query the same domain 30 times
	for i := 0; i < 30; i++ {
		trafficMonitor.RecordRequest(testIP, "same-domain.com", "A")
	}

	result := detector.AnalyzeTraffic(testIP, trafficMonitor)

	if !result.IsAttack {
		t.Error("Expected repeated queries to be detected as attack")
	}

	if result.AttackType != "repeated_queries" {
		t.Errorf("Expected attack type 'repeated_queries', got '%s'", result.AttackType)
	}
}

func TestRandomSubdomainAttack(t *testing.T) {
	log, _ := logger.NewLogger("/tmp/test.log")
	detector := NewDDoSDetector(100, log)
	trafficMonitor := monitor.NewTrafficMonitor()

	testIP := "192.168.1.102"
	
	// Query 30 random subdomains
	for i := 0; i < 30; i++ {
		domain := "random123abc" + string(rune(i)) + ".example.com"
		trafficMonitor.RecordRequest(testIP, domain, "A")
	}

	result := detector.AnalyzeTraffic(testIP, trafficMonitor)

	if !result.IsAttack {
		t.Error("Expected random subdomain attack to be detected")
	}

	if result.AttackType != "random_subdomain" {
		t.Errorf("Expected attack type 'random_subdomain', got '%s'", result.AttackType)
	}
}

func TestNormalTraffic(t *testing.T) {
	log, _ := logger.NewLogger("/tmp/test.log")
	detector := NewDDoSDetector(100, log)
	trafficMonitor := monitor.NewTrafficMonitor()

	testIP := "192.168.1.103"
	
	// Normal queries - different domains, low rate
	domains := []string{"google.com", "github.com", "example.com", "cloudflare.com"}
	for i := 0; i < 20; i++ {
		domain := domains[i%len(domains)]
		trafficMonitor.RecordRequest(testIP, domain, "A")
		time.Sleep(10 * time.Millisecond)
	}

	result := detector.AnalyzeTraffic(testIP, trafficMonitor)

	if result.IsAttack {
		t.Errorf("Normal traffic should not be detected as attack, got: %s", result.AttackType)
	}
}
