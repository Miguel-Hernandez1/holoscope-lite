package observability

import "time"

type TraceRecord struct {
	TraceID    string    `json:"trace_id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	LatencyMs  float64   `json:"latency_ms"`
	Timestamp  time.Time `json:"timestamp"`
	IsError    bool      `json:"is_error"`
}

type EndpointStats struct {
	Path           string  `json:"path"`
	RequestCount   int64   `json:"request_count"`
	ErrorCount     int64   `json:"error_count"`
	TotalLatencyMs float64 `json:"-"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
}
