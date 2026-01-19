package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ddd/internal/blocker"
	"ddd/internal/detector"
	"ddd/internal/dns"
	"ddd/internal/logger"
	"ddd/internal/monitor"
)

func main() {
	// Command line flags
	var (
		port         = flag.Int("port", 8053, "DNS server port")
		upstreamDNS  = flag.String("upstream", "8.8.8.8:53", "Upstream DNS server")
		logFile      = flag.String("log", "logs/dns-defense.log", "Log file path")
		rateLimit    = flag.Int("rate-limit", 100, "Max requests per IP per minute")
		blockTime    = flag.Int("block-time", 300, "Block duration in seconds")
	)
	flag.Parse()

	// Initialize logger
	log, err := logger.NewLogger(*logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting DNS DDoS Defense System",
		"port", *port,
		"upstream", *upstreamDNS,
		"rate_limit", *rateLimit,
		"block_time", *blockTime,
	)

	// Initialize components
	trafficMonitor := monitor.NewTrafficMonitor()
	ddosDetector := detector.NewDDoSDetector(*rateLimit, log)
	ipBlocker := blocker.NewIPBlocker(*blockTime, log)

	// Initialize DNS server
	dnsServer := dns.NewServer(
		*port,
		*upstreamDNS,
		trafficMonitor,
		ddosDetector,
		ipBlocker,
		log,
	)

	// Start background cleanup routines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go trafficMonitor.StartCleanup(ctx)
	go ipBlocker.StartCleanup(ctx)

	// Start DNS server
	go func() {
		if err := dnsServer.Start(); err != nil {
			log.Error("DNS server error", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("DNS server started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down DNS server...")
	dnsServer.Stop()
	log.Info("Server stopped gracefully")
}
