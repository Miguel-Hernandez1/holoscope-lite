package api

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	obs "github.com/mighdz/holoscope-lite/internal/observability"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func randomID() string {
	return fmt.Sprintf("%x", rand.Int63())
}

// --- Sample app endpoints ---

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func HandleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	users := []map[string]any{
		{"id": "u1", "name": "Alice Chen", "email": "alice@example.com", "role": "admin"},
		{"id": "u2", "name": "Bob Torres", "email": "bob@example.com", "role": "user"},
		{"id": "u3", "name": "Carol Kim", "email": "carol@example.com", "role": "user"},
	}
	writeJSON(w, http.StatusOK, users)
}

func HandleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	products := []map[string]any{
		{"id": "p1", "name": "Widget Pro", "price": 29.99, "stock": 142},
		{"id": "p2", "name": "Gadget Plus", "price": 49.99, "stock": 57},
		{"id": "p3", "name": "Doohickey Max", "price": 99.99, "stock": 23},
	}
	writeJSON(w, http.StatusOK, products)
}

func HandleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"order_id":   randomID(),
		"status":     "created",
		"created_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func HandleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"checkout_id": randomID(),
		"status":      "completed",
		"total":       99.99,
		"paid_at":     time.Now().UTC().Format(time.RFC3339),
	})
}

func HandleSimulateSlow(w http.ResponseWriter, r *http.Request) {
	delayMs := 500 + rand.Intn(1501) // 500–2000 ms
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	writeJSON(w, http.StatusOK, map[string]any{
		"simulated":  "slow",
		"latency_ms": delayMs,
	})
}

func HandleSimulateError(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "simulated internal error",
	})
}

// NotFound handles unmatched routes.
func NotFound(w http.ResponseWriter, r *http.Request) {
	// Only fire for truly unknown paths; skip sub-path matches.
	if r.URL.Path != "/" {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("route not found: %s %s", r.Method, r.URL.Path),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"service": "holoscope-lite", "status": "running"})
}

// --- Observability API endpoints ---

type metricsResponse struct {
	TotalRequests int64               `json:"total_requests"`
	TotalErrors   int64               `json:"total_errors"`
	ErrorRatePct  float64             `json:"error_rate_pct"`
	AvgLatencyMs  float64             `json:"avg_latency_ms"`
	Endpoints     []obs.EndpointStats `json:"endpoints"`
}

func HandleMetrics(store *obs.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Single lock acquisition: compute globals and endpoint list from one snapshot.
		metrics := store.Metrics()

		var totalReqs, totalErrors int64
		var totalLatency float64
		endpoints := make([]obs.EndpointStats, 0, len(metrics))
		for _, stat := range metrics {
			totalReqs += stat.RequestCount
			totalErrors += stat.ErrorCount
			totalLatency += stat.TotalLatencyMs
			endpoints = append(endpoints, stat)
		}

		var avgLatency, errorRate float64
		if totalReqs > 0 {
			avgLatency = totalLatency / float64(totalReqs)
			errorRate = float64(totalErrors) / float64(totalReqs) * 100
		}

		writeJSON(w, http.StatusOK, metricsResponse{
			TotalRequests: totalReqs,
			TotalErrors:   totalErrors,
			ErrorRatePct:  round2(errorRate),
			AvgLatencyMs:  round2(avgLatency),
			Endpoints:     endpoints,
		})
	}
}

func HandleTraces(store *obs.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traces := store.Traces()
		// Return at most 50 most recent
		if len(traces) > 50 {
			traces = traces[:50]
		}
		writeJSON(w, http.StatusOK, traces)
	}
}

func HandleTraceByID(store *obs.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
		id := parts[len(parts)-1]

		trace, ok := store.TraceByID(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "trace not found"})
			return
		}
		writeJSON(w, http.StatusOK, trace)
	}
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}
