# Makefile for OpenTrusty Core

.PHONY: build test lint clean help

help:
	@echo "OpenTrusty Core Makefile"
	@echo "Usage:"
	@echo "  make build    - Verify compilation of all packages"
	@echo "  make test     - Run all tests"
	@echo "  make lint     - Run linter (requires golangci-lint)"
	@echo "  make clean    - Clean build artifacts"

build:
	go build ./...

deps:
	go mod download
	go mod tidy

test: test-service

test-unit:
	go test -v -short ./...

test-service:
	go test -v ./...

lint:
	golangci-lint run ./...

clean:
	go clean -cache
	rm -f coverage.out
