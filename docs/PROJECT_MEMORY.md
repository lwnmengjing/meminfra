# MemInfra Project Memory

> **Authoritative project checkpoint.** Future maintainers and coding agents must read this file before changing the project.
>
> Last updated: 2026-08-02
> Active redesign branch: `refactor/memory-core-v2`
> Redesign base: `main` at `0426749be268a40754ff062341b2087a9cb04f9a`
> Current product stage: non-production demo; breaking redesign is explicitly allowed.

## 1. Why This File Exists

The existing repository is only a demonstration and has no production data or compatibility obligations. The first implementation proved that Go, SQLite, FTS5, CLI, and MCP can be connected, but it optimized the outer surfaces before proving the actual product.

This file freezes the intended product, architecture, development order, safety model, and validation criteria before more code is written. It is deliberately stored in the repository so a process restart, a new coding agent, or a long gap between sessions does not erase the design context.

This document is the primary source of truth until the individual sections are promoted into stable ADRs and focused design documents. When another document conflicts with this file, this file wins unless a later ADR explicitly supersedes it.

## 2. Product Definition

MemInfra is an **AI-native operational memory for infrastructure**.

It is a local-first system that records infrastructure evidence, preserves temporal and causal context, derives current state and changes, and exposes explainable operational context to AI agents and human operators.

MemInfra exists to help an agent answer questions such as:

1. What is the current known state of this resource?
2. What evidence supports that state, when was it observed, and how trustworthy is it?
3. What changed recently?
4. Which events, topology changes, actions, and incidents are correlated with the change?
5. Have we seen a similar incident before?
6. What decision was made, what action was taken, and what happened afterward?
7. Is an old conclusion still valid for the current topology and time window?

The first-class consumer is an AI agent such as Codex, Claude Code, OpenCode, Cursor, or another MCP client. Humans remain first-class operators through a CLI and inspectable files, but MemInfra is not designed around a traditional dashboard.

## 3. Product Thesis

Ordinary infrastructure systems usually expose one incomplete slice:

- monitoring systems expose metrics and alerts;
- CMDBs expose intended inventory;
- cloud APIs expose provider state;
- ticket systems expose human summaries;
- logs expose high-volume events;
- automation systems expose actions;
- wikis expose retrospective knowledge.

An AI agent cannot safely reason from any one of these in isolation. It needs a durable memory that preserves:

- **identity**: what resource the information refers to;
- **time**: when the fact was observed and when it was valid;
- **provenance**: which source produced it;
- **confidence**: whether it is observed, asserted, inferred, confirmed, or retracted;
- **causality**: what event or action caused the next record;
- **topology**: how the resource relates to other resources;
- **outcomes**: what happened after a decision or action;
- **traceability**: which evidence supports every explanation.

MemInfra is that memory layer. It does not replace the source systems. It imports, normalizes, relates, projects, retrieves, and explains their operational evidence.

## 4. Core Principles

### 4.1 Memory First

The durable evidence and its semantics are the product. CLI commands, MCP tools, FTS indexes, and future HTTP APIs are adapters over the memory core.

### 4.2 Observed Truth

Current state must be derived from timestamped evidence. A resource row by itself is not truth. Cloud inventory, Prometheus samples, WireGuard state, SSH probes, operator notes, deployment events, and action outcomes are all evidence with different reliability and freshness.

### 4.3 Temporal by Default

Every meaningful record must distinguish at least:

- `observed_at`: when the source observed it;
- `ingested_at`: when MemInfra stored it;
- `valid_from` and optional `valid_to`: when a fact is considered valid.

Queries must accept or establish a time window. Stale facts must not silently masquerade as current state.

### 4.4 Provenance and Explainability

Every derived state, change, hypothesis, or answer must be traceable back to evidence identifiers. The system must prefer a qualified answer with evidence over a confident-looking answer with no provenance.

### 4.5 Append-Only Evidence

Raw evidence is immutable after ingestion. Corrections are represented by retraction, supersession, or newer evidence. Mutable tables are projections and indexes that can be rebuilt.

### 4.6 Deterministic Core Before LLM Reasoning

Normalization, deduplication, ordering, freshness, topology traversal, change detection, and evidence selection must be deterministic Go code. An LLM may summarize or rank context later, but it must not be the only mechanism preserving operational truth.

### 4.7 Local First

The first complete product runs as a single local process over SQLite. It must remain easy to inspect, back up, export, and delete. Centralized services, remote databases, and distributed coordination are later concerns.

### 4.8 Safe Before Autonomous

The initial system is read-oriented and record-oriented. Collectors are read-only. Action execution is not part of the first production-ready milestone. Future actuators require approval, dry-run support, idempotency, audit records, and post-action observation.

### 4.9 Rebuildable Projections

Current resource state, topology views, search documents, summaries, and similarity indexes are projections over durable records. They must be versioned and rebuildable.

### 4.10 Small Vertical Slices

A feature is not complete because a table and CRUD command exist. A valid slice includes evidence ingestion, persistence, projection, query, adapter exposure, tests, documentation, and an end-to-end scenario.

## 5. Explicit Non-Goals

The following are not near-term product goals:

- a Grafana replacement;
- a general monitoring platform;
- a full CMDB or asset-management suite;
- a general log storage system;
- a workflow/ticketing product;
- a generic graph database;
- an infrastructure-as-code engine;
- an autonomous remediation platform;
- a multi-tenant SaaS control plane;
- a traditional Web administration UI;
- storing secrets or private keys;
- storing high-cardinality raw metrics indefinitely.

A future feature that moves MemInfra toward one of these products must justify how it strengthens operational memory rather than duplicating an existing system.

## 6. Canonical Operational Loop

The long-term loop is:

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

The first implementation must complete the loop through **retrieve and explain**. Decisions, actions, and outcomes are recorded before execution automation is introduced.

## 7. Canonical Acceptance Scenarios

Architecture and APIs are judged against scenarios, not against the number of commands or tables.

### Scenario A: RTT Spike After a Route or Tunnel Change

Input evidence:

- resource inventory for two edge nodes;
- a WireGuard relationship between them;
- periodic RTT observations;
- a route or tunnel state change event;
- a subsequent RTT spike;
- an operator note or incident conclusion.

Required answers:

- show the current state of the affected node;
- show the timeline around the spike;
- identify the topology change that preceded the spike;
- cite the supporting evidence IDs;
- distinguish observed correlation from confirmed root cause;
- find a prior similar incident and its outcome.

### Scenario B: S3 Upload Failures for Large Objects

Input evidence:

- device or gateway resources;
- upload success/failure events with object size classes;
- network ASN or route changes;
- MTU/MSS probe observations;
- operator hypotheses and test actions;
- post-action success measurements.

Required answers:

- explain why small uploads succeed while large uploads fail;
- show supporting and contradicting evidence;
- show when the behavior began;
- show which resources and topology paths are affected;
- show the test action and whether the outcome improved.

### Scenario C: Resource State Drift

Input evidence:

- intended resource facts from configuration;
- observed facts from a collector;
- a deployment/change event;
- a later observation that conflicts with the intended state.

Required answers:

- display intended versus observed state;
- identify the first conflicting evidence;
- mark the current state as uncertain or drifted rather than silently choosing one source;
- cite freshness and source priority rules.

The initial evaluation dataset must contain deterministic fixtures for all three scenarios.

## 8. Domain Model V2

The current demo schema is not a compatibility constraint. The V2 model may replace it completely.

### 8.1 Resource

A resource is a stable operational identity, not the current truth about that identity.

Required fields:

- `id`: stable generated identifier;
- `resource_key`: human-readable stable key, unique within a workspace;
- `kind`: server, service, endpoint, bucket, tunnel, cluster, node, device group, and so on;
- `display_name`;
- `identity_json`: identifiers required to correlate source records;
- `created_at`;
- `retired_at` optional.

Provider, region, IP addresses, hostnames, roles, and labels are normally represented as evidence and projected state. A small identity envelope may be stored directly only when it is required to correlate incoming evidence.

### 8.2 Evidence Envelope

Evidence is the immutable foundation.

Required fields:

- `id`: stable sortable identifier;
- `workspace_id`: reserved for future separation; one local default workspace initially;
- `subject_key`: resource key or operational subject;
- `kind`: observation, event, fact, relationship, operator_note, decision, action, outcome, or retraction;
- `schema`: versioned payload schema name;
- `payload_json`: validated payload;
- `source_type`: manual, prometheus, wireguard, ssh, cloud, deploy, mcp, cli, importer, or another adapter;
- `source_ref`: stable reference to the originating alert, sample, file, command, ticket, or API object;
- `collector_id` optional;
- `observed_at`;
- `ingested_at`;
- `valid_from`;
- `valid_to` optional;
- `confidence`: normalized value plus semantic status;
- `status`: observed, asserted, inferred, confirmed, contradicted, retracted;
- `correlation_id` optional;
- `causation_id` optional;
- `supersedes_id` optional;
- `dedupe_key`: unique idempotency key for a source record;
- `content_hash`;
- `tags_json`;
- `metadata_json`.

Evidence records are never updated in place except for narrowly defined storage repair procedures. A semantic correction creates new evidence.

### 8.3 Typed Evidence Payloads

The common envelope is generic, while payload schemas remain typed and versioned.

Initial schemas:

- `observation.numeric.v1`
  - metric, value, unit, dimensions;
- `event.v1`
  - event_type, severity, message, data;
- `fact.v1`
  - predicate, value, value_type;
- `relationship.v1`
  - src_resource_key, dst_resource_key, relation_type, attributes;
- `operator_note.v1`
  - title, body, author_ref;
- `decision.v1`
  - question, selected_option, alternatives, rationale;
- `action.v1`
  - action_type, target, parameters, approval_ref, idempotency_key;
- `outcome.v1`
  - action_evidence_id, success, summary, measurements;
- `retraction.v1`
  - target_evidence_id, reason.

Go structs and JSON Schema fixtures must define these contracts. Unknown schemas can be retained but are not projected until a projector supports them.

### 8.4 Resource State Projection

`resource_state` is a rebuildable projection containing:

- resource ID/key;
- projection version;
- state JSON grouped by predicates;
- evidence IDs supporting each value;
- freshness timestamps;
- conflict markers;
- computed health/status fields only where deterministic rules exist;
- updated_at.

State selection rules must account for source priority, observed time, validity, status, and retractions. Conflicting valid evidence is represented explicitly.

### 8.5 Relationship Projection

`relationship_state` is a time-aware projection:

- source resource;
- destination resource;
- relation type;
- active/inactive/uncertain state;
- valid interval;
- attributes;
- supporting evidence IDs;
- projection version.

One-hop queries are the first supported traversal. Arbitrary graph query languages are not required.

### 8.6 Change Records

Changes are derived records that compare projection versions or consecutive evidence:

- subject;
- change type;
- before JSON;
- after JSON;
- detected_at;
- effective_at;
- supporting evidence IDs;
- correlation ID;
- severity and confidence where deterministic.

Changes can be rebuilt and must not replace original evidence.

### 8.7 Incident Memory

An incident is a durable case that groups evidence and learning.

Required concepts:

- incident identity, title, status, severity, opened_at, closed_at;
- affected resources;
- linked evidence;
- timeline entries;
- hypotheses with status and supporting/contradicting evidence;
- decisions;
- actions;
- outcomes;
- final conclusion;
- reusable lessons and applicability constraints.

A root cause must not be represented as plain text alone. It is a conclusion with a status and evidence links.

### 8.8 Search and Context Documents

FTS5 remains a lexical retrieval index, not the source of truth.

Search documents are rebuildable projections with:

- document type;
- stable subject/ref;
- title/body/tags;
- source evidence IDs;
- valid time range;
- projection version;
- updated_at.

Search results must always include stable references that can be traced back to evidence.

## 9. Target Architecture

```text
cmd/meminfra                 cmd/meminfra-mcp
      |                              |
      +---------- adapters ----------+
                     |
              internal/core
     commands, queries, policies, ports
                     |
      +--------------+--------------+
      |              |              |
 repositories    projectors      retrievers
      |              |              |
 internal/adapters/sqlite    internal/adapters/fts5
      |                              |
             SQLite durable store

collectors/importers -> core ingestion commands
```

### 9.1 `internal/core`

Owns:

- domain types and invariants;
- command/query DTOs independent of GORM;
- use cases;
- evidence validation and normalization policy;
- deduplication semantics;
- time/freshness rules;
- source priority and conflict policy;
- projection orchestration;
- repository, transaction, clock, ID, and retriever interfaces;
- typed application errors.

It must not import GORM, SQLite drivers, CLI packages, or MCP protocol types.

### 9.2 `internal/adapters/sqlite`

Owns:

- schema and migrations;
- GORM or `database/sql` persistence models;
- transactions;
- repository implementations;
- projection storage;
- SQLite pragmas and connection management;
- backup/export primitives.

Persistence structs are not returned across the adapter boundary.

### 9.3 `internal/adapters/fts5`

Owns:

- FTS schema;
- safe query construction;
- document writes and rebuild;
- ranking and filters;
- mapping results back to stable references.

### 9.4 `internal/projection`

Owns deterministic projectors for:

- resource current state;
- relationships/topology;
- changes;
- memory/search documents;
- incident context summaries where deterministic.

Projectors are versioned. A rebuild command can reconstruct projections from evidence.

### 9.5 `internal/ingest`

Owns source-specific normalization and cursor handling. Collectors call core commands rather than writing the database directly.

Initial adapters:

1. JSONL/manual evidence importer for deterministic fixtures and operator use;
2. Prometheus importer;
3. WireGuard topology/state importer.

SSH and cloud inventory follow only after the first two real collectors prove the model.

### 9.6 CLI Adapter

The CLI is for operators, scripts, fixtures, inspection, export, and repair. It must remain thin: parse flags, call core, render text/JSON.

### 9.7 MCP Adapter

MCP exposes memory capabilities, not database CRUD as the primary abstraction.

Initial read tools after V2 core exists:

- `search_memory`;
- `get_resource_state`;
- `get_resource_timeline`;
- `get_recent_changes`;
- `trace_evidence`;
- `query_topology`;
- `explain_resource`;
- `build_incident_context`;
- `find_similar_incidents`.

Early write tools are limited to recording operator notes and incident conclusions. Infrastructure execution is excluded.

## 10. Query Contracts

### 10.1 Resource State

A resource-state result must include:

- selected state values;
- freshness;
- confidence/status;
- conflicts;
- supporting evidence IDs;
- last change;
- relevant active relationships.

### 10.2 Timeline

A timeline query accepts:

- subject/resource;
- explicit start/end or a bounded default window;
- evidence kinds;
- correlation filters;
- limit/cursor.

Results are ordered by effective/observed time with ingestion time retained to expose delayed records.

### 10.3 Explanation

`explain_resource` is deterministic context assembly, optionally followed by summarization. It returns:

- concise state summary;
- recent changes;
- supporting evidence;
- contradicting evidence;
- relevant topology;
- open incidents;
- similar resolved incidents;
- hypotheses with confidence and status;
- data gaps/staleness warnings.

No explanation is valid without evidence references.

### 10.4 Similar Incidents

The first implementation uses deterministic filters plus FTS5. Vector retrieval is postponed until a representative evaluation set demonstrates a lexical retrieval gap.

## 11. Storage and Migration Strategy

Because the current project is an unused demo:

- backward compatibility with the V1 SQLite schema is not required;
- existing CLI and MCP contracts may break;
- no V1-to-V2 production migration is required;
- the repository history preserves the demo;
- V2 starts with a new schema version and may use a new default database file during development.

Migration rules:

- numbered SQL migrations or an explicit migration package;
- schema version table;
- migrations tested from an empty database and every released V2 schema;
- projection schemas carry projector versions;
- rebuild commands must be idempotent;
- raw evidence survives projection rebuilds;
- destructive repair commands require explicit flags and backups.

## 12. Security and Safety Model

- Never store credentials, private keys, tokens, or raw secret values.
- Store secret references only when necessary.
- Collectors default to read-only access.
- Raw payloads must support field redaction before persistence.
- MCP write capabilities are disabled by default until authorization exists.
- Future action execution requires:
  - explicit approval;
  - dry-run when supported;
  - idempotency key;
  - target scope validation;
  - action evidence before execution;
  - outcome evidence after execution;
  - immutable audit linkage.
- Local database permissions must be documented and checked by installers.
- Export commands need redaction options.
- Inputs have size limits and strict JSON schema validation.

## 13. Development Roadmap

### Milestone 0: Design and Repository Reset

Goal: freeze the product contract and make the repository resumable.

Deliverables:

- this project memory committed and pushed;
- focused product contract, architecture, data model, and ADR documents derived from it;
- README rewritten to describe V2 direction and clearly label current code as demo/legacy;
- roadmap and issue/milestone structure;
- CI baseline verified;
- unresolved PR review defects catalogued;
- no new product feature code before this checkpoint exists.

Exit criteria:

- another engineer can resume solely from repository documents;
- design conflicts in existing docs are removed or marked obsolete;
- V2 implementation order is unambiguous.

### Milestone 1: Memory Kernel V2

Goal: implement immutable evidence, real core boundaries, and rebuildable state.

Deliverables:

- domain IDs, timestamps, source/provenance, confidence/status types;
- typed evidence schemas and JSON validation;
- core command/query interfaces independent of persistence;
- SQLite V2 migrations;
- append-evidence transaction with dedupe keys;
- resource identity repository;
- resource-state projector;
- evidence trace query;
- timeline query;
- JSONL importer;
- projection rebuild command;
- CLI commands for ingest, state, timeline, trace, and rebuild;
- tests and deterministic fixtures.

Exit criteria:

- duplicate ingestion is idempotent;
- out-of-order evidence preserves observed time and produces correct state;
- conflicting evidence is surfaced;
- retraction/supersession changes projections without mutating raw evidence;
- projections rebuild to byte-equivalent semantic results;
- Scenario C passes end to end.

### Milestone 2: Topology and Change Memory

Goal: represent time-aware relationships and explain changes.

Deliverables:

- relationship evidence schema and projector;
- one-hop topology query with time window;
- change detector and change projection;
- recent-change query;
- correlation and causation links;
- MCP read tools for state, timeline, evidence, changes, and topology;
- topology/change fixtures.

Exit criteria:

- active and historical relationships are queryable;
- a topology update creates an explainable change;
- Scenario A can identify the relationship/event preceding the RTT spike;
- every returned change cites evidence.

### Milestone 3: Real Observed-Truth Ingestion

Goal: prove the memory model with real sources.

Deliverables:

- Prometheus importer with source cursor/checkpoint and bounded query windows;
- alert/event ingestion where available;
- metric downsampling/retention policy for memory-relevant observations;
- WireGuard topology/state importer;
- collector health and ingestion statistics;
- idempotent retry behavior;
- source-specific fixtures and integration tests.

Exit criteria:

- collectors can run repeatedly without duplicates;
- delayed records are represented correctly;
- source outages do not silently mark stale state as current;
- a real local environment can populate resource state and topology.

### Milestone 4: Incident Context and Operational Learning

Goal: preserve reasoning and outcomes, not only observations.

Deliverables:

- incidents, affected resources, evidence links, and timelines;
- hypotheses with supporting and contradicting evidence;
- decisions and alternatives;
- action records and outcome records without automatic execution;
- deterministic incident-context builder;
- similar-incident retrieval using filters plus FTS5;
- MCP tools for incident context and similarity;
- operator-note and conclusion recording tools with audit source.

Exit criteria:

- root cause cannot be marked confirmed without linked evidence;
- action outcome links to pre/post observations;
- Scenario B records a hypothesis, test action, outcome, and reusable lesson;
- similar-incident retrieval is evaluated against golden fixtures.

### Milestone 5: Explainable Agent Experience

Goal: make MemInfra useful in a real coding/operations-agent session.

Deliverables:

- `explain_resource` context assembly;
- freshness and data-gap warnings;
- structured MCP outputs with stable schemas;
- optional pluggable summarizer that receives only selected evidence;
- prompt-independent evaluation harness;
- installation and backup/restore experience;
- version reporting and protocol conformance fixes.

Exit criteria:

- an MCP client answers canonical questions with evidence citations;
- deterministic structured context is useful without an LLM;
- optional summaries do not invent evidence IDs;
- all three canonical scenarios meet evaluation thresholds.

### Milestone 6: Controlled Action Loop (Deferred)

Goal: introduce safe action orchestration only after read/explain quality is proven.

Deliverables may include:

- actuator interfaces;
- approval records;
- policy engine;
- dry-run and idempotency;
- limited adapters;
- post-action observation and outcome evaluation.

This milestone must not begin merely because MCP write tools are easy to add.

## 14. Testing and Evaluation Strategy

### 14.1 Unit Tests

Cover:

- domain validation;
- time and freshness rules;
- dedupe keys;
- source priority;
- conflict handling;
- retractions/supersession;
- projector determinism;
- query limits and pagination;
- JSON schema validation.

### 14.2 SQLite Integration Tests

Each repository and projector runs against a temporary real SQLite database with FTS5 enabled. Mocks are not sufficient for SQL, constraints, transactions, or search behavior.

### 14.3 Migration Tests

- empty database to latest;
- each released V2 schema to latest;
- failed migration rollback;
- projection rebuild after migration;
- backup before destructive repair.

### 14.4 Collector Contract Tests

Recorded source fixtures test normalization, deduplication, cursor progression, retry, delayed evidence, and redaction.

### 14.5 End-to-End Scenario Tests

The three canonical scenarios are checked into `testdata/scenarios`. Each scenario includes input evidence, expected state, expected timeline, expected evidence links, and expected incident context.

### 14.6 Agent Evaluation

Evaluate structured outputs on:

- evidence citation correctness;
- temporal correctness;
- stale-data warnings;
- conflict disclosure;
- root-cause status correctness;
- similar-incident precision;
- reproducibility from the same database snapshot.

### 14.7 Performance Baseline

Initial reference dataset:

- 10,000 resources;
- 1,000,000 evidence records;
- 100,000 active/historical relationships;
- representative FTS documents.

Provisional targets on a documented developer machine:

- append and project at least 200 evidence records/second for batch imports;
- resource-state p95 below 100 ms;
- 24-hour resource timeline p95 below 250 ms;
- explanation context assembly p95 below 500 ms excluding optional LLM summarization;
- deterministic projection rebuild with bounded memory.

Targets may be revised only with benchmark evidence.

## 15. CLI Direction

Proposed V2 command families:

```text
meminfra init
meminfra ingest jsonl
meminfra evidence get
meminfra evidence trace
meminfra resource get
meminfra resource state
meminfra resource timeline
meminfra change list
meminfra topology query
meminfra incident create
meminfra incident link-evidence
meminfra incident add-hypothesis
meminfra incident record-decision
meminfra incident record-action
meminfra incident record-outcome
meminfra search
meminfra projection rebuild
meminfra db backup
meminfra db verify
```

The exact names can change during implementation, but commands must represent memory capabilities rather than mirror every database table.

## 16. MCP Direction

The MCP adapter must:

- implement the negotiated protocol correctly;
- return stable snake_case structured content;
- report a build-derived server version;
- distinguish tool/application errors from JSON-RPC protocol errors;
- include evidence references in memory answers;
- keep read tools enabled by default;
- gate write tools explicitly;
- never execute infrastructure changes in early milestones.

Existing demo MCP code is not a compatibility contract.

## 17. Dependency Policy

Prefer the Go standard library and a small dependency set.

Allowed foundational dependencies must have a clear purpose, for example:

- SQLite driver;
- JSON Schema validation if implementing it correctly in-house would be wasteful;
- stable sortable ID generation;
- MCP protocol library only if it reduces conformance risk without owning business behavior.

Do not add vector databases, message queues, Web frameworks, or distributed systems before an acceptance scenario requires them.

## 18. Documentation Structure to Create

This checkpoint will later be decomposed into:

```text
docs/product-contract.md
docs/architecture.md
docs/data-model.md
docs/roadmap.md
docs/evaluation.md
docs/operations.md
docs/adr/0001-memory-first-redesign.md
docs/adr/0002-append-only-evidence.md
docs/adr/0003-sqlite-local-first.md
docs/adr/0004-rebuildable-projections.md
```

`docs/meminfra-progress.md` is legacy session memory and must not remain an authoritative roadmap. Useful historical facts can be retained, but contradictory current-state claims must be removed or archived.

## 19. Contribution and Change Rules

1. Read this file and applicable ADRs before implementation.
2. Work in small, reviewable commits.
3. Do not add a new adapter before a core interface/use case exists.
4. Do not let `internal/core` import persistence or protocol packages.
5. Do not let collectors write SQLite directly.
6. Do not mutate raw evidence.
7. Every explanation/query returning a conclusion must return evidence references.
8. Every new feature must extend at least one canonical scenario or add a justified new scenario.
9. Update tests and documentation in the same change.
10. Record validation commands and actual results; do not claim unrun checks.
11. Update the checkpoint section below before ending a substantial work session.
12. Preserve user data and Git history. Do not force-push shared branches.

## 20. Restart and Resume Procedure

After a restart or when a new agent takes over:

1. Fetch the repository and branch `refactor/memory-core-v2`.
2. Read `docs/PROJECT_MEMORY.md` completely.
3. Read the latest ADRs and the last five commits.
4. Check the `Current Checkpoint` section below.
5. Inspect open pull requests and unresolved review threads.
6. Run the validation commands recorded by the last checkpoint when the environment permits.
7. Continue the first incomplete item in `Immediate Execution Plan`.
8. Before ending, update this file with actual status, commit SHA, validation, risks, and next action.

Never infer progress from planned checkboxes alone. Confirm it in the repository.

## 21. Immediate Execution Plan

The implementation order after this checkpoint commit is:

- [ ] Derive focused product, architecture, data-model, roadmap, evaluation, and ADR documents from this file.
- [ ] Rewrite README to state that the current implementation is a demo and V2 is under redesign.
- [ ] Mark `docs/meminfra-progress.md` as historical/non-authoritative or replace it with a concise current status.
- [ ] Inspect and catalogue current CI/build status and unresolved PR review defects.
- [ ] Create a V2 package skeleton and dependency-boundary tests.
- [ ] Implement domain identifiers, time/provenance/status types, evidence envelope, and typed payload validation.
- [ ] Implement SQLite V2 schema and append-evidence repository.
- [ ] Implement JSONL ingestion and evidence trace/timeline queries.
- [ ] Implement resource-state projection and rebuild.
- [ ] Add Scenario C end-to-end fixture and make it pass.
- [ ] Continue with topology/change memory only after Milestone 1 exits cleanly.

## 22. Current Checkpoint

Status as of this document creation:

- The existing repository is an unused demo.
- `main` contains the original SQLite/GORM/FTS5 CLI and read-oriented MCP demo plus dependency updates.
- The previous `internal/core` is largely a forwarding wrapper over `internal/store`; it is not the target V2 core.
- The previous store mixes validation, policy, persistence, projection, and indexing.
- Existing entity CRUD, installer, CI, and MCP code may be reused selectively, but none is a compatibility contract.
- No V2 code has been implemented yet.
- No current database contains production data that must be migrated.
- This redesign intentionally prioritizes evidence semantics and vertical acceptance scenarios over adding more CRUD surfaces.

First next action after this commit: split the design into focused documents and update the public README without changing product code.

## 23. Definition of Success

MemInfra V2 succeeds when an operator can ingest real infrastructure evidence locally and an AI agent can answer the canonical questions with:

- correct time context;
- explicit provenance;
- supporting and contradicting evidence;
- topology and recent-change context;
- incident/action/outcome history;
- honest uncertainty and stale-data warnings;
- deterministic, inspectable structured output.

Success is not measured by the number of tables, commands, protocols, collectors, or dashboards. It is measured by whether the memory allows an agent to reason more accurately and safely about infrastructure than it could from disconnected source systems alone.
