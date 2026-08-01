# MemInfra V1 Demo Progress — Historical Record

> **Historical and non-authoritative.** This file no longer defines current architecture, roadmap, environment, or next work.
>
> Replaced on 2026-08-02 by:
>
> - [`PROJECT_MEMORY.md`](PROJECT_MEMORY.md)
> - [`product-contract.md`](product-contract.md)
> - [`architecture.md`](architecture.md)
> - [`data-model-v2.md`](data-model-v2.md)
> - [`roadmap-v2.md`](roadmap-v2.md)
> - [`evaluation-v2.md`](evaluation-v2.md)
> - [`adr/0001-memory-first-redesign.md`](adr/0001-memory-first-redesign.md)

## Historical Scope

The V1 repository was an unused demonstration created in May 2026. It proved that the following components could be connected:

- Go CLI;
- SQLite through GORM;
- FTS5 search;
- resource, observation, event, incident, and relationship tables;
- one-hop topology queries;
- text and JSON CLI output;
- a minimal read-oriented MCP stdio server;
- installation and GitHub Actions scaffolding.

The demo had no production deployment, data, or compatibility requirement.

## Why V1 Was Replaced as the Design Contract

V1 centered mutable entity CRUD and a large persistence service. The nominal `internal/core` mostly forwarded to `internal/store`, while the store owned validation, defaulting, policy, transactions, projections, and search writes.

The result was useful as a technical spike but did not yet implement the defining MemInfra semantics:

- immutable evidence;
- observed versus ingestion versus validity time;
- source provenance and confidence;
- conflict, retraction, and supersession;
- rebuildable state/change/topology projections;
- evidence-linked hypotheses, decisions, actions, and outcomes;
- deterministic explanation context.

V2 therefore permits breaking the V1 schema, CLI, and MCP contracts.

## Historical Implementation Summary

V1 included:

```text
resources
observations
events
incidents
relationships
memory_documents
memory_fts
```

It exposed CLI commands for adding, listing, getting, searching, and querying one-hop relationships, plus MCP tools for search and read/list operations.

The exact historical implementation remains available in Git history before `refactor/memory-core-v2`.

## Historical Validation Claims

The earlier session record reported successful Go format/build/test and CLI smoke checks in its original environment in May 2026. Those results are historical only and do not establish the current branch state.

Current validation must be recorded in `PROJECT_MEMORY.md` with the commit and environment actually checked.

## Known Demo Review Items

PR #1 retained unresolved review comments that should be considered when V2 adapter behavior is implemented:

- CLI search JSON used Go field names instead of the project’s snake_case JSON convention;
- JSON-RPC parse errors could omit `id` instead of returning `id: null`;
- the MCP server version was hard-coded;
- installation documentation did not clearly state CGO/C toolchain prerequisites for the selected demo driver.

These are not V2 compatibility requirements, but the corresponding correctness concerns remain valid.

## Current Source of Truth

Read [`PROJECT_MEMORY.md`](PROJECT_MEMORY.md) before changing the project. Resume from its current checkpoint and the active V2 roadmap, not from this historical file.
