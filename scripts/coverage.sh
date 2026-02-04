#!/bin/bash
# Cadence Code Coverage Script
# Generates coverage reports and enforces coverage thresholds

set -e

THRESHOLD=${1:-85}
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

echo "Running tests with coverage..."
go test -v -coverprofile=$COVERAGE_FILE -covermode=atomic ./...

if [ $? -ne 0 ]; then
    echo "❌ Tests failed!"
    exit 1
fi

echo ""
echo "📊 Generating HTML coverage report..."
go tool cover -html=$COVERAGE_FILE -o $COVERAGE_HTML

echo ""
echo "📈 Coverage Summary:"
go tool cover -func=$COVERAGE_FILE | tail -1

echo ""
# Calculate coverage percentage
COVERAGE=$(go tool cover -func=$COVERAGE_FILE | grep total | awk '{print $3}' | sed 's/%//')

echo "Total Coverage: ${COVERAGE}%"
echo "Threshold: ${THRESHOLD}%"
echo ""

# Check against threshold
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "❌ Coverage ${COVERAGE}% is below ${THRESHOLD}% threshold"
    echo "📁 Open $COVERAGE_HTML to view coverage details"
    exit 1
fi

echo "✅ Coverage ${COVERAGE}% meets or exceeds ${THRESHOLD}% threshold"
echo ""
echo "📁 Coverage report: $COVERAGE_HTML"
echo "💡 Run 'go tool cover -html=$COVERAGE_FILE' to view details in browser"
