#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

cd "$tmp"

cat > .mcp.json <<'JSON'
{
  "mcpServers": {
    "existing": {
      "command": "existing-server"
    }
  }
}
JSON

PATH="/usr/bin:/bin" "$script_dir/install-agent" --agents claude --scope project --bin /tmp/meminfra-mcp --db /tmp/meminfra.db >/tmp/meminfra-install-agent-claude.out 2>/tmp/meminfra-install-agent-claude.err

if ! grep -q '"existing"' .mcp.json; then
  echo ".mcp.json existing content was modified" >&2
  exit 1
fi
if grep -q '"meminfra"' .mcp.json; then
  echo ".mcp.json was overwritten with meminfra content" >&2
  exit 1
fi

cat > opencode.jsonc <<'JSON'
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "existing": {
      "type": "local",
      "command": "existing-server"
    }
  }
}
JSON

"$script_dir/install-agent" --agents opencode --scope project --bin /tmp/meminfra-mcp --db /tmp/meminfra.db >/tmp/meminfra-install-agent-opencode.out 2>/tmp/meminfra-install-agent-opencode.err

if ! grep -q '"existing"' opencode.jsonc; then
  echo "opencode.jsonc existing content was modified" >&2
  exit 1
fi
if grep -q '"meminfra"' opencode.jsonc; then
  echo "opencode.jsonc was overwritten with meminfra content" >&2
  exit 1
fi

mkdir fresh
cd fresh
"$script_dir/install-agent" --agents opencode --scope project --bin /tmp/meminfra-mcp --db /tmp/meminfra.db >/tmp/meminfra-install-agent-fresh.out
if ! grep -q '"meminfra"' opencode.jsonc; then
  echo "fresh opencode.jsonc did not include meminfra" >&2
  exit 1
fi
