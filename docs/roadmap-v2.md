# MemInfra V2 Development Roadmap

Last updated: 2026-08-02
Status: active execution plan for `refactor/memory-core-v2`

## 1. Execution Policy

The roadmap is ordered by dependency and product proof, not by implementation convenience.

Rules:

1. Complete one vertical slice before adding another surface.
2. Do not start a milestone until the previous milestone exit criteria are met or explicitly waived in an ADR.
3. Every slice includes domain behavior, persistence, query or projection, tests, documentation, and scenario coverage where applicable.
4. A table plus CRUD command is not a completed feature.
5. CLI, MCP, and collectors are adapters over core contracts.
6. Raw evidence is immutable; projections are rebuildable.
7. Validation results are recorded in `docs/PROJECT_MEMORY.md` before ending substantial work.
8. No force-push or history rewriting on shared branches.

## 2. Branch and Integration Strategy

Current design branch:

```text
refactor/memory-core-v2
```

This branch contains the authoritative redesign and may introduce breaking changes to the unused demo.

Recommended integration sequence:

1. merge the design/reset PR;
2. create milestone branches from updated `main`;
3. keep each implementation PR focused on one listed slice;
4. require canonical scenario and migration tests before merging core behavior;
5. tag the first explicit compatibility contract only after Milestone 5.

No V1 compatibility promise is implied by earlier tags or commands.

## 3. Milestone 0 — Design and Repository Reset

### Goal

Make the project direction durable, remove contradictory documentation, and establish a verifiable implementation starting point.

### Deliverables

- [x] authoritative `docs/PROJECT_MEMORY.md` committed and pushed;
- [x] README states that current code is a demo and V2 is a breaking redesign;
- [x] focused product contract;
- [x] V2 target architecture;
- [x] V2 logical data model;
- [x] executable roadmap;
- [ ] evaluation and canonical scenario specification;
- [ ] ADR for memory-first redesign;
- [ ] ADR for append-only evidence;
- [ ] ADR for SQLite local-first storage;
- [ ] ADR for rebuildable projections;
- [ ] old progress document marked historical/non-authoritative;
- [ ] current CI and unresolved review defects catalogued;
- [ ] design PR opened against `main`.

### Exit Criteria

- a new maintainer can resume from repository files alone;
- product scope and non-goals are explicit;
- data model and dependency direction are explicit;
- implementation order is unambiguous;
- no current document claims the demo is the production design;
- design changes are pushed and reviewable.

## 4. Milestone 1 — Memory Kernel V2

### Goal

Implement immutable evidence, stable resource identity, real core boundaries, trace/timeline queries, and a rebuildable resource-state projection.

### Slice 1.1 — Package Boundary Skeleton

Deliverables:

- create target package skeleton;
- define core ports without persistence imports;
- add an architecture/dependency fitness test;
- add typed application error package;
- add clock and ID generator ports;
- wire a minimal bootstrap composition root.

Acceptance:

- `internal/core` does not import GORM, SQLite, CLI, MCP, or collector packages;
- the fitness test fails if forbidden imports are introduced;
- demo code remains buildable or is clearly isolated during transition.

### Slice 1.2 — Domain Primitives

Deliverables:

- workspace, resource, evidence, incident, correlation, and change IDs;
- resource key validation;
- evidence kind/status/confidence types;
- observed/ingested/validity time types;
- source/provenance types;
- typed payload registry interface;
- constructors and validation tests.

Acceptance:

- invalid intervals, empty required fields, unsupported statuses, and invalid confidence cannot create domain evidence;
- IDs serialize consistently in JSON;
- timestamps round-trip in UTC.

### Slice 1.3 — Typed Evidence Schemas

Deliverables:

- `fact.v1`;
- `observation.numeric.v1`;
- `event.v1`;
- `relationship.v1`;
- `operator_note.v1`;
- `hypothesis.v1`;
- `decision.v1`;
- `action.v1`;
- `outcome.v1`;
- `retraction.v1`;
- schema examples and validators;
- canonical encoding and SHA-256 content hashing.

Acceptance:

- known payloads reject unknown/invalid required fields according to contract;
- canonical encoding yields stable hashes;
- payload size limits are enforced;
- unknown schemas are rejected unless explicit opaque-import mode is selected.

### Slice 1.4 — SQLite V2 Migrations

Deliverables:

- migration framework and schema-version table;
- workspaces, resources, aliases, evidence, evidence links, incidents, and link tables;
- projection checkpoint and resource-state tables;
- tested SQLite pragmas;
- empty-database migration tests;
- integrity constraints and indexes.

Acceptance:

- migrations are transactional where SQLite permits;
- foreign keys are enabled and tested;
- duplicate dedupe keys fail at the database constraint;
- evidence repository has no update/delete method in its normal interface;
- an empty database migrates deterministically to the latest version.

### Slice 1.5 — Append Evidence Use Case

Deliverables:

- register/resolve resource identity;
- append one evidence record;
- append bounded batch;
- deterministic dedupe-key policy;
- duplicate disposition returning the existing evidence ID;
- source and actor audit fields;
- transaction implementation;
- ingestion result counters.

Acceptance:

- retrying the same source record is idempotent;
- two records with identical payload but distinct source identities remain distinct;
- failed validation creates no partial rows;
- batch results separate inserted, duplicate, rejected, and failed records;
- concurrent duplicate inserts resolve safely through the unique constraint.

### Slice 1.6 — JSONL Importer and CLI Ingestion

Deliverables:

- versioned JSONL envelope format;
- streaming bounded-memory parser;
- dry-run validation mode;
- import summary;
- fixture replay;
- CLI adapter over append commands.

Acceptance:

- a fixture can be replayed repeatedly with no duplicate evidence;
- malformed lines report line numbers and do not obscure other valid records according to selected mode;
- large files are streamed rather than loaded entirely into memory;
- dry-run writes nothing.

### Slice 1.7 — Evidence Trace and Timeline Queries

Deliverables:

- evidence get;
- inbound/outbound evidence links;
- resource/subject timeline;
- explicit time windows;
- pagination cursor;
- CLI JSON/text output;
- query integration tests.

Acceptance:

- timeline order follows observed/effective time, not ingestion order;
- delayed records retain ingestion time;
- default windows are bounded and returned in metadata;
- evidence links are traversable without parsing payload text.

### Slice 1.8 — Resource-State Projection

Deliverables:

- fact selection policy;
- source priority configuration;
- freshness policy;
- validity and retraction handling;
- conflict representation;
- projection checkpointing;
- current-state query;
- derived change record on selected-state or conflict transition.

Acceptance:

- latest-row-wins is not the only rule;
- equal semantic values combine evidence references;
- unresolved conflicting valid values are exposed;
- stale values are marked stale;
- retraction removes a target from the projection without deleting it;
- late evidence recalculates correct state.

### Slice 1.9 — Projection Rebuild

Deliverables:

- versioned projector;
- stable evidence replay order;
- rebuild into a fresh projection version;
- activation/swap;
- resume/restart behavior;
- semantic comparison utility;
- CLI rebuild command.

Acceptance:

- rebuilding from the same evidence yields semantically equivalent state;
- interruption leaves the prior active projection usable;
- raw evidence is unchanged;
- projector version is visible in queries.

### Slice 1.10 — Scenario C: Resource Drift

Deliverables:

- checked-in intended and observed fact fixtures;
- conflict/drift expectation file;
- end-to-end ingestion, state, timeline, trace, and rebuild test;
- documented CLI/MCP-shaped output example.

Milestone Exit Criteria:

- Scenario C passes end to end;
- evidence append and rebuild are idempotent;
- conflict, freshness, retraction, and late arrival are tested;
- no core/persistence dependency violation exists;
- current-state results cite evidence IDs.

## 5. Milestone 2 — Topology and Change Memory

### Goal

Represent time-aware relationships and explain material changes around resources.

### Slice 2.1 — Relationship Projection

Deliverables:

- relationship evidence projector;
- deterministic relationship key;
- active/inactive/uncertain state;
- validity intervals;
- supporting evidence;
- historical and point-in-time one-hop query.

Acceptance:

- a relationship can become inactive without deleting history;
- point-in-time topology returns the state valid at that time;
- conflicting relationship evidence is represented explicitly.

### Slice 2.2 — Change Detection

Deliverables:

- state-value changes;
- conflict-opened/conflict-resolved changes;
- relationship changes;
- severity/confidence where deterministic;
- correlation and causation linkage;
- recent-change query.

Acceptance:

- every change cites source evidence;
- replay produces the same semantic changes;
- duplicate evidence does not duplicate changes.

### Slice 2.3 — Read-Oriented MCP V2

Initial tools:

- `get_resource_state`;
- `get_resource_timeline`;
- `get_recent_changes`;
- `trace_evidence`;
- `query_topology`;
- `search_memory` after the new FTS projection exists.

Acceptance:

- JSON-RPC and negotiated MCP lifecycle are conformant;
- outputs are stable snake_case structures;
- tool limits are bounded;
- server version comes from build information;
- protocol errors and tool errors are distinct.

### Slice 2.4 — Scenario A: RTT Spike

Deliverables:

- node, tunnel, route event, and RTT observation fixtures;
- expected topology before/after;
- expected timeline and change correlation;
- hypothesis remaining correlation-only until confirmed evidence exists.

Milestone Exit Criteria:

- Scenario A passes;
- current and historical one-hop topology are correct;
- changes and correlations cite evidence;
- MCP tools expose the same semantics as core queries.

## 6. Milestone 3 — Observed-Truth Collectors

### Goal

Prove the model with real read-only infrastructure sources.

### Slice 3.1 — Collector Runtime

Deliverables:

- collector identity and configuration;
- cursor/checkpoint persistence;
- bounded batch delivery to core;
- retries and backoff;
- ingestion statistics;
- last-success/freshness reporting;
- redaction hook;
- cancellation and clean shutdown.

Acceptance:

- collectors never write SQLite directly;
- repeated fetches are idempotent;
- cursor advances only after committed batches;
- collector outage is visible as stale source state.

### Slice 3.2 — Prometheus Importer

Deliverables:

- bounded range queries;
- selected metric mappings;
- alert/event mapping where available;
- aggregation/anomaly/incident-window policy;
- source-reference construction;
- recorded fixture tests.

Acceptance:

- MemInfra does not ingest unbounded raw time series;
- delayed samples retain their source timestamps;
- repeated overlapping windows do not duplicate evidence;
- configured metric mappings produce typed observations.

### Slice 3.3 — WireGuard Importer

Deliverables:

- peer/resource correlation;
- tunnel relationship evidence;
- endpoint and handshake observations;
- freshness/state policy;
- recorded fixture tests.

Acceptance:

- peer endpoint changes produce explainable topology/state changes;
- stale handshake is distinguished from confirmed tunnel failure;
- source secrets/private keys are never stored.

Milestone Exit Criteria:

- a real local environment populates evidence, resource state, and topology;
- collector freshness is visible;
- no direct storage coupling exists;
- replay and retry are safe.

## 7. Milestone 4 — Incident Context and Learning

### Goal

Preserve reasoning, decisions, actions, and measured outcomes as evidence-linked operational memory.

### Slice 4.1 — Incident Aggregate

Deliverables:

- incident creation/identity;
- affected-resource links;
- linked evidence roles;
- incident-state projection;
- timeline/context query.

### Slice 4.2 — Hypotheses and Conclusions

Deliverables:

- hypothesis evidence;
- supports/contradicts links;
- status transitions through append-only evidence;
- confirmation policy;
- explicit data gaps.

Acceptance:

- root cause cannot become confirmed without supporting evidence;
- contradicted and rejected hypotheses remain in history;
- conclusions are actor/source attributable.

### Slice 4.3 — Decisions, Actions, and Outcomes

Deliverables:

- decision evidence with alternatives and rationale;
- action record with approval/idempotency references;
- outcome record linked to action;
- pre/post observation comparison;
- no infrastructure execution.

Acceptance:

- action and outcome lineage is traceable;
- a success claim cites post-action measurements;
- recording an action never invokes an external system.

### Slice 4.4 — Similar Incident Retrieval

Deliverables:

- incident search projection;
- deterministic filters;
- FTS5 retrieval;
- applicability metadata;
- golden evaluation fixture.

Vector retrieval remains deferred until lexical evaluation demonstrates a measurable gap.

### Slice 4.5 — Scenario B: Large S3 Upload Failure

Deliverables:

- upload size-class events;
- ASN/path change;
- MTU/MSS observations;
- competing hypotheses;
- reversible test-action record;
- pre/post outcome evidence;
- reusable lesson with applicability constraints.

Milestone Exit Criteria:

- Scenario B passes;
- incident context cites supporting and contradicting evidence;
- action/outcome history is measurable and reusable;
- similar-incident retrieval meets the evaluation threshold.

## 8. Milestone 5 — Explainable Agent Experience

### Goal

Make the memory directly useful and safe in a real AI-agent operations session.

### Slice 5.1 — Deterministic Context Builder

Deliverables:

- `explain_resource` query;
- current state and conflicts;
- recent changes;
- bounded timeline;
- relevant one-hop topology;
- open incidents;
- similar resolved incidents;
- stale/missing-data warnings;
- evidence references throughout.

### Slice 5.2 — Optional Summarizer Port

Deliverables:

- provider-neutral summarizer interface;
- structured context input only;
- output validation;
- prohibition on invented evidence IDs;
- deterministic operation without a summarizer.

### Slice 5.3 — Final MCP Read Surface

Tools:

- `search_memory`;
- `get_resource_state`;
- `get_resource_timeline`;
- `get_recent_changes`;
- `trace_evidence`;
- `query_topology`;
- `explain_resource`;
- `build_incident_context`;
- `find_similar_incidents`.

Early write tools, if enabled explicitly:

- `record_operator_note`;
- `record_incident_conclusion`.

### Slice 5.4 — Operations and Release Readiness

Deliverables:

- database verify/integrity command;
- backup/restore/export/import;
- file permission checks;
- redacted export;
- version reporting;
- install prerequisites and supported platforms;
- release artifacts and checksums;
- schema compatibility policy;
- benchmark report.

Milestone Exit Criteria:

- all three canonical scenarios pass through core and MCP-shaped contracts;
- answers expose time, evidence, conflict, and data gaps;
- structured outputs are useful without an LLM;
- optional summaries do not change evidence semantics;
- backup/restore and projection rebuild are tested;
- the first V2 alpha compatibility contract can be tagged.

## 9. Milestone 6 — Controlled Action Loop (Deferred)

This milestone is intentionally blocked until Milestone 5 quality is demonstrated.

Potential deliverables:

- actuator ports;
- approval and authorization policy;
- target scope constraints;
- dry-run;
- idempotent execution;
- immutable action attempts;
- post-action collection and outcome evaluation;
- limited adapters.

Starting this milestone requires a new ADR and explicit safety review.

## 10. Continuous Validation Matrix

Every implementation PR runs, as applicable:

```text
gofmt / format check
go vet
unit tests
SQLite integration tests
migration tests
projection rebuild tests
architecture boundary tests
CLI contract tests
MCP protocol/contract tests
canonical scenario tests
race tests for core concurrency-sensitive code
benchmarks on benchmark-specific changes
CodeQL
installer/packaging checks when touched
```

Claims in PR descriptions must distinguish checks actually run from checks unavailable in the environment.

## 11. Review Checklist

A reviewer asks:

- Does this change improve a canonical question or scenario?
- Is durable information evidence or an accidental mutable projection?
- Are time and provenance preserved?
- Is ingestion idempotent?
- Can the projection be rebuilt?
- Are conflicts and stale state explicit?
- Does every conclusion cite evidence?
- Does core remain independent of adapters?
- Is the collector read-only and storage-independent?
- Is the MCP/CLI layer thin?
- Are secrets excluded?
- Is the action loop still record-only unless explicitly approved?

## 12. Known Risks and Mitigations

### Risk: Generic Evidence Becomes an Unqueryable JSON Dump

Mitigation:

- typed versioned payloads;
- indexed envelope fields;
- deterministic projectors;
- reject unknown schemas by default;
- scenario-driven query contracts.

### Risk: SQLite Is Blamed Before the Model Is Proven

Mitigation:

- benchmark the documented reference dataset;
- keep repository ports;
- optimize indexes and projection strategy first;
- introduce remote storage only with measured evidence.

### Risk: Core Becomes a Store Wrapper Again

Mitigation:

- architecture fitness test;
- no persistence type aliases;
- core-owned commands, queries, and policies;
- adapter DTO mapping.

### Risk: MemInfra Turns Into a Monitoring or CMDB Product

Mitigation:

- feature decision test in the product contract;
- memory-centered acceptance scenarios;
- no dashboard roadmap;
- collectors retain only memory-relevant observations.

### Risk: LLM Summaries Become the Source of Truth

Mitigation:

- deterministic structured context first;
- evidence IDs required;
- summarizer is optional and stateless;
- conclusions remain domain records with validation.

### Risk: Premature Autonomous Actions

Mitigation:

- no actuator port before Milestone 6;
- action payloads are record-only;
- explicit ADR and safety review required.

## 13. Immediate Next Changes

After the design checkpoint:

1. finish evaluation and ADR documents;
2. mark the old progress document historical;
3. catalogue current demo defects and CI state;
4. open the design PR;
5. begin Milestone 1 with the package-boundary skeleton and architecture fitness test;
6. do not implement collectors, new MCP tools, or incident features before the evidence kernel exists.
