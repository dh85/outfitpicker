#!/bin/bash

# Quick script to test for flakiness on macOS
# Usage: ./test-macos-flakiness.sh [number_of_runs]

RUNS=${1:-100}
FAILED=0
SUCCESS=0

echo "Testing for flakiness on macOS with $RUNS runs..."
echo "Started at: $(date)"

for i in $(seq 1 $RUNS); do
    if [ $((i % 10)) -eq 0 ]; then
        echo "Progress: $i/$RUNS"
    fi
    
    if go test ./cmd/outfitpicker ./cmd/outfitpicker-admin -timeout=30s > /dev/null 2>&1; then
        SUCCESS=$((SUCCESS + 1))
    else
        FAILED=$((FAILED + 1))
        echo "❌ FAILURE at run $i"
        # Show the failure
        go test ./cmd/outfitpicker ./cmd/outfitpicker-admin -v -timeout=30s
        
        # Stop if too many failures
        if [ $FAILED -gt 5 ]; then
            echo "Too many failures, stopping"
            break
        fi
    fi
done

echo "Completed at: $(date)"
echo "Results: $SUCCESS passed, $FAILED failed"

if [ $FAILED -eq 0 ]; then
    echo "🎉 No flakiness detected!"
    exit 0
else
    echo "❌ Flakiness detected"
    exit 1
fi