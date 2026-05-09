# MemInfra MCP Contract

Last updated: 2026-05-09

MemInfra exposes a local MCP stdio server through:

```zsh
bin/meminfra-mcp --db ./meminfra.db
```

The server uses JSON-RPC 2.0 framing over stdio and supports the MCP lifecycle methods needed by common clients:

- `initialize`
- `notifications/initialized`
- `ping`
- `tools/list`
- `tools/call`

## Tool Surface

### `search_memory`

Searches the local FTS5 memory index.

Input:

```json
{
  "query": "frankfurt wg_tunnel",
  "limit": 10
}
```

Output:

```json
{
  "results": [
    {
      "id": 1,
      "doc_type": "relationship",
      "ref_id": 1,
      "title": "node/frankfurt-01 wg_tunnel node/london-01",
      "body": "...",
      "tags": "relationship wg_tunnel manual",
      "rank": -1.23
    }
  ]
}
```

### `list_resources`

Lists recently seen resources.

Input:

```json
{
  "limit": 50
}
```

Output:

```json
{
  "resources": []
}
```

### `get_resource`

Gets one resource by key.

Input:

```json
{
  "key": "node/frankfurt-01"
}
```

Output:

```json
{
  "resource": {
    "id": 1,
    "resource_key": "node/frankfurt-01",
    "kind": "server",
    "metadata_json": {
      "role": "edge"
    }
  }
}
```

### `query_topology`

Queries one-hop topology edges around a resource.

Input:

```json
{
  "resource": "node/frankfurt-01",
  "direction": "out",
  "type": "wg_tunnel",
  "limit": 50
}
```

Direction values:

- `both`
- `in`
- `out`
- `incoming`
- `outgoing`

Output:

```json
{
  "edges": [
    {
      "relationship": {
        "id": 1,
        "relation_type": "wg_tunnel",
        "metadata_json": {
          "interface": "wg0"
        }
      },
      "src_resource": {
        "resource_key": "node/frankfurt-01"
      },
      "dst_resource": {
        "resource_key": "node/london-01"
      }
    }
  ]
}
```

### `list_observations`

Lists observations, optionally filtered by resource and metric.

Input:

```json
{
  "resource": "node/frankfurt-01",
  "metric": "rtt_ms",
  "limit": 50
}
```

Output:

```json
{
  "observations": []
}
```

### `list_events`

Lists events, optionally filtered by resource and event type.

Input:

```json
{
  "resource": "node/frankfurt-01",
  "type": "rtt_spike",
  "limit": 50
}
```

Output:

```json
{
  "events": []
}
```

### `list_incidents`

Lists recent incident memories.

Input:

```json
{
  "limit": 50
}
```

Output:

```json
{
  "incidents": []
}
```

## Output Rules

- Tool calls return both MCP `content` text and `structuredContent`.
- JSON fields such as `metadata_json` and `event_data_json` are embedded JSON values, not double-encoded strings.
- Missing resource filters on list/topology-style calls return empty arrays where list semantics apply.
- FTS queries are safe by default and quote user tokens before `MATCH`.

## Current Scope

This first MCP adapter is read/query oriented. Data can be written through the CLI:

```zsh
bin/meminfra resource upsert --db ./meminfra.db --key node/frankfurt-01 --kind server
bin/meminfra relationship add --db ./meminfra.db --src node/frankfurt-01 --dst node/london-01 --type wg_tunnel
```

Future MCP write tools can be added after the read contract has settled.
