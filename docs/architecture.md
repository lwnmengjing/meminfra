# MemInfra Target Architecture

Last updated: 2026-05-09

## Direction Calibration

The current implementation is still on track with the original MemInfra direction:

- AI-native memory, not a dashboard.
- Local-first SQLite database.
- FTS5-backed retrieval.
- CLI and JSON output for agents.
- Resources, observations, events, and incidents as the first memory surfaces.

The main drift risk is implementation shape, not product direction:

- `cmd/meminfra` is growing beyond argument parsing.
- `internal/store` is carrying persistence, validation, document projection, and query policy.
- If more features are added directly here, the project will become a CLI tool instead of a reusable memory core.

The next implementation work should restore layering before adding HTTP, MCP, discovery, or reconciliation.

## Final Shape

```text
Codex / Claude Code / OpenCode / MCP Client
                  |
              MCP / CLI / HTTP
                  |
            internal/core
                  |
      store + indexing + migration
                  |
            SQLite + FTS5
```

Target packages:

- `cmd/meminfra`
  - CLI parsing and output only.
- `internal/core`
  - Stable application service layer for resources, observations, events, incidents, relationships, search, and future reconciliation.
- `internal/store`
  - Persistence only: GORM CRUD, transactions, migrations, and FTS table writes.
- `internal/model`
  - Database/domain structs shared by core and store.
- `internal/index`
  - Future home for memory document projection and FTS query policy.
- `internal/discovery`
  - Future Prometheus, WireGuard, SSH, cloud, and manual discovery adapters.
- `internal/reconcile`
  - Future controller-like loop: discover, compare, update memory, emit events, refresh topology.
- `internal/mcp`
  - Future MCP tool adapter over `internal/core`.

## Domain Roadmap

Current:

- resources
- observations
- events
- incidents
- memory_documents
- memory_fts

Next domain model:

- relationships
  - `src_resource_id`
  - `dst_resource_id`
  - `relation_type`
  - `metadata_json`
  - `source`
  - `created_at`
  - `updated_at`

Relationship examples:

- WireGuard tunnel
- service dependency
- upstream/downstream
- route relationship
- load balancer backend

## Near-Term Implementation Order

1. Keep CLI behavior compatible while shrinking `cmd/meminfra`.
2. Add `relationships` after the core split.
3. Add relationship search/projection into `memory_documents`.
4. Move memory document projection toward `internal/index`.
5. Add MCP only after core interfaces are stable enough to reuse.

## Guardrails

- Do not add a traditional UI.
- Do not add HTTP before core layering exists.
- Do not add more large CLI command logic directly into `cmd/meminfra`.
- Keep list semantics automation-friendly: filters with no matches return empty arrays, not errors.
- Keep JSON output agent-friendly: JSON fields must be embedded JSON values, not double-encoded strings.
- Keep FTS user queries safe by default; raw FTS query mode should be explicit if added later.
