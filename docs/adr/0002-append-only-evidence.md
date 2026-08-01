# ADR 0002: Append-Only Evidence as Durable Truth

- Status: Accepted
- Date: 2026-08-02
- Depends on: ADR 0001

## Context

Infrastructure information changes, arrives late, conflicts across sources, and is sometimes corrected after ingestion.

A mutable resource row cannot answer:

- what was previously believed;
- which source supplied a value;
- when the source observed it;
- whether a later record corrected or merely contradicted it;
- which evidence supported an incident conclusion;
- what changed before and after an action.

Traditional update-in-place CRUD would erase operational history or require ad hoc audit tables for every entity.

## Decision

MemInfra stores source and operator records as immutable evidence envelopes.

Normal application code may append evidence and evidence links, but may not update or delete existing evidence rows.

Corrections use explicit new records and links:

- `supersedes`: replaces a prior claim for a defined semantic scope;
- `retracts`: removes a target from active projections while preserving history;
- `contradicts`: preserves incompatible active evidence;
- `supports`: links evidence to a hypothesis or conclusion;
- `outcome_of`: links measured outcome to an action.

Current state, topology, changes, incident context, and search documents are projections over this evidence.

## Evidence Requirements

Every record includes:

- stable evidence ID;
- subject;
- kind and versioned payload schema;
- source type and source reference;
- observed, ingestion, and validity time;
- status and confidence;
- dedupe key and content hash;
- optional correlation, causation, and supersession references;
- tags and metadata with size limits.

## Idempotency

Ingestion is at-least-once safe.

A workspace-scoped unique dedupe key ensures that retrying the same source record returns the existing evidence ID rather than creating a duplicate.

Payload equality alone does not collapse evidence from different sources because provenance is meaningful.

## Retention

Append-only does not mean MemInfra stores every raw metric sample forever.

Collectors decide which records are memory-relevant through explicit policies such as:

- bounded periodic summaries;
- state changes;
- alerts/anomalies;
- incident windows;
- operator-selected evidence.

Once accepted as evidence, a record remains durable unless an explicit administrative retention policy and export/backup process is introduced by a later ADR.

## Exceptions

Direct evidence mutation is permitted only for narrowly defined storage repair where:

- corruption or implementation defect is proven;
- a backup exists;
- the operation is explicit and audited;
- semantic corrections are still represented by new evidence when possible.

No normal repository interface exposes evidence update/delete methods.

## Consequences

### Positive

- Full operational history is preserved.
- Late and conflicting records can be represented honestly.
- Projection logic can evolve and rebuild.
- Incident reasoning and action outcomes remain traceable.
- Explanations can cite stable records.
- Import retry is safe.

### Negative

- Storage grows over time.
- Corrections require link-resolution logic.
- State queries require projections or temporal evaluation.
- Projector bugs can produce incorrect derived state until rebuilt.
- Administrative repair is more deliberate than editing a row.

## Alternatives Rejected

### Mutable Entity Tables With Updated-At

Rejected because previous values and source conflict are lost.

### Generic Audit Log Alongside Mutable State

Rejected as the primary model because audit logs often become incomplete secondary data while application behavior continues to trust mutable rows.

### Event Sourcing for Every Internal Operation

Partially related but rejected as terminology and scope. MemInfra records operational evidence, not every internal software event. The evidence model is domain-oriented and source-oriented.

## Verification

The implementation must prove:

- repository interfaces do not expose ordinary update/delete;
- duplicate concurrent append resolves through a database constraint;
- retraction changes projections without changing the target row;
- out-of-order evidence preserves chronology;
- projection rebuild leaves evidence hashes and counts unchanged;
- every derived result includes evidence references.
