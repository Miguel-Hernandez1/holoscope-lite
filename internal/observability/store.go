package observability

import "sync"

const maxTraces = 1000

type Store struct {
	mu      sync.RWMutex
	traces  []TraceRecord
	metrics map[string]*EndpointStats
}

func NewStore() *Store {
	return &Store{
		metrics: make(map[string]*EndpointStats),
	}
}

func (s *Store) Record(t TraceRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ring buffer: drop oldest when full
	if len(s.traces) >= maxTraces {
		s.traces = s.traces[1:]
	}
	s.traces = append(s.traces, t)

	key := t.Method + " " + t.Path
	stat, ok := s.metrics[key]
	if !ok {
		stat = &EndpointStats{Path: key}
		s.metrics[key] = stat
	}
	stat.RequestCount++
	stat.TotalLatencyMs += t.LatencyMs
	stat.AvgLatencyMs = stat.TotalLatencyMs / float64(stat.RequestCount)
	if t.IsError {
		stat.ErrorCount++
	}
}

// Traces returns a copy of all traces, newest first.
func (s *Store) Traces() []TraceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]TraceRecord, len(s.traces))
	for i, t := range s.traces {
		out[len(s.traces)-1-i] = t
	}
	return out
}

// TraceByID finds a trace by its ID.
func (s *Store) TraceByID(id string) (TraceRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := len(s.traces) - 1; i >= 0; i-- {
		if s.traces[i].TraceID == id {
			return s.traces[i], true
		}
	}
	return TraceRecord{}, false
}

// Metrics returns a snapshot of per-endpoint stats.
func (s *Store) Metrics() map[string]EndpointStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]EndpointStats, len(s.metrics))
	for k, v := range s.metrics {
		out[k] = *v
	}
	return out
}
