BIN := "./bin/qobserver"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

build-linux:
	GOOS=linux GOARCH=amd64 go build -v -o $(BIN)_linux -ldflags "$(LDFLAGS)" ./cmd

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.3.1

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build build-linux test lint
