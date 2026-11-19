#!/usr/bin/env bash
# Usage: ./scripts/healthcheck.sh [URL] [RETRIES]
# Exits 0 when the URL answers with HTTP 2xx, exits 1 after all retries fail.
set -euo pipefail

URL="${1:-http://localhost:8080/health}"
RETRIES="${2:-10}"

for i in $(seq 1 "$RETRIES"); do
  if curl -fsS --max-time 3 "$URL" > /dev/null; then
    echo "OK: $URL is healthy"
    exit 0
  fi
  echo "Attempt $i/$RETRIES failed, retrying in 2s..."
  sleep 2
done

echo "FAIL: $URL is not healthy after $RETRIES attempts" >&2
exit 1
