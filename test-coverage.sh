#!/bin/bash

echo "Running Backend Tests and Coverage Analysis..."
echo "================================================"

# Run tests with coverage
echo "Running unit tests with coverage..."
go test -v -coverprofile=coverage.out ./internal/...

# Generate coverage report
echo "Generating coverage report..."
go tool cover -func=coverage.out

echo ""
echo "Coverage Summary:"
echo "===================="

# Get total coverage percentage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Total Coverage: ${TOTAL_COVERAGE}%"

# Check if we meet atleast  70%
if (( $(echo "$TOTAL_COVERAGE >= 70" | bc -l) )); then
    echo "Coverage requirement met (70%+)"
else
    echo "Coverage requirement NOT met (need 70%+, have ${TOTAL_COVERAGE}%)"
fi

echo ""
echo "🧹 Cleaning up..."
rm -f coverage.out
