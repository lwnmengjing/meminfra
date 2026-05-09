GO ?= /home/lwx/.g/go/bin/go
GOFMT ?= /home/lwx/.g/go/bin/gofmt
GOFLAGS ?= -tags sqlite_fts5

.PHONY: fmt tidy test build

fmt:
	$(GOFMT) -w ./cmd ./internal

tidy:
	$(GO) mod tidy

test:
	$(GO) test $(GOFLAGS) ./...

build:
	$(GO) build $(GOFLAGS) -o bin/meminfra ./cmd/meminfra
	$(GO) build $(GOFLAGS) -o bin/meminfra-mcp ./cmd/meminfra-mcp
