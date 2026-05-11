# MemInfra Installation

MemInfra is designed to be installed by humans or by coding agents.

## For Humans

From a cloned checkout:

```zsh
script/install
```

From a published repository, the intended one-line installer is:

```zsh
curl -fsSL https://raw.githubusercontent.com/mss-boot-io/meminfra/main/script/install | sh
```

Default install locations:

```text
~/.local/bin/meminfra
~/.local/bin/meminfra-mcp
~/.local/share/meminfra/meminfra.db
```

## For LLM Agents

Fetch and follow this guide:

```text
Install MemInfra by reading:
https://raw.githubusercontent.com/mss-boot-io/meminfra/main/docs/install.md
```

Preferred command from an existing checkout:

```zsh
script/install --agents all --scope user
```

Project-local setup:

```zsh
script/install --agents all --scope project --db "$PWD/meminfra.db"
```

## Options

```zsh
script/install --help
```

Useful examples:

```zsh
script/install --agents none
script/install --agents claude,codex
script/install --agents opencode --scope project
script/install --prefix /usr/local --db /var/lib/meminfra/meminfra.db
```

Environment variables:

```text
GO
MEMINFRA_PREFIX
MEMINFRA_DB
MEMINFRA_AGENTS
MEMINFRA_AGENT_SCOPE
MEMINFRA_REPO
MEMINFRA_REF
MEMINFRA_SOURCE_DIR
MEMINFRA_KEEP_SOURCE
```

## Agent Installation

Agent configuration can be run separately:

```zsh
script/install-agent --agents all --scope user \
  --bin "$HOME/.local/bin/meminfra-mcp" \
  --db "$HOME/.local/share/meminfra/meminfra.db"
```

Supported agents:

- Claude Code
- Codex
- OpenCode

### Claude Code

When the `claude` CLI is available, the installer runs:

```zsh
claude mcp add --transport stdio --scope user meminfra \
  -- "$HOME/.local/bin/meminfra-mcp" --db "$HOME/.local/share/meminfra/meminfra.db"
```

For project scope without `claude`, the installer writes `.mcp.json`.
If `.mcp.json` already exists, the installer leaves it unchanged and prints the `meminfra` server snippet to merge manually.

### Codex

The installer writes:

```text
~/.codex/config.toml
```

or for project scope:

```text
.codex/config.toml
```

Configuration:

```toml
[mcp_servers.meminfra]
command = "/home/you/.local/bin/meminfra-mcp"
args = ["--db", "/home/you/.local/share/meminfra/meminfra.db"]
supports_parallel_tool_calls = true
```

### OpenCode

The installer writes:

```text
~/.config/opencode/opencode.jsonc
```

or for project scope:

```text
opencode.jsonc
```

If the OpenCode config already exists, the installer leaves it unchanged and prints the `meminfra` server snippet to merge manually.

Configuration:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "meminfra": {
      "type": "local",
      "enabled": true,
      "command": "/home/you/.local/bin/meminfra-mcp",
      "args": ["--db", "/home/you/.local/share/meminfra/meminfra.db"]
    }
  }
}
```

## Verification

```zsh
meminfra init --db "$HOME/.local/share/meminfra/meminfra.db"
meminfra-mcp --db "$HOME/.local/share/meminfra/meminfra.db"
```

For a raw MCP stdio smoke test:

```zsh
printf 'Content-Length: 46\r\n\r\n{"jsonrpc":"2.0","id":1,"method":"tools/list"}' \
  | meminfra-mcp --db "$HOME/.local/share/meminfra/meminfra.db"
```
