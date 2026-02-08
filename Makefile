## Makefile for Cadence (common developer tasks)
## Cross-platform build with automatic version injection from git tags

.PHONY: all build install test fmt tidy lint vet run clean help webhook webhook-dev dev check build-all audit deps-update version

BINARY := cadence
OUT := bin/$(BINARY)

# Detect OS - check for Windows first (more reliable on Windows systems)
ifdef OS
    # Windows (native make or mingw)
    OS_TYPE := windows
    BINARY := cadence.exe
    OUT := cadence.exe
else
    # Unix-like systems
    UNAME_S := $(shell uname -s 2>/dev/null)
    ifeq ($(UNAME_S),Darwin)
        OS_TYPE := darwin
    else
        OS_TYPE := unix
    endif
endif

# Unix/Linux/macOS build with version injection
ifneq ($(OS_TYPE),windows)
VERSION := $(shell git describe --tags 2>/dev/null | sed 's/-[0-9]*-g[0-9a-f]*$$//' || echo "0.1.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags="-X github.com/trycadence/cadence/internal/version.Version=$(VERSION) -X github.com/trycadence/cadence/internal/version.GitCommit=$(COMMIT) -X github.com/trycadence/cadence/internal/version.BuildTime=$(BUILD_TIME)"
endif

all: build

# Platform-specific build targets
build:
ifeq ($(OS_TYPE),windows)
	@powershell -ExecutionPolicy Bypass -File .\scripts\build.ps1
else
	@mkdir -p bin
	@echo "Building Cadence..."
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Time:    $(BUILD_TIME)"
	@go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/cadence || (echo "Build failed!"; exit 1)
	@echo "Build complete: bin/$(BINARY)"
endif

install:
ifeq ($(OS_TYPE),windows)
	@powershell -ExecutionPolicy Bypass -File .\scripts\build.ps1 -Install
else
	go install $(LDFLAGS) ./cmd/cadence
endif

test:
	go test ./...

coverage:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-report:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

coverage-strict:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	THRESHOLD=85; \
	if [ "$$(echo "$$COVERAGE < $$THRESHOLD" | bc -l)" -eq 1 ]; then \
		echo "⚠️  Coverage $$COVERAGE% is below $$THRESHOLD% threshold"; \
		exit 1; \
	else \
		echo "✓ Coverage $$COVERAGE% meets threshold"; \
	fi

fmt:
	go fmt ./...

tidy:
	go mod tidy

lint:
	@golangci-lint run || true

vet:
	go vet ./...

run:
	go run ./cmd/cadence

webhook:
	go run ./cmd/cadence webhook --config .cadence.yaml --port 8000

webhook-dev:
	go run ./cmd/cadence webhook --config .cadence.yaml --port 8000

dev: fmt vet test build
	@echo "✓ Development build complete"

check: fmt vet lint test
	@echo "✓ All checks passed"

build-all:
ifeq ($(OS_TYPE),windows)
	@powershell -ExecutionPolicy Bypass -File .\scripts\build-all.ps1
else
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/cadence-linux-amd64 ./cmd/cadence
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/cadence-darwin-amd64 ./cmd/cadence
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/cadence-darwin-arm64 ./cmd/cadence
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/cadence-windows-amd64.exe ./cmd/cadence
	@echo "✓ Built all platforms: bin/cadence-*"
endif

audit:
	go list -json -m all | go run golang.org/x/vuln/cmd/govulncheck@latest

deps-update:
	go get -u ./...
	go mod tidy
	@echo "✓ Dependencies updated"

version:
	@echo "Cadence $(VERSION)"

clean:
	@rm -rf bin coverage.out coverage.html

help:
	@echo "Cadence Makefile targets:"
	@echo ""
	@echo "Build & Install:"
	@echo "  make build           - Build binary with automatic version injection from git tags"
	@echo "  make build-all       - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make install         - Install binary to \$$GOPATH/bin"
	@echo ""
	@echo "Development:"
	@echo "  make run             - Run application (default CLI)"
	@echo "  make webhook         - Run webhook server on port 8000"
	@echo "  make webhook-dev     - Run webhook server with verbose output"
	@echo "  make dev             - Run full dev cycle: fmt, vet, test, build"
	@echo "  make check           - Run all checks: fmt, vet, lint, test"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test            - Run all tests"
	@echo "  make coverage        - Run tests with coverage report (generates coverage.html)"
	@echo "  make coverage-report - Run tests and display coverage summary"
	@echo "  make coverage-strict - Run tests and enforce 85% coverage threshold"
	@echo "  make lint            - Run linter"
	@echo "  make vet             - Run go vet"
	@echo "  make fmt             - Format code"
	@echo ""
	@echo "Dependencies:"
	@echo "  make tidy            - Tidy dependencies"
	@echo "  make deps-update     - Update all dependencies"
	@echo "  make audit           - Audit dependencies for vulnerabilities"
	@echo ""
	@echo "Utilities:"
	@echo "  make version         - Display current version"
	@echo "  make clean           - Clean build artifacts and coverage files"

