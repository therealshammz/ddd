package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(logFile string) (*Logger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{logFile, "stdout"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}, nil
}

// LogDNSQuery logs a DNS query
func (l *Logger) LogDNSQuery(clientIP, domain, qtype string) {
	l.Infow("DNS Query",
		"client_ip", clientIP,
		"domain", domain,
		"query_type", qtype,
		"event", "dns_query",
	)
}

// LogDDoSDetected logs when DDoS is detected
func (l *Logger) LogDDoSDetected(clientIP, reason string, requestCount int) {
	l.Warnw("DDoS Pattern Detected",
		"client_ip", clientIP,
		"reason", reason,
		"request_count", requestCount,
		"event", "ddos_detected",
	)
}

// LogIPBlocked logs when an IP is blocked
func (l *Logger) LogIPBlocked(clientIP, reason string, duration int) {
	l.Warnw("IP Blocked",
		"client_ip", clientIP,
		"reason", reason,
		"block_duration_seconds", duration,
		"event", "ip_blocked",
		"action", "block",
	)
}

// LogIPRateLimited logs when an IP is rate limited
func (l *Logger) LogIPRateLimited(clientIP string) {
	l.Warnw("IP Rate Limited",
		"client_ip", clientIP,
		"event", "rate_limited",
		"action", "rate_limit",
	)
}

// LogMitigationAction logs any mitigation action taken
func (l *Logger) LogMitigationAction(clientIP, action, reason string) {
	l.Infow("Mitigation Action",
		"client_ip", clientIP,
		"action", action,
		"reason", reason,
		"event", "mitigation",
	)
}
