GO ?= go
GOFMT ?= gofmt
GOFLAGS ?= -tags sqlite_fts5

.PHONY: fmt fmt-check tidy vet test build install installer-test ci

fmt:
	$(GOFMT) -w ./cmd ./internal

fmt-check:
	@test -z "$$($(GOFMT) -l ./cmd ./internal)" || (echo "gofmt required:"; $(GOFMT) -l ./cmd ./internal; exit 1)

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet $(GOFLAGS) ./...

test:
	$(GO) test $(GOFLAGS) ./...

build:
	$(GO) build $(GOFLAGS) -o bin/meminfra ./cmd/meminfra
	$(GO) build $(GOFLAGS) -o bin/meminfra-mcp ./cmd/meminfra-mcp

install:
	script/install --source-dir .

ci: fmt-check installer-test vet test build

installer-test:
	sh script/install-agent-test.sh
