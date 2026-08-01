# ADR 0001: Memory-First V2 Redesign

- Status: Accepted
- Date: 2026-08-02
- Decision owners: MemInfra maintainers
- Supersedes: the demo implementation as an implicit product/architecture contract

## Context

The initial repository demonstrated:

- Go CLI commands;
- SQLite/GORM persistence;
- FTS5 search;
- resource, observation, event, incident, and relationship CRUD;
- a read-oriented MCP stdio server;
- installation and CI scaffolding.

The repository has no production deployment or data that requires compatibility.

The demo optimized adapters and entity CRUD before proving the original product thesis: an AI-native infrastructure memory that preserves observed truth, time, provenance, changes, topology, reasoning, actions, and outcomes.

The existing `internal/core` is mostly a forwarding wrapper over `internal/store`, while the store owns validation, policy, persistence, projection, and indexing. Continuing this shape would produce a lightweight CMDB/search tool rather than a reusable operational-memory core.

## Decision

MemInfra will undergo a breaking V2 redesign centered on immutable operational evidence and scenario-driven product behavior.

The authoritative product is defined as:

> A local-first system that records infrastructure evidence, preserves temporal and causal context, derives current state and changes, and exposes explainable operational context to AI agents and human operators.

The redesign adopts these rules:

1. Evidence semantics are implemented before expanding CLI, MCP, collector, or UI surfaces.
2. Current state, topology, changes, and search are projections over durable evidence.
3. Core commands, queries, policies, and ports are independent of persistence and protocols.
4. The V1 schema, CLI, and MCP contracts may be replaced without migration support.
5. Development is accepted through canonical end-to-end scenarios rather than feature count.
6. Action execution is deferred; early action records are memory only.
7. Repository documentation and `docs/PROJECT_MEMORY.md` are part of the engineering state and must remain resumable.

## Canonical Scenarios

The design must prove:

- resource state drift;
- RTT spike after route/tunnel change;
- large S3 upload failure with MTU/MSS hypotheses, test action, and measured outcome.

## Consequences

### Positive

- Product differentiation is explicit.
- Time, provenance, uncertainty, and evidence traceability become first-class.
- CLI and MCP can evolve without owning business behavior.
- Storage can change behind core ports if measured limits require it.
- Real collectors and incident learning can be added as vertical slices.
- A restart or maintainer handoff no longer loses project intent.

### Negative

- Existing demo commands and databases may stop working.
- Some already-written code will be deleted or substantially refactored.
- V2 requires more up-front domain and evaluation work than adding CRUD features.
- Projection and evidence semantics add complexity that a simple entity store would avoid.

### Risks

- A generic evidence envelope could become an unstructured JSON dump.
- Architecture layering could become ceremonial.
- Scenario documents could drift from executable tests.

Mitigations:

- typed versioned payloads;
- architecture fitness tests;
- real SQLite integration tests;
- checked-in golden scenario fixtures;
- mandatory evidence references in query results;
- milestone exit criteria.

## Alternatives Rejected

### Continue Incremental V1 Refactoring

Rejected because retaining V1 entity/store contracts would constrain the project around the wrong abstraction despite having no production compatibility need.

### Build MCP and Collectors First

Rejected because adapters would encode incomplete semantics and make later evidence/time changes more expensive.

### Build a Traditional CMDB With an AI Search Layer

Rejected because mutable current-state rows and prose search cannot preserve source conflict, time validity, or action outcomes reliably.

### Start With an LLM/Vector Memory System

Rejected because deterministic time, provenance, conflict, topology, and evidence selection must be correct before probabilistic summarization or vector retrieval adds value.

## Implementation Notes

The accepted roadmap begins with documentation/reset, then a memory kernel, topology/change projections, real collectors, incident learning, and explainable MCP tools.

Any proposal that reverses this decision requires a superseding ADR with evidence from canonical scenarios or benchmarks.
