package observability

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware records every request into the store.
func Middleware(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := fmt.Sprintf("%x", rand.Int63())
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			latency := float64(time.Since(start).Microseconds()) / 1000.0

			store.Record(TraceRecord{
				TraceID:    traceID,
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: wrapped.status,
				LatencyMs:  latency,
				Timestamp:  start,
				IsError:    wrapped.status >= 500,
			})
		})
	}
}
