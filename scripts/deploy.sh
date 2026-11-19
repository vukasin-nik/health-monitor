#!/usr/bin/env bash
# Usage: ./scripts/deploy.sh
# Builds and starts the whole stack, then waits until the service is healthy.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> Building and starting the stack"
docker compose up -d --build

echo "==> Waiting for the service to become healthy"
./scripts/healthcheck.sh http://localhost:8080/health 15

echo "==> Stack is running"
docker compose ps
echo "Service:    http://localhost:8080/metrics"
echo "Prometheus: http://localhost:9090/targets"
echo "Grafana:    http://localhost:3000 (admin / admin)"
