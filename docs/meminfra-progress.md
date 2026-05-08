# MemInfra Progress Memory

Last updated: 2026-05-08

## Project Memory

MemInfra is an AI-native infrastructure memory layer.

It is not intended to be a traditional monitoring dashboard, CMDB, or Grafana replacement. The first-class user is an AI agent such as Codex, Claude Code, OpenCode, Cursor, or an MCP client.

Core principles:

- AI Native: the system stores infrastructure state in forms that agents can query and reason over.
- MCP First later: MCP is the intended AI access protocol, but it is not part of the current MVP.
- Local First: a single SQLite database is the primary memory store.
- Observed Truth: infrastructure truth comes from observations, events, probes, metrics, and operator actions, not only cloud APIs.
- No traditional UI for MVP: CLI and future MCP tools are the interaction layer.

Long-term direction:

- Observe infrastructure.
- Analyze state and history.
- Decide on likely causes or changes.
- Act through controlled automation.
- Learn from outcomes and preserve operational memory.

## Current MVP Scope

The active MVP is intentionally small:

- Go CLI runtime.
- SQLite storage through `gorm.io/driver/sqlite` and `gorm.io/gorm`.
- Resources, observations, events, and memory documents.
- SQLite FTS5 search over memory documents.
- No HTTP API.
- No MCP adapter.
- No discovery engine.
- No reconciliation loop.
- No incident workflow beyond the generic searchable memory document foundation.

## Implemented So Far

Project files added:

- `go.mod` and `go.sum`
- `Makefile`
- `README.md`
- `.gitignore`
- `cmd/meminfra/main.go`
- `internal/model/models.go`
- `internal/store/store.go`
- `internal/store/store_test.go`

Implemented behavior:

- `meminfra init --db PATH`
  - Opens SQLite and applies GORM migrations.
  - Creates the FTS5 virtual table.
- `meminfra resource upsert`
  - Inserts or updates resources by `resource_key`.
  - Preserves `first_seen`.
  - Updates `last_seen`.
  - Refreshes the resource search document.
- `meminfra observe add`
  - Records numeric observations for existing resources.
  - Creates searchable observation memory documents.
- `meminfra event add`
  - Records events for existing resources.
  - Creates searchable event memory documents.
- `meminfra search`
  - Queries SQLite FTS5 and returns matching memory documents.

## Data Model

Current tables:

- `resources`
  - `resource_key`, `kind`, `hostname`, `ipv4`, `ipv6`, `provider`, `region`, `source`, `metadata_json`, `first_seen`, `last_seen`
- `observations`
  - `resource_id`, `metric`, `value`, `unit`, `source`, `metadata_json`, `observed_at`
- `events`
  - `resource_id`, `event_type`, `event_data_json`, `source`, `created_at`
- `memory_documents`
  - `doc_type`, `ref_id`, `title`, `body`, `tags`, `created_at`, `updated_at`
- `memory_fts`
  - FTS5 virtual table with `title`, `body`, `tags`

FTS implementation note:

- FTS5 is maintained manually as an independent virtual table.
- It is not using SQLite external-content mode.
- This avoids malformed index behavior observed during early tests with manual delete/insert against an external-content table.

## Environment Memory

The preferred shell for future work is zsh:

```zsh
/usr/bin/zsh
```

The Codex process may not have Go in `PATH`. Use absolute Go tool paths:

```zsh
/home/lwx/.g/go/bin/go
/home/lwx/.g/go/bin/gofmt
```

Known local Go environment:

- Go version: `go1.26.0 linux/amd64`
- `CGO_ENABLED=1`
- `GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct`
- `GOROOT=/home/lwx/.g/versions/1.26.0`
- `GOPATH=/home/lwx/go`

Sandbox note:

- Default Go build cache under `/home/lwx/.cache/go-build` was read-only from Codex.
- Default module cache under `/home/lwx/go/pkg/mod` was also not writable from Codex.
- Use `/tmp` caches when running Go commands from Codex:

```zsh
GOCACHE=/tmp/meminfra-go-build GOMODCACHE=/tmp/meminfra-go-mod /home/lwx/.g/go/bin/go test -tags sqlite_fts5 ./...
```

## Build And Test Memory

FTS5 requires the `sqlite_fts5` build tag because `gorm.io/driver/sqlite` uses `github.com/mattn/go-sqlite3` underneath.

Use:

```zsh
make test
make build
```

Equivalent explicit commands:

```zsh
GOCACHE=/tmp/meminfra-go-build GOMODCACHE=/tmp/meminfra-go-mod /home/lwx/.g/go/bin/go test -tags sqlite_fts5 ./...
GOCACHE=/tmp/meminfra-go-build GOMODCACHE=/tmp/meminfra-go-mod /home/lwx/.g/go/bin/go build -tags sqlite_fts5 -o /tmp/meminfra ./cmd/meminfra
```

Verified on 2026-05-08:

- `go mod tidy`: passed after using `/tmp` caches and approved network access.
- `go test -tags sqlite_fts5 ./...`: passed.
- `go build -tags sqlite_fts5 -o /tmp/meminfra ./cmd/meminfra`: passed.
- CLI smoke test passed with:
  - `init`
  - `resource upsert`
  - `observe add`
  - `event add`
  - `search`

Review fixes applied on 2026-05-08:

- FTS search now converts user input into safe quoted FTS phrases before `MATCH`.
- IPv6-style queries such as `2001:db8::1` are covered by tests and CLI smoke validation.
- JSON text fields are strict: invalid `metadata_json` and `event_data_json` inputs are rejected before persistence.
- Subcommand `-h` output is visible and exits successfully.

Smoke database path used:

```zsh
/tmp/meminfra-smoke.db
```

## Important Decisions

- SQLite driver must be `gorm.io/driver/sqlite`.
- GORM is used for ordinary table schema and CRUD.
- Raw SQL is used for FTS5 virtual table creation and search.
- JSON payload fields are stored as `text` columns for now.
- No `gorm.io/datatypes` dependency is used in the current implementation.
- The MVP stores resource, observation, and event content into generic `memory_documents` so future incident memory and MCP retrieval can build on the same search surface.

## Current Git State

The repository started empty.

Current work is uncommitted and consists of new project files. No previous user code was modified.

Before committing, run:

```zsh
git status --short
make test
```

## Recommended Next Steps

Next implementation slice:

- Add CLI JSON output mode for agent-friendly consumption.
- Add incident memory as a first-class command and table.
- Add list/get commands for resources, observations, and events.
- Add a small query API layer inside `internal/core` before introducing HTTP or MCP.

Future larger slices:

- Discovery engine.
- Reconciliation loop.
- HTTP JSON API.
- MCP adapter.
- Incident/RCA workflow.
- Relationship graph.
- Prometheus/WireGuard/SSH importers.
