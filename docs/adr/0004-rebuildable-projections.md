# ADR 0004: Rebuildable Versioned Projections

- Status: Accepted
- Date: 2026-08-02
- Depends on: ADR 0001, ADR 0002, ADR 0003

## Context

MemInfra must answer current-state, topology, change, search, and incident-context queries efficiently.

Computing every answer from all raw evidence at query time would be expensive and would duplicate selection policy across adapters. Making mutable projection rows the durable truth would erase history and make future policy changes unsafe.

Projection logic will evolve as source priority, freshness, conflict, topology, and incident semantics improve.

## Decision

MemInfra stores immutable evidence as durable truth and maintains versioned, rebuildable projections for query efficiency.

Initial projections:

- resource current state;
- relationship current and historical state;
- derived changes;
- incident state/context indexes;
- search documents and FTS5 index;
- optional deterministic context caches.

Each projection has:

- a stable name;
- an integer projector version;
- a workspace-scoped checkpoint;
- deterministic input ordering;
- an active/inactive build state;
- rebuild and verification behavior.

## Initial Consistency Model

For the local V2 kernel, projections required for read-after-write behavior are updated synchronously inside the evidence append transaction when practical.

A durable asynchronous projection worker is deferred until benchmarks demonstrate that synchronous work violates required ingestion throughput or latency.

The product contract does not depend on eventual consistency unless a later ADR explicitly introduces it.

## Rebuild Procedure

A safe rebuild:

1. creates or clears a fresh projection version/namespace;
2. records a rebuilding checkpoint;
3. replays evidence in stable semantic order;
4. applies retractions, supersession, validity, and policy deterministically;
5. periodically persists progress if resumable rebuilding is supported;
6. verifies completion and semantic invariants;
7. atomically activates the new version;
8. retains the previous active version until cleanup is safe.

An interruption must not leave queries reading a partial projection.

## Determinism

Given:

- the same durable evidence and links;
- the same projector version;
- the same policy configuration;
- the same fixed clock/query time where required;

a rebuild produces semantically equivalent projection output.

Generated projection-row IDs, physical row order, and timestamps that are not part of the contract may differ.

## Evidence References

Projection output retains supporting evidence IDs.

A state value, conflict, relationship, change, incident conclusion, or search document that cannot identify its source evidence is invalid.

## Projector Versioning

Increment projector version when semantic output can change, including:

- selection priority;
- freshness rules;
- conflict grouping;
- payload interpretation;
- relationship identity;
- change detection;
- search document shaping.

A schema migration alone does not necessarily change projector version; a semantic change does.

## Consequences

### Positive

- Query performance does not require sacrificing evidence history.
- Projection bugs and policy changes can be repaired by replay.
- Search indexes are explicitly disposable.
- Current state can expose evidence lineage.
- Adapter queries share one deterministic semantic layer.
- The project can compare projector versions before activation.

### Negative

- Additional storage is required.
- Projector version and checkpoint management add complexity.
- Rebuilds can be expensive and require operational tooling.
- Synchronous projection limits ingestion throughput until measured optimization.
- Determinism requires careful clock and ordering design.

## Alternatives Rejected

### Compute Everything at Query Time

Rejected for repeated cost, inconsistent policy implementation, and poor bounded-latency behavior.

### Treat Current-State Tables as Durable Truth

Rejected because history, late evidence, retraction, and semantic policy evolution become unsafe.

### Update Projections Asynchronously From an In-Memory Queue Immediately

Rejected initially because process failure could lose work and eventual consistency would complicate the first local product. A durable queue/worker may be introduced from measurements later.

### Store Only Search Documents

Rejected because FTS documents cannot represent structured state, conflicts, temporal topology, or evidence lineage reliably.

## Verification

The implementation must test:

- rebuild from empty projection state;
- interruption with previous version still readable;
- deterministic semantic comparison;
- out-of-order evidence replay;
- retraction and supersession;
- duplicate ingestion not duplicating projection changes;
- projector-version visibility in query results;
- raw evidence counts and hashes unchanged after rebuild;
- bounded memory on the reference dataset.
