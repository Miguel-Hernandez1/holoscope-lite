#!/usr/bin/env bash
# Sends a continuous stream of requests to all endpoints (~2 req/sec).
# Usage: ./scripts/generate_traffic.sh [base_url]
# Default base URL: http://localhost:8080

BASE="${1:-http://localhost:8080}"

if command -v curl > /dev/null 2>&1; then
  HTTP_CLIENT="curl"
elif command -v python3 > /dev/null 2>&1; then
  HTTP_CLIENT="python3"
else
  echo "error: neither curl nor python3 found. Install one to run this script." >&2
  exit 1
fi

echo "Generating traffic → $BASE  (${HTTP_CLIENT}, Ctrl+C to stop)"

do_request() {
  local method="$1" url="$2"
  if [ "$HTTP_CLIENT" = "curl" ]; then
    curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url"
  else
    python3 -c "
import sys, urllib.request, urllib.error
req = urllib.request.Request(sys.argv[2], method=sys.argv[1])
try:
    with urllib.request.urlopen(req) as r: print(r.status, end='')
except urllib.error.HTTPError as e: print(e.code, end='')
except urllib.error.URLError: print('000', end='')
" "$method" "$url"
  fi
}

ENDPOINTS=(
  "GET /health"
  "GET /users"
  "GET /products"
  "POST /orders"
  "POST /checkout"
  "GET /simulate/slow"
  "GET /simulate/error"
)

COUNT=0
while true; do
  for ep in "${ENDPOINTS[@]}"; do
    METHOD="${ep%% *}"
    REQPATH="${ep#* }"
    STATUS=$(do_request "$METHOD" "$BASE$REQPATH")
    printf "  %-6s %-22s → %s\n" "$METHOD" "$REQPATH" "$STATUS"
    COUNT=$((COUNT + 1))
    sleep 0.45
  done
  echo "  — batch done (total: $COUNT requests) —"
done
