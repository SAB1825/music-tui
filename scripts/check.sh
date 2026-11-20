#!/usr/bin/env bash
set -euo pipefail

echo "Checking Go formatting..."
unformatted="$(gofmt -l $(git ls-files -co --exclude-standard '*.go'))"

if [[ -n "$unformatted" ]]; then
    echo "$unformatted"
    exit 1
fi

echo "Running go vet..."
go vet ./...

echo "Running tests..."
go test ./...

echo "Running race detector..."
go test -race ./...

echo "Running linter..."
golangci-lint run

echo "All checks passed."
