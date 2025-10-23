#!/bin/bash
set -e

COVERAGE_THRESHOLD=75

echo "Running tests with coverage..."
go test -coverprofile=coverage.out ./...

echo "Generating coverage report..."
go tool cover -func=coverage.out > coverage.txt

COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo "Current coverage: ${COVERAGE}%"
echo "Required coverage: ${COVERAGE_THRESHOLD}%"

if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
    echo "❌ Coverage ${COVERAGE}% is below threshold ${COVERAGE_THRESHOLD}%"
    exit 1
else
    echo "✅ Coverage ${COVERAGE}% meets threshold ${COVERAGE_THRESHOLD}%"
fi

echo "Generating HTML report..."
go tool cover -html=coverage.out -o coverage.html
echo "HTML report generated: coverage.html"