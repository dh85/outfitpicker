#!/bin/bash

# Script to test for flakiness by running tests many times
TOTAL_RUNS=1000
FAILED_RUNS=0
SUCCESS_RUNS=0

echo "Running tests $TOTAL_RUNS times to check for flakiness..."
echo "Started at: $(date)"

for i in $(seq 1 $TOTAL_RUNS); do
    if [ $((i % 100)) -eq 0 ]; then
        echo "Progress: $i/$TOTAL_RUNS runs completed"
    fi
    
    # Run tests and capture output
    if go test ./cmd/outfitpicker ./cmd/outfitpicker-admin -timeout=30s > /dev/null 2>&1; then
        SUCCESS_RUNS=$((SUCCESS_RUNS + 1))
    else
        FAILED_RUNS=$((FAILED_RUNS + 1))
        echo "FAILURE at run $i"
        # Run again with verbose output to see the failure
        echo "Re-running with verbose output:"
        go test ./cmd/outfitpicker ./cmd/outfitpicker-admin -v -timeout=30s
        echo "---"
    fi
    
    # Stop if we hit too many failures
    if [ $FAILED_RUNS -gt 10 ]; then
        echo "Too many failures ($FAILED_RUNS), stopping early"
        break
    fi
done

echo "Completed at: $(date)"
echo "Results:"
echo "  Total runs: $((SUCCESS_RUNS + FAILED_RUNS))"
echo "  Successful: $SUCCESS_RUNS"
echo "  Failed: $FAILED_RUNS"
echo "  Success rate: $(echo "scale=2; $SUCCESS_RUNS * 100 / ($SUCCESS_RUNS + $FAILED_RUNS)" | bc -l)%"

if [ $FAILED_RUNS -eq 0 ]; then
    echo "✅ No flakiness detected!"
    exit 0
else
    echo "❌ Flakiness detected: $FAILED_RUNS failures"
    exit 1
fi