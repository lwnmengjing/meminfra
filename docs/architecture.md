# MemInfra V2 Target Architecture

Last updated: 2026-08-02
Status: accepted target architecture for `refactor/memory-core-v2`

## 1. Architectural Goal

MemInfra is an evidence-based operational memory, not a CRUD application around infrastructure tables.

The architecture must make these statements true:

- immutable evidence is the durable source of truth;
- current state, topology, changes, and search documents are rebuildable projections;
- domain and application policy live in a persistence-independent core;
- CLI, MCP, collectors, SQLite, and FTS5 are adapters;
- every explanation can trace its conclusions to evidence identifiers;
- time, provenance, conflict, retraction, and freshness are core semantics rather than optional metadata.

The existing demo is allowed to be replaced because it has no production compatibility requirement.

## 2. System Context

```text
Prometheus     WireGuard     Cloud/SSH     Operator/JSONL
     |              |            |               |
     +---------- source adapters / importers -----+
                            |
                     ingestion commands
                            |
+----------------------- internal/core -----------------------+
| domain invariants, commands, queries, policies, ports       |
| evidence validation, dedupe, time, provenance, conflict     |
+-----------+------------------+-------------------+-----------+
            |                  |                   |
      repositories        projectors          retrievers
            |                  |                   |
            +--------- SQLite transaction boundary -----------+
                               |
                  durable evidence and identities
                               |
             state/topology/change/search projections
                               |
                       CLI and MCP adapters
                               |
                      human and AI operators
```

## 3. Dependency Rule

Dependencies point inward toward `internal/core`.

```text
cmd/*
internal/adapters/*
internal/ingest/*
internal/projection/*
        -> internal/core
```

`internal/core` must not import:

- GORM;
- SQLite drivers;
- MCP protocol packages;
- CLI flag packages;
- Prometheus or cloud SDKs;
- concrete clocks, ID generators, or file systems.

Persistence models and protocol DTOs must not leak through core interfaces.

## 4. Proposed Package Layout

```text
cmd/
  meminfra/
  meminfra-mcp/

internal/
  core/
    domain/
    command/
    query/
    policy/
    port/

  adapters/
    sqlite/
      migration/
      repository/
      model/
    fts5/
    mcp/
    cli/

  projection/
    resource_state/
    relationship_state/
    change/
    search_document/

  ingest/
    jsonl/
    prometheus/
    wireguard/

  app/
    bootstrap/

testdata/
  scenarios/
    rtt-spike/
    s3-large-upload-failure/
    resource-drift/
```

The exact folder depth may be simplified while implementing, but the dependency boundaries are mandatory.

## 5. Core Responsibilities

### 5.1 Domain

Domain types represent semantics rather than database rows:

- `ResourceID`, `ResourceKey`, `EvidenceID`, `CorrelationID`;
- `EvidenceKind`, `EvidenceStatus`, `Confidence`;
- `ObservedTime`, `ValidityInterval`;
- resource identity;
- evidence envelope and typed payloads;
- state values and conflicts;
- relationships and changes;
- incidents, hypotheses, decisions, actions, and outcomes.

Constructors validate invariants. Invalid domain objects cannot be persisted through normal application paths.

### 5.2 Commands

Initial commands:

- register or resolve resource identity;
- append evidence;
- append evidence batch;
- retract or supersede evidence;
- rebuild projections;
- create incident;
- link evidence to incident;
- record hypothesis, decision, action, and outcome.

Commands return domain/application results, not storage models.

### 5.3 Queries

Initial queries:

- get evidence and evidence lineage;
- get resource current state;
- get resource timeline;
- list recent changes;
- query one-hop topology at a point or interval in time;
- search projected memory documents;
- build resource explanation context;
- build incident context;
- find similar incidents.

### 5.4 Policies

Core policies include:

- source priority;
- freshness thresholds;
- validity handling;
- conflict detection;
- retraction and supersession;
- dedupe-key construction rules;
- projection versioning;
- conclusion confirmation requirements;
- query limits and bounded default windows.

Policies must be explicit and testable. They must not be hidden inside SQL or CLI defaults.

### 5.5 Ports

Core defines interfaces for:

- transaction execution;
- resource repository;
- evidence repository;
- projection repository;
- search retriever;
- clock;
- ID generator;
- schema registry/validator;
- optional event notification after commit.

Adapters implement these ports.

## 6. Durable Data Versus Projections

### 6.1 Durable Records

Durable records include:

- workspace and resource identity;
- immutable evidence envelopes and payloads;
- incident case identity and explicit evidence links;
- append-only reasoning/action/outcome records;
- schema and migration metadata.

These records survive projection deletion and rebuild.

### 6.2 Rebuildable Projections

Rebuildable projections include:

- resource current state;
- relationship current/historical state;
- derived changes;
- FTS5/search documents;
- context caches;
- similarity features;
- collector statistics that can be recalculated.

Projection rows carry a projector version. A projector version change triggers rebuild or controlled migration.

## 7. Transaction Boundaries

### 7.1 Append Evidence

A successful append transaction must atomically:

1. resolve or validate the subject identity;
2. enforce dedupe uniqueness;
3. store immutable evidence;
4. update synchronous projections required for read-after-write consistency, or enqueue a durable local projection task;
5. store projection cursor/version state;
6. commit once.

For the initial local product, synchronous deterministic projections are preferred until measured throughput requires a durable asynchronous worker.

### 7.2 Batch Ingestion

Batch ingestion:

- validates each item before transaction entry when possible;
- uses bounded batches;
- reports inserted, duplicate, rejected, and failed counts separately;
- does not convert a duplicate into an error;
- preserves input source references;
- is safe to retry.

### 7.3 Projection Rebuild

Projection rebuild:

- reads immutable evidence in a stable order;
- writes to fresh versioned projection tables or a temporary namespace;
- verifies completion and cursor state;
- swaps/activates the new version atomically;
- leaves durable evidence unchanged;
- can be interrupted and restarted safely.

## 8. Evidence Ordering

Evidence ordering uses multiple timestamps deliberately:

- `observed_at` drives operational chronology;
- `valid_from`/`valid_to` drive fact applicability;
- `ingested_at` exposes delayed arrival;
- stable evidence ID resolves deterministic ties.

Out-of-order ingestion must not corrupt chronology. A late record may cause a projection correction and a newly derived change, but the original ingestion history remains visible.

## 9. Conflict and Freshness Handling

A state projector never selects a value based only on “latest row wins.”

Selection considers:

- evidence status;
- validity interval;
- observed time;
- source priority;
- confidence;
- retraction/supersession;
- freshness policy by predicate/source.

When two valid records cannot be safely reconciled, the projection records a conflict with all candidate evidence IDs.

A stale value may remain the last known value, but it must be marked stale and must not be reported as freshly observed.

## 10. SQLite Adapter

The SQLite adapter owns:

- numbered migrations;
- connection pragmas;
- one-writer connection policy where appropriate;
- transaction implementation;
- persistence structs;
- indexes and constraints;
- repository implementations;
- backup and integrity-check support.

Recommended pragmas are established and tested explicitly rather than relying on driver defaults, including foreign keys, busy timeout, and journal mode decisions.

GORM may be retained inside this adapter if it does not obscure migrations, append-only guarantees, or query performance. Core interfaces must allow replacing it with `database/sql` without product changes.

## 11. FTS5 Adapter

FTS5 is a lexical retrieval projection.

It stores documents that contain:

- stable document reference;
- document type;
- title/body/tags;
- validity interval;
- source evidence IDs;
- projector version.

It does not own evidence, resource state, incident conclusions, or ranking policy beyond lexical retrieval.

Raw FTS syntax is never accepted by default. Safe token/phrase query construction and explicit filters are required.

## 12. Ingestion Architecture

Collectors and importers normalize source data into typed evidence commands.

They must not:

- mutate projection tables;
- write SQLite directly;
- invent resource identity without correlation policy;
- discard source timestamps;
- hide duplicate or rejected counts;
- store secret values.

### 12.1 JSONL Importer

Implemented first because it supports deterministic fixtures, manual imports, replay, and debugging.

### 12.2 Prometheus Importer

Supports bounded time ranges, source cursors, retry, delayed samples, metric selection, and retention/downsampling policy. MemInfra does not become a raw time-series database.

### 12.3 WireGuard Importer

Normalizes peer and tunnel relationships, handshake freshness, endpoint changes, and state observations into evidence.

## 13. CLI Adapter

The CLI:

- parses flags;
- calls one core use case;
- renders text or stable JSON;
- maps typed errors to exit codes;
- does not implement validation or defaulting that belongs to core;
- does not depend on SQLite models.

Commands are capability-oriented, such as state, timeline, trace, changes, topology, ingest, and rebuild.

## 14. MCP Adapter

MCP exposes structured memory capabilities.

Requirements:

- protocol-conformant JSON-RPC behavior;
- negotiated protocol version;
- build-derived server version;
- stable snake_case structured content;
- bounded inputs and outputs;
- evidence references in every explanatory result;
- correct distinction between protocol errors and tool errors;
- read tools enabled by default;
- write tools gated explicitly;
- no infrastructure execution in early milestones.

The adapter calls the same core queries as CLI and cannot query persistence directly.

## 15. Explanation Pipeline

`explain_resource` is not a free-form LLM prompt over the database.

The deterministic pipeline:

1. resolve the resource and time window;
2. load current state and conflicts;
3. load recent changes;
4. load bounded timeline evidence;
5. load relevant one-hop topology;
6. load open incidents;
7. retrieve similar resolved incidents;
8. classify stale and missing data;
9. return structured context with evidence references;
10. optionally pass only that context to a summarizer.

An optional summarizer cannot add new evidence IDs or silently upgrade a hypothesis to confirmed root cause.

## 16. Error Model

Core exposes typed errors such as:

- invalid input;
- unsupported evidence schema;
- resource not found;
- evidence not found;
- duplicate evidence;
- conflict;
- stale precondition;
- projection unavailable;
- authorization/policy denied;
- internal storage failure.

Adapters translate them without parsing error strings.

List queries return empty collections for valid filters with no matches. Identity lookups return explicit not-found errors.

## 17. Observability of MemInfra

MemInfra records operational health about itself separately from imported infrastructure memory:

- ingestion counts and rejection reasons;
- collector cursor/freshness;
- projection lag/version;
- rebuild progress;
- SQLite size and integrity status;
- query latency;
- backup status.

Self-observability must not recursively flood the primary evidence store.

## 18. Security Boundaries

- raw credentials and private keys are prohibited;
- payload fields can be redacted before persistence;
- database file permissions are checked;
- collectors use least-privilege read access;
- future action ports are absent from default builds or disabled by policy;
- export supports redaction;
- write tools record actor/source references;
- all external inputs have strict size and schema limits.

## 19. Architecture Fitness Checks

CI must eventually enforce:

- `internal/core` has no imports from adapter/persistence/protocol packages;
- persistence structs do not escape repositories;
- every projector has deterministic rebuild tests;
- collectors call core interfaces;
- MCP and CLI contract tests share core fixtures;
- canonical scenarios pass end to end;
- migrations and SQLite integration tests use real temporary databases.

A lightweight dependency-boundary test should be introduced before substantial V2 code.

## 20. Evolution Rules

- Do not add HTTP before core query/command contracts are stable.
- Do not add a Web UI to compensate for weak query semantics.
- Do not add vector retrieval before lexical evaluation demonstrates a gap.
- Do not introduce a queue before synchronous projection throughput is measured.
- Do not introduce remote databases or multi-tenancy before the local product is useful.
- Do not add action execution before evidence-citing explanations and recorded outcomes are proven.
- Prefer one completed vertical scenario over many unfinished adapters.
