#!/usr/bin/env bash
set -euo pipefail

echo "=== ERP-AIOps Test Suite ==="
echo ""

FAIL=0

echo "--- Go Tests ---"
cd "$(dirname "$0")/.."
go test ./... 2>&1 || FAIL=$((FAIL + 1))

echo ""
echo "--- Rust Tests ---"
cargo test --workspace 2>&1 || FAIL=$((FAIL + 1))

echo ""
echo "--- Python Tests ---"
cd services/ai-brain
python -m pytest 2>&1 || FAIL=$((FAIL + 1))
cd ../..

echo ""
echo "--- Frontend Tests ---"
cd web
npm run test 2>&1 || FAIL=$((FAIL + 1))
cd ..

echo ""
if [ "$FAIL" -gt 0 ]; then
    echo "=== $FAIL test suite(s) failed ==="
    exit 1
else
    echo "=== All tests passed ==="
fi
