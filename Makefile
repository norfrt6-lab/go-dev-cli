BINARY_NAME=devx
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-s -w -X github.com/norfrt6-lab/go-dev-cli/cmd.version=$(VERSION) -X github.com/norfrt6-lab/go-dev-cli/cmd.commit=$(COMMIT) -X github.com/norfrt6-lab/go-dev-cli/cmd.buildTime=$(BUILD_TIME)"

.PHONY: build run test lint clean install fmt vet

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

run: build
	./$(BINARY_NAME)

test:
	go test -race -coverprofile=coverage.out ./...

test-verbose:
	go test -race -v ./...

lint:
	golangci-lint run --timeout=5m

fmt:
	gofumpt -w .

vet:
	go vet ./...

clean:
	rm -f $(BINARY_NAME) coverage.out

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/
