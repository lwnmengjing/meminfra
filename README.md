# MemInfra

MemInfra is an AI-native infrastructure memory layer.

The current MVP is intentionally local-first:

- SQLite database
- GORM-backed resource, observation, and event storage
- SQLite FTS5 memory search
- CLI-only workflow

FTS5 requires building the SQLite driver with the `sqlite_fts5` tag. The
provided `Makefile` sets this for `make test` and `make build`.

## Quick Start

```zsh
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra init --db ./meminfra.db
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra resource upsert --db ./meminfra.db --key node/frankfurt-01 --kind server --hostname frankfurt-01 --provider ovh --region fra --ipv4 192.0.2.10
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra observe add --db ./meminfra.db --resource node/frankfurt-01 --metric rtt_ms --value 82 --unit ms --source manual
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra event add --db ./meminfra.db --resource node/frankfurt-01 --type rtt_spike --data '{"region":"fra"}' --source manual
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra incident add --db ./meminfra.db --title "Frankfurt RTT spike" --symptoms "RTT increased" --root-cause "OVH upstream congestion" --solution "Shift traffic to London" --result "Latency recovered" --tags "frankfurt rtt ovh"
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra search --db ./meminfra.db "Frankfurt RTT"
```

Useful query commands:

```zsh
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra resource get --db ./meminfra.db --key node/frankfurt-01 --output json
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra resource list --db ./meminfra.db --output json
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra observe list --db ./meminfra.db --resource node/frankfurt-01 --metric rtt_ms --output json
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra event list --db ./meminfra.db --resource node/frankfurt-01 --type rtt_spike --output json
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra incident list --db ./meminfra.db --output json
```

For agent-friendly output, add `--output json` to any command:

```zsh
/home/lwx/.g/go/bin/go run -tags sqlite_fts5 ./cmd/meminfra search --db ./meminfra.db --output json "Frankfurt RTT"
```
