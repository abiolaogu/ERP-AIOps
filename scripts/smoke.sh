#!/usr/bin/env bash
set -euo pipefail

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8090}"
RUST_API_URL="${RUST_API_URL:-http://localhost:8080}"
AI_BRAIN_URL="${AI_BRAIN_URL:-http://localhost:8001}"

PASS=0
FAIL=0

check() {
    local name="$1"
    local url="$2"
    local expected_status="${3:-200}"

    status=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    if [ "$status" = "$expected_status" ]; then
        echo "  PASS  $name ($url) -> $status"
        PASS=$((PASS + 1))
    else
        echo "  FAIL  $name ($url) -> $status (expected $expected_status)"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== ERP-AIOps Smoke Tests ==="
echo ""

echo "--- Gateway Health ---"
check "Gateway /healthz" "$GATEWAY_URL/healthz"
check "Gateway /v1/capabilities" "$GATEWAY_URL/v1/capabilities"
check "Gateway /v1/aiops/health" "$GATEWAY_URL/v1/aiops/health"

echo ""
echo "--- Rust API ---"
check "Rust API /healthz" "$RUST_API_URL/healthz"
check "Rust API /api/v1/incidents" "$RUST_API_URL/api/v1/incidents"
check "Rust API /api/v1/anomalies" "$RUST_API_URL/api/v1/anomalies"
check "Rust API /api/v1/rules" "$RUST_API_URL/api/v1/rules"
check "Rust API /api/v1/topology" "$RUST_API_URL/api/v1/topology"
check "Rust API /api/v1/cost/analysis" "$RUST_API_URL/api/v1/cost/analysis"
check "Rust API /api/v1/security/findings" "$RUST_API_URL/api/v1/security/findings"

echo ""
echo "--- AI Brain ---"
check "AI Brain /health" "$AI_BRAIN_URL/health"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
