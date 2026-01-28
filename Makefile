## Makefile for Cadence (common developer tasks)

.PHONY: all build install test fmt tidy lint vet run clean

BINARY := cadence
OUT := bin/$(BINARY)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags="-X github.com/codemeapixel/cadence/internal/version.Version=$(VERSION) -X github.com/codemeapixel/cadence/internal/version.GitCommit=$(COMMIT) -X github.com/codemeapixel/cadence/internal/version.BuildTime=$(BUILD_TIME)"

all: build

build:
	@mkdir -p bin
	go build $(LDFLAGS) -o $(OUT) ./cmd/cadence

install:
	go install $(LDFLAGS) ./cmd/cadence

test:
	go test ./...

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

clean:
	@rm -rf bin

