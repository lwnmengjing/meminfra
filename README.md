# MemInfra

MemInfra is an **AI-native operational memory for infrastructure**.

It records timestamped infrastructure evidence, preserves provenance and causal context, derives current state and changes, and exposes explainable operational context to AI agents and human operators.

## Development Status

The repository currently contains an early SQLite/GORM/FTS5 CLI and MCP demonstration. It is **not production-ready and is not a compatibility contract**.

A memory-first V2 redesign is active on `refactor/memory-core-v2`. The redesign intentionally permits breaking the demo schema, CLI, and MCP contracts because no production deployment or data migration must be preserved.

Before contributing, read the authoritative checkpoint:

- [Project Memory](docs/PROJECT_MEMORY.md)
- [Product Contract](docs/product-contract.md)
- [Target Architecture](docs/architecture.md)
- [V2 Data Model](docs/data-model-v2.md)
- [Development Roadmap](docs/roadmap-v2.md)
- [Evaluation Plan](docs/evaluation-v2.md)

## What MemInfra Must Answer

MemInfra is designed to help an agent answer:

1. What is the current known state of this resource?
2. Which evidence supports that state, when was it observed, and how trustworthy is it?
3. What changed recently?
4. Which events, topology changes, decisions, and actions are correlated with the change?
5. Have we seen a similar incident before?
6. What action was taken, and what happened afterward?
7. Is an older conclusion still valid for the current time and topology?

Every meaningful answer must be traceable to stable evidence identifiers and must expose stale, conflicting, or missing information.

## Core Principles

- **Memory first:** durable evidence and its semantics are the product; CLI, MCP, and indexes are adapters.
- **Observed truth:** state is derived from observations, events, probes, source systems, and operator actions.
- **Temporal by default:** observed time, ingestion time, and validity are distinct.
- **Append-only evidence:** corrections use supersession or retraction; projections are rebuildable.
- **Explainable:** conclusions return supporting and contradicting evidence.
- **Deterministic core before LLM reasoning:** ordering, freshness, conflicts, topology, and change detection are implemented in Go.
- **Local first:** the first complete product runs over an inspectable SQLite database.
- **Safe before autonomous:** early collectors are read-only and infrastructure execution is deferred.

## Product Boundary

MemInfra is not intended to become:

- a Grafana or monitoring replacement;
- a full CMDB;
- a generic log platform;
- a ticketing system;
- a graph database;
- an infrastructure-as-code engine;
- an autonomous remediation platform;
- a traditional Web administration dashboard.

It imports and relates operational evidence from those systems; it does not replace them.

## Target Operational Loop

```text
Observe
  -> Normalize and preserve evidence
  -> Project state, changes, topology, and incident context
  -> Retrieve and explain
  -> Decide
  -> Approve and act
  -> Observe the result
  -> Evaluate the outcome
  -> Learn reusable operational memory
```

The initial production-capable target ends at **retrieve and explain**. Decisions, actions, and outcomes will first be recorded as memory before any execution automation is introduced.

## Target Architecture

```text
CLI / MCP / importers
          |
     internal/core
 domain rules, commands, queries, ports
          |
 repositories + projectors + retrievers
          |
 SQLite durable evidence + rebuildable projections + FTS5
```

The V2 core must not depend on GORM, SQLite drivers, CLI packages, or MCP protocol types. Collectors must call core ingestion commands and must never write the database directly.

## Roadmap Summary

1. Freeze product contract, architecture, data model, evaluation, and repository memory.
2. Build the immutable evidence kernel and real core boundaries.
3. Add time-aware topology and derived change memory.
4. Prove observed truth with Prometheus and WireGuard ingestion.
5. Build incident context, hypotheses, decisions, actions, and outcomes.
6. Deliver evidence-citing MCP tools such as `get_resource_state`, `get_resource_timeline`, `trace_evidence`, `explain_resource`, and `build_incident_context`.
7. Consider controlled action execution only after read/explain quality is measured and stable.

See [docs/roadmap-v2.md](docs/roadmap-v2.md) for milestone exit criteria.

## Current Demo

The current commands and MCP server remain useful only as implementation experiments while V2 is built. They may be removed or changed without migration support.

The historical demo build still requires SQLite FTS5 and CGO:

```zsh
make test
make build
```

Do not deploy the demo as an operational source of truth.

## Restarting Work

A new maintainer or coding agent must:

1. fetch `refactor/memory-core-v2`;
2. read `docs/PROJECT_MEMORY.md` completely;
3. inspect the latest commits and ADRs;
4. resume the first incomplete item in the checkpoint;
5. update the checkpoint before ending a substantial work session.

This repository treats durable design memory as part of the product engineering process.
