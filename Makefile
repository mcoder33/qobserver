BIN := ./bin/qobserver

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

.DEFAULT_GOAL := help

help: ## Show available targets
	@awk 'BEGIN {FS=":.*##"; printf "Usage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

download-deps: ## Download Go module dependencies
	go mod download

build: download-deps ## Build binary for current OS/arch into $(BIN)
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

build-linux: download-deps ## Build Linux amd64 binary into $(BIN)_linux
	GOOS=linux GOARCH=amd64 go build -v -o $(BIN)_linux -ldflags "$(LDFLAGS)" ./cmd

test: ## Run tests with race detector
	go test -race ./internal/...

install-lint-deps: ## Install golangci-lint if missing
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.3.1

lint: install-lint-deps ## Run golangci-lint
	go vet ./...
	$(shell go env GOPATH)/bin/golangci-lint run ./...

check: test lint

.PHONY: help download-deps build build-linux test install-lint-deps lint
