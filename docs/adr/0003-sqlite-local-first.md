# ADR 0003: SQLite as the Local-First V2 Store

- Status: Accepted
- Date: 2026-08-02
- Depends on: ADR 0001, ADR 0002

## Context

MemInfra must be easy for an infrastructure engineer or AI coding agent to install, inspect, back up, move, and remove.

The first product does not require:

- multi-tenant service operation;
- distributed writes;
- high availability across nodes;
- a remote database control plane;
- unbounded metric or log ingestion.

The demo already proved the basic viability of SQLite and FTS5, but the V2 architecture must avoid coupling core semantics to a particular ORM or driver.

## Decision

The first complete V2 deployment uses one SQLite database as its durable local store.

SQLite stores:

- workspace and resource identities;
- immutable evidence and links;
- incidents and explicit links;
- migration metadata;
- rebuildable projections;
- FTS5 search projection.

The core defines repository and transaction ports. SQLite, GORM, `database/sql`, and driver details remain inside adapters.

## Driver Decision

The final V2 SQLite driver is selected through a focused implementation spike and contract tests.

Selection criteria:

- FTS5 behavior;
- transaction and foreign-key correctness;
- backup/integrity support;
- cancellation and busy-timeout behavior;
- supported build platforms;
- CGO/cross-compilation impact;
- performance on the reference dataset;
- maintenance and dependency risk.

The current CGO driver is not a compatibility requirement.

## Connection and Durability Policy

The adapter explicitly configures and tests:

- foreign keys;
- busy timeout;
- journal mode;
- synchronous/durability mode;
- connection count and writer behavior;
- context cancellation;
- integrity checking;
- backup behavior.

No product guarantee relies on undocumented driver defaults.

## Scale Boundary

Initial reference dataset:

- 10,000 resources;
- 1,000,000 evidence records;
- 100,000 active/historical relationships;
- representative projections and search documents.

The project measures append, projection, state, timeline, topology, search, rebuild, and database-size behavior.

A remote or distributed store is considered only after:

1. canonical scenarios are useful;
2. the reference benchmark is implemented;
3. a measured SQLite limitation affects a required deployment;
4. repository ports and semantic tests can be reused;
5. a superseding ADR defines operational cost and migration.

## Backup and Portability

V2 provides explicit commands or library operations for:

- consistent backup;
- integrity verification;
- restore;
- export/import with stable IDs;
- optional projection omission and rebuild;
- redacted export.

Copying a live database file without the documented backup mechanism is not the supported backup contract.

## Consequences

### Positive

- Minimal installation and operational burden.
- Inspectable single-file persistence.
- Good fit for local agents and developer tools.
- Transactional evidence and projections.
- Built-in FTS5 option.
- Easy deterministic integration testing.
- Storage abstraction remains behind core ports.

### Negative

- Single-writer behavior requires batching and measured projection design.
- Large/high-frequency raw metric ingestion is inappropriate.
- Multi-host concurrent access is not a first-class deployment.
- Driver/platform behavior needs explicit testing.
- Backup and WAL behavior must be documented.

## Alternatives Rejected

### PostgreSQL First

Rejected because it introduces service operation and installation complexity before local product value is proven.

### Embedded Key-Value Store

Rejected because V2 requires relational integrity, temporal queries, joins, migrations, and FTS-style retrieval.

### Vector Database First

Rejected because vectors do not replace durable evidence, temporal state, relational links, or deterministic lexical retrieval.

### Files/JSONL as the Primary Database

JSONL remains an import/export and fixture format, but is rejected as the query/projection database because transactional constraints and indexed temporal queries are required.

## Verification

- all persistence integration tests use real temporary SQLite databases;
- foreign keys and uniqueness constraints are tested;
- migrations run from empty and released V2 snapshots;
- projection rebuild operates within bounded memory;
- backup/restore and integrity checks are exercised;
- reference benchmarks record driver and pragma configuration;
- core packages contain no SQLite/ORM imports.
