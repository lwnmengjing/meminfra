GO ?= /home/lwx/.g/go/bin/go
GOFMT ?= /home/lwx/.g/go/bin/gofmt
GOFLAGS ?= -tags sqlite_fts5

.PHONY: fmt fmt-check tidy test build install ci

fmt:
	$(GOFMT) -w ./cmd ./internal

fmt-check:
	@test -z "$$($(GOFMT) -l ./cmd ./internal)" || (echo "gofmt required:"; $(GOFMT) -l ./cmd ./internal; exit 1)

tidy:
	$(GO) mod tidy

test:
	$(GO) test $(GOFLAGS) ./...

build:
	$(GO) build $(GOFLAGS) -o bin/meminfra ./cmd/meminfra
	$(GO) build $(GOFLAGS) -o bin/meminfra-mcp ./cmd/meminfra-mcp

install:
	script/install --source-dir .

ci: fmt-check test build
