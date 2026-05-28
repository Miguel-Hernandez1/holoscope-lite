# Architecture

## Data Flow

```
HTTP Request
     │
     ▼
┌─────────────────────────────┐
│   Observability Middleware  │  ← generate trace_id, record start time
│   internal/observability/   │
│   middleware.go             │
└────────────┬────────────────┘
             │
             ▼
┌─────────────────────────────┐
│        HTTP Handler         │  ← business logic (handlers.go)
│   /health, /users, etc.     │
└────────────┬────────────────┘
             │
             ▼
┌─────────────────────────────┐
│   Middleware (post-handler) │  ← record status, latency, store TraceRecord
└────────────┬────────────────┘
             │
             ▼
┌─────────────────────────────┐
│      In-Memory Store        │
│   internal/observability/   │
│   store.go                  │
│                             │
│  traces[]       ─── ring buffer, last 1000 records    │
│  metrics{}      ─── per-endpoint aggregate stats      │
└────────────┬────────────────┘
             │
             ▼ (polled every 5s)
┌─────────────────────────────┐
│     Dashboard (browser)     │
│   GET /observability/metrics│
│   GET /observability/traces │
│   Plain HTML/CSS/JS         │
└─────────────────────────────┘
```

## Components

| Component | Location | Responsibility |
|---|---|---|
| HTTP Server | `cmd/server/main.go` | Route registration, server startup |
| Sample Handlers | `internal/api/handlers.go` | Fake business endpoints + observability API |
| Middleware | `internal/observability/middleware.go` | Intercept every request, capture timing |
| Store | `internal/observability/store.go` | Thread-safe in-memory trace + metrics storage |
| Types | `internal/observability/types.go` | Shared data structures |
| Dashboard | `web/dashboard.html` + `web/static/` | Live browser UI |

## Design Decisions

- **Single binary**: everything — API, observability, and static files — is served by one Go process. No separate services.
- **In-memory storage**: traces reset on restart. This is intentional for v0.1; a future upgrade could add SQLite or Prometheus.
- **No dependencies**: only Go stdlib. Keeps the binary small and the code easy to read.
- **Trace ID**: a random `int63` hex string. Collision probability is negligible for a demo.

## Future Upgrades

- Export metrics in Prometheus format (`/metrics`)
- Persist traces to SQLite
- Add a `/alerts` endpoint for threshold-based alerts
- Add p95/p99 latency percentiles to the metrics store
- Add log streaming to the dashboard
