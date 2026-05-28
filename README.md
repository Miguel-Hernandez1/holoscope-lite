# Holoscope Lite

A lightweight observability platform for Go services that captures request traces, tracks endpoint latency, and visualizes live system behavior through a real-time dashboard.

Built with:
- Go
- Docker
- HTML/CSS/JavaScript
- Concurrent in-memory data structures
- Custom request tracing middleware

<img width="500" alt="image" src="https://github.com/user-attachments/assets/39c643d0-c35e-4998-94c2-e737b34fff50" /> 
<img width="500" alt="image" src="https://github.com/user-attachments/assets/8e617bba-0979-4806-8b80-987c59ae505d" />

Key Features:
- Request tracing with unique trace IDs
- Live latency monitoring
- Error-rate tracking
- Real-time dashboard
- Docker deployment
- Zero external dependencies

A request tracing and latency observability tool for Go HTTP services. Runs entirely locally in a single Docker container — no Datadog account, no Prometheus server, no cloud setup required.

```
● holoscope-lite    reqs 142  errors 12 (8.5%)  avg 45.2ms    refreshed 14:23:01
────────────────────────────────────────────────────────────────────────────────
 ENDPOINT LATENCY           │  TRACES
                            │
 GET /simulate/slow ██████  │  2e4c285679   GET  /health            200   1.2ms
 POST /checkout     ██      │  5f9c12ab34   GET  /simulate/slow     200   987ms
 POST /orders       ██      │  3a1b994f21   GET  /simulate/error    500   0.8ms
 GET /users         █       │  8c3f441d88   POST /orders            201   9.1ms
 GET /health        █       │  4d8e229c15   GET  /users             200   5.3ms
```

---

## What it is

Holoscope Lite instruments a Go HTTP server with a custom middleware layer that records every request as a structured trace. Those traces flow into a concurrent in-memory store that computes per-endpoint aggregates. A JSON API exposes the data, and a browser dashboard polls it every 5 seconds.

The result: a complete observability pipeline — from request capture to live UI — implemented from scratch with no external dependencies beyond the Go standard library.

## Why it matters

Most real observability stacks (OpenTelemetry, Jaeger, Datadog) require significant infrastructure just to get started. This project implements the same core mechanics — trace capture, metric aggregation, and live visualization — in a single Go binary that anyone can clone and run in 60 seconds. It demonstrates how those systems work, not just how to configure them.

**Engineering problems this solves:**
- How does a middleware layer capture timing across async handler execution?
- How do you aggregate per-endpoint statistics concurrently without data races?
- How do you serve a live dashboard from the same binary that runs the API?
- How do you build a consistent metrics snapshot when multiple goroutines write simultaneously?

---

## Tech stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22, stdlib only |
| Frontend | HTML, CSS, vanilla JS — no framework |
| Visualization | CSS bar chart (no Chart.js or canvas) |
| Containerization | Docker, Docker Compose |
| Testing | Go `testing` package |

---

## Architecture

```
HTTP Request
    │
    ▼
┌─────────────────────────────────────────┐
│  Observability Middleware               │
│  · generates hex trace ID              │
│  · records request start time          │
└──────────────────┬──────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────┐
│  HTTP Handler                           │
│  /health  /users  /products             │
│  /orders  /checkout  /simulate/*        │
└──────────────────┬──────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────┐
│  Middleware (post-response)             │
│  · captures status code + elapsed ms   │
│  · calls store.Record()                │
└──────────────────┬──────────────────────┘
                   │
          ┌────────▼────────┐
          │  In-Memory Store │  sync.RWMutex
          │                  │
          │  traces[]        │  ring buffer, 1000 cap
          │  metrics{}       │  per-endpoint aggregates
          └────────┬─────────┘
                   │ polled every 5s
                   ▼
┌─────────────────────────────────────────┐
│  Dashboard (browser)                    │
│  · endpoint latency ranking            │
│  · live trace stream, click-to-inspect │
│  · inline stats bar                    │
└─────────────────────────────────────────┘
```

**Key design decisions:**

- **Single binary** — API, observability endpoints, and static files are all served by one Go process. No sidecar, no agent.
- **One lock per snapshot** — `HandleMetrics` calls `store.Metrics()` once and derives global totals from the returned copy, avoiding the inconsistency of two separate lock acquisitions.
- **Ring buffer** — the trace slice is capped at 1000 entries by shifting oldest on write, keeping memory bounded without a GC-heavy data structure.
- **No external deps** — pure stdlib. The binary is ~7MB and builds in under 3 seconds.

---

## Repo structure

```
holoscope-lite/
├── cmd/server/main.go                  # Entry point, route registration
├── internal/
│   ├── api/handlers.go                 # All HTTP handlers
│   └── observability/
│       ├── middleware.go               # Request capture
│       ├── store.go                    # Concurrent metrics store
│       ├── store_test.go               # Unit tests
│       └── types.go                    # TraceRecord, EndpointStats
├── web/
│   ├── dashboard.html                  # Single-page dashboard
│   └── static/
│       ├── app.js                      # Fetch, render, modal logic
│       └── styles.css                  # Gruvbox-dark theme, no framework
├── scripts/
│   └── generate_traffic.sh             # Traffic generator (curl or python3)
├── docs/architecture.md
├── Dockerfile                          # Multi-stage Go build
└── docker-compose.yml
```

---

## Setup

**Prerequisites:** Docker (recommended) or Go 1.22+

### Docker Compose — recommended

```bash
git clone https://github.com/mighdz/holoscope-lite
cd holoscope-lite
docker compose up --build
```

Open **http://localhost:8080/dashboard**

### Local Go

```bash
git clone https://github.com/mighdz/holoscope-lite
cd holoscope-lite
go run cmd/server/main.go
```

Open **http://localhost:8080/dashboard**

---

## Demo

**Step 1 — Start the server**

```bash
docker compose up --build
```

**Step 2 — Generate traffic**

In a second terminal:

```bash
chmod +x scripts/generate_traffic.sh
./scripts/generate_traffic.sh
```

The script sends ~2 requests/second to every endpoint in a round-robin loop. It uses `curl` if available, and falls back to `python3 urllib` automatically.

**Step 3 — Open the dashboard**

http://localhost:8080/dashboard

The dashboard refreshes every 5 seconds. Within 10–15 seconds you will see:

- The header stats bar fill with live request counts, error rate, and average latency
- The left panel populate with a latency ranking — `GET /simulate/slow` will dominate with a wide bar (500ms–2000ms per request)
- The right panel stream incoming traces, newest first
- Clicking any trace row opens an inspection modal with the full trace JSON

**Step 4 — Inspect a slow trace**

Click any row with a high latency value. The modal shows the raw `TraceRecord`:

```json
{
  "trace_id": "5f9c12ab34c11e7a",
  "method": "GET",
  "path": "/simulate/slow",
  "status_code": 200,
  "latency_ms": 1204.5,
  "timestamp": "2025-09-14T14:23:00Z",
  "is_error": false
}
```

**Step 5 — Query the API directly**

```bash
curl http://localhost:8080/observability/metrics | jq .
curl http://localhost:8080/observability/traces | jq '.[0]'
curl http://localhost:8080/observability/traces/<trace_id> | jq .
```

---

## Simulation endpoints

Two endpoints exist specifically to generate observable behavior for demo and testing purposes:

| Endpoint | Behavior |
|---|---|
| `GET /simulate/slow` | Sleeps a random 500–2000ms before responding. Use this to see high-latency traces dominate the latency ranking. |
| `GET /simulate/error` | Always returns HTTP 500. Use this to drive up the error rate in the header stats. |

These endpoints make it possible to demonstrate the observability pipeline without waiting for real errors or slowdowns to occur naturally.

---

## API reference

### Sample app endpoints

| Method | Path | Response |
|---|---|---|
| GET | `/health` | `{"status":"ok","time":"..."}` |
| GET | `/users` | Array of 3 users |
| GET | `/products` | Array of 3 products |
| POST | `/orders` | `{"order_id":"...","status":"created"}` |
| POST | `/checkout` | `{"checkout_id":"...","total":99.99}` |
| GET | `/simulate/slow` | Delayed 500–2000ms |
| GET | `/simulate/error` | HTTP 500 |

### Observability API

| Method | Path | Response |
|---|---|---|
| GET | `/observability/metrics` | Global totals + per-endpoint stats |
| GET | `/observability/traces` | Last 50 traces, newest first |
| GET | `/observability/traces/{id}` | Single trace by ID |

### Dashboard

| Method | Path | |
|---|---|---|
| GET | `/dashboard` | Live browser UI |
| GET | `/static/*` | CSS, JS assets |

---

## Tests

```bash
go test ./internal/observability/... -v
```

Tests cover: trace storage and retrieval, ring buffer cap enforcement, per-endpoint metric aggregation, and trace lookup by ID.

---

## Screenshots

> Run `./scripts/generate_traffic.sh` for 15 seconds, then capture the dashboard.

<img width="800" alt="image" src="https://github.com/user-attachments/assets/8e617bba-0979-4806-8b80-987c59ae505d" />

---

## Roadmap

| Feature | Notes |
|---|---|
| Prometheus `/metrics` export | Drop-in scrape target for existing Prometheus setups |
| SQLite persistence | Traces survive server restart |
| p95 / p99 latency percentiles | Requires histogram or t-digest per endpoint |
| Threshold alerting | `GET /observability/alerts` when error rate or latency exceeds a threshold |
| Time-range filtering | Query traces by start/end timestamp |
| Log line correlation | Attach structured log lines to a trace by trace ID |

---

## License

MIT — see [LICENSE](LICENSE)
