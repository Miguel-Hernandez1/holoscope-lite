package main

import (
	"log"
	"net/http"

	"github.com/mighdz/holoscope-lite/internal/api"
	"github.com/mighdz/holoscope-lite/internal/observability"
)

func main() {
	store := observability.NewStore()
	mux := http.NewServeMux()

	// Sample app endpoints
	mux.HandleFunc("/health", api.HandleHealth)
	mux.HandleFunc("/users", api.HandleUsers)
	mux.HandleFunc("/products", api.HandleProducts)
	mux.HandleFunc("/orders", api.HandleOrders)
	mux.HandleFunc("/checkout", api.HandleCheckout)
	mux.HandleFunc("/simulate/slow", api.HandleSimulateSlow)
	mux.HandleFunc("/simulate/error", api.HandleSimulateError)

	// Observability API
	mux.HandleFunc("/observability/metrics", api.HandleMetrics(store))
	mux.HandleFunc("/observability/traces/", api.HandleTraceByID(store))
	mux.HandleFunc("/observability/traces", api.HandleTraces(store))

	// Dashboard + static assets
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/dashboard.html")
	})

	// Catch-all
	mux.HandleFunc("/", api.NotFound)

	handler := observability.Middleware(store)(mux)

	log.Println("Holoscope Lite listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
