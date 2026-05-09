# MemInfra

MemInfra is an AI-native infrastructure memory layer.

The current MVP is intentionally local-first:

- SQLite database
- GORM-backed resource, observation, event, incident, and relationship storage
- SQLite FTS5 memory search
- CLI workflow
- MCP stdio server for AI agents

FTS5 requires building the SQLite driver with the `sqlite_fts5` tag. The
provided `Makefile` sets this for `make test` and `make build`.

## Installation

For humans:

```zsh
script/install
```

For LLM agents, paste this instruction into Claude Code, Codex, OpenCode, Cursor, or another coding agent:

```text
Install and configure MemInfra by following:
https://raw.githubusercontent.com/lwnmengjing/ai-infra-operator/main/docs/install.md
```

One-line installer for the published repository:

```zsh
curl -fsSL https://raw.githubusercontent.com/lwnmengjing/ai-infra-operator/main/script/install | sh
```

Agent-specific install:

```zsh
script/install-agent --agents claude,codex,opencode --scope user
```

See [docs/install.md](docs/install.md) for all options.

## Quick Start

Build the CLI and MCP server:

```zsh
make build
```

This creates:

```text
bin/meminfra
bin/meminfra-mcp
```

Seed a local memory database:

```zsh
bin/meminfra init --db ./meminfra.db
bin/meminfra resource upsert --db ./meminfra.db --key node/frankfurt-01 --kind server --hostname frankfurt-01 --provider ovh --region fra --ipv4 192.0.2.10 --metadata '{"role":"edge"}'
bin/meminfra observe add --db ./meminfra.db --resource node/frankfurt-01 --metric rtt_ms --value 82 --unit ms --source manual
bin/meminfra event add --db ./meminfra.db --resource node/frankfurt-01 --type rtt_spike --data '{"region":"fra"}' --source manual
bin/meminfra incident add --db ./meminfra.db --title "Frankfurt RTT spike" --symptoms "RTT increased" --root-cause "OVH upstream congestion" --solution "Shift traffic to London" --result "Latency recovered" --tags "frankfurt rtt ovh"
bin/meminfra resource upsert --db ./meminfra.db --key node/london-01 --kind server --hostname london-01
bin/meminfra relationship add --db ./meminfra.db --src node/frankfurt-01 --dst node/london-01 --type wg_tunnel --metadata '{"interface":"wg0"}'
bin/meminfra search --db ./meminfra.db "Frankfurt RTT"
```

Useful query commands:

```zsh
bin/meminfra resource get --db ./meminfra.db --key node/frankfurt-01 --output json
bin/meminfra resource list --db ./meminfra.db --output json
bin/meminfra observe list --db ./meminfra.db --resource node/frankfurt-01 --metric rtt_ms --output json
bin/meminfra event list --db ./meminfra.db --resource node/frankfurt-01 --type rtt_spike --output json
bin/meminfra incident list --db ./meminfra.db --output json
bin/meminfra relationship list --db ./meminfra.db --resource node/frankfurt-01 --type wg_tunnel --output json
bin/meminfra relationship topology --db ./meminfra.db --resource node/frankfurt-01 --direction out --output json
```

For agent-friendly output, add `--output json` to any command:

```zsh
bin/meminfra search --db ./meminfra.db --output json "Frankfurt RTT"
```

## MCP Server

Run the local MCP stdio server:

```zsh
bin/meminfra-mcp --db ./meminfra.db
```

Example MCP client configuration:

```json
{
  "mcpServers": {
    "meminfra": {
      "command": "/home/lwx/go/src/github.com/lwnmengjing/ai-infra-operator/bin/meminfra-mcp",
      "args": [
        "--db",
        "/home/lwx/go/src/github.com/lwnmengjing/ai-infra-operator/meminfra.db"
      ]
    }
  }
}
```

Available MCP tools:

```text
search_memory
list_resources
get_resource
query_topology
list_observations
list_events
list_incidents
```

See `docs/mcp-contract.md` for the tool input and output contract.
