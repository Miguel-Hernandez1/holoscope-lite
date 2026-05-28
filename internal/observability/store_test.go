package observability

import (
	"testing"
	"time"
)

func makeTrace(id, method, path string, status int, latency float64, isError bool) TraceRecord {
	return TraceRecord{
		TraceID:    id,
		Method:     method,
		Path:       path,
		StatusCode: status,
		LatencyMs:  latency,
		Timestamp:  time.Now(),
		IsError:    isError,
	}
}

func TestRecord_StoresTrace(t *testing.T) {
	s := NewStore()
	s.Record(makeTrace("abc123", "GET", "/health", 200, 5.0, false))

	traces := s.Traces()
	if len(traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(traces))
	}
	if traces[0].TraceID != "abc123" {
		t.Errorf("expected trace_id abc123, got %s", traces[0].TraceID)
	}
}

func TestRecord_CapAt1000(t *testing.T) {
	s := NewStore()
	for i := range 1001 {
		s.Record(makeTrace(string(rune('a'+i%26)), "GET", "/health", 200, 1.0, false))
	}
	if len(s.Traces()) != maxTraces {
		t.Errorf("expected %d traces, got %d", maxTraces, len(s.Traces()))
	}
}

func TestMetrics_Aggregation(t *testing.T) {
	s := NewStore()
	s.Record(makeTrace("t1", "GET", "/users", 200, 10.0, false))
	s.Record(makeTrace("t2", "GET", "/users", 200, 20.0, false))
	s.Record(makeTrace("t3", "GET", "/users", 500, 30.0, true))

	m := s.Metrics()
	stat, ok := m["GET /users"]
	if !ok {
		t.Fatal("expected metrics for GET /users")
	}
	if stat.RequestCount != 3 {
		t.Errorf("expected 3 requests, got %d", stat.RequestCount)
	}
	if stat.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", stat.ErrorCount)
	}
	if stat.AvgLatencyMs != 20.0 {
		t.Errorf("expected avg latency 20.0, got %.1f", stat.AvgLatencyMs)
	}
}

func TestTraceByID_Found(t *testing.T) {
	s := NewStore()
	s.Record(makeTrace("xyz999", "POST", "/orders", 201, 15.0, false))

	tr, ok := s.TraceByID("xyz999")
	if !ok {
		t.Fatal("expected to find trace xyz999")
	}
	if tr.Path != "/orders" {
		t.Errorf("expected path /orders, got %s", tr.Path)
	}
}

func TestTraceByID_NotFound(t *testing.T) {
	s := NewStore()
	_, ok := s.TraceByID("doesnotexist")
	if ok {
		t.Fatal("expected trace not found")
	}
}
