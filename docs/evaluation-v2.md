# MemInfra V2 Evaluation Plan

Last updated: 2026-08-02
Status: required acceptance specification

## 1. Purpose

MemInfra is not validated by counting tables, commands, MCP tools, or collectors. It is validated by whether an agent can retrieve accurate, temporal, evidence-backed operational context.

This document defines deterministic scenario fixtures, required query outputs, semantic assertions, agent-facing evaluation, and performance baselines.

The evaluation suite must work without an LLM. Optional summarization is evaluated separately and cannot compensate for incorrect structured context.

## 2. Evaluation Layers

### 2.1 Domain Tests

Validate evidence and state semantics:

- identifiers;
- time intervals;
- status and confidence;
- typed payloads;
- dedupe;
- retraction and supersession;
- source priority;
- freshness;
- conflict handling.

### 2.2 Persistence and Projection Tests

Validate:

- SQLite constraints;
- transactions;
- migrations;
- stable replay order;
- projector determinism;
- projection rebuild;
- FTS document rebuild;
- query plans where performance matters.

### 2.3 Contract Tests

Run the same scenario expectations through:

- core queries;
- CLI JSON output;
- MCP structured content.

Adapters may format results differently for humans, but semantic fields and evidence references remain equivalent.

### 2.4 Canonical Scenario Tests

Three checked-in scenarios prove product behavior end to end.

### 2.5 Agent Evaluation

Evaluates whether structured context enables correct explanations and whether optional summaries preserve evidence semantics.

### 2.6 Performance Evaluation

Measures the documented reference dataset and prevents accidental algorithmic regressions.

## 3. Fixture Format

Each scenario is stored under:

```text
testdata/scenarios/<scenario>/
  manifest.json
  resources.jsonl
  evidence.jsonl
  expected/
    state.json
    timeline.json
    changes.json
    topology.json
    incident-context.json
    search.json
```

`manifest.json` defines:

```json
{
  "scenario": "resource-drift",
  "schema_version": 1,
  "clock": "2026-08-02T12:00:00Z",
  "default_workspace": "local",
  "queries": [],
  "notes": ""
}
```

Fixtures use stable explicit IDs where deterministic comparison benefits. Tests must not depend on wall-clock time.

## 4. Common Assertions

Every scenario checks the following where applicable.

### 4.1 Evidence Integrity

- all records retain source references;
- observed and ingestion times remain distinct;
- duplicate replay returns existing evidence IDs;
- content hashes and dedupe keys are stable;
- invalid records are rejected without partial writes.

### 4.2 Temporal Correctness

- timeline is ordered by effective/observed time;
- delayed ingestion is visible;
- point-in-time queries exclude evidence outside validity;
- stale state is marked;
- default query windows are reported.

### 4.3 Traceability

- selected state values include supporting evidence IDs;
- changes include source evidence IDs;
- topology edges include relationship evidence IDs;
- incident conclusions include supporting and contradicting evidence;
- search results link to stable subject/evidence references.

### 4.4 Uncertainty

- unresolved conflicts are not hidden;
- correlation is not reported as confirmed causation;
- missing evidence produces a data-gap warning;
- contradicted hypotheses remain visible;
- stale data is not described as current observation.

### 4.5 Rebuildability

- dropping projections and rebuilding yields semantically equivalent outputs;
- raw evidence row count and hashes do not change;
- duplicate replay after rebuild remains idempotent.

## 5. Scenario C — Resource State Drift

Implemented first because it proves the evidence kernel and state projector without requiring topology or incidents.

### 5.1 Input

Resources:

- `node/frankfurt-01`.

Evidence sequence:

1. intended fact: `network.ipv4 = 192.0.2.10` from configuration;
2. observed fact: `network.ipv4 = 192.0.2.10` from cloud inventory;
3. deployment event;
4. later observed fact: `network.ipv4 = 192.0.2.11`;
5. stale configuration assertion still says `192.0.2.10`;
6. optional retraction correcting one malformed observation.

### 5.2 Required State Result

The result must distinguish:

- intended value;
- observed value;
- source and observed time;
- whether the values conflict;
- freshness;
- supporting evidence IDs;
- first effective drift time.

The projector must not silently overwrite intended state with observed state or vice versa.

### 5.3 Required Timeline Result

Timeline includes all records in operational order and retains the later ingestion time for delayed evidence.

### 5.4 Required Change Result

At minimum:

- deployment event;
- observed IP change;
- conflict/drift opened;
- conflict resolved if a later intended-state update occurs.

### 5.5 Pass Criteria

- exact semantic state fixture matches;
- drift is explicit;
- every selected/conflicting value cites evidence;
- replay is duplicate-safe;
- retraction produces the expected corrected projection;
- rebuild produces equivalent results.

## 6. Scenario A — RTT Spike After Tunnel or Route Change

Implemented after topology and change projections.

### 6.1 Input

Resources:

- `node/frankfurt-01`;
- `node/london-01`;
- optional service/endpoints using the path.

Evidence sequence:

1. active WireGuard peer relationship;
2. baseline RTT observations around 30 ms;
3. endpoint or preferred-route change event;
4. relationship evidence reflecting the new endpoint/path;
5. RTT observations around 80 ms;
6. operator hypothesis that route change caused the increase;
7. optional confirming or contradicting probe evidence.

### 6.2 Required Topology Result

- relationship before the change;
- relationship after the change;
- validity intervals;
- evidence IDs;
- active state at each queried time.

### 6.3 Required Change Result

- route/endpoint relationship change;
- RTT anomaly/change;
- shared correlation window or ID where supplied;
- no duplicate changes after fixture replay.

### 6.4 Required Explanation Context

Must contain:

- current state and freshness;
- recent relationship change;
- RTT observations before and after;
- hypothesis status;
- supporting evidence;
- contradicting evidence;
- wording/structured status that distinguishes correlation from confirmed root cause.

### 6.5 Pass Criteria

- point-in-time topology is correct;
- timeline order is correct despite any delayed record;
- explanation links the preceding change without overclaiming causation;
- confirmed root cause appears only when confirmation policy is satisfied.

## 7. Scenario B — Large S3 Upload Failure

Implemented after incident, hypothesis, action, and outcome memory.

### 7.1 Input

Resources:

- device group or representative devices;
- S3 upload gateway;
- network path/edge;
- target object-storage endpoint.

Evidence sequence:

1. baseline upload success observations;
2. ASN, route, or endpoint change;
3. failure events grouped by object size;
4. small-object successes;
5. large-object failures;
6. MTU/MSS or PMTU probe observations;
7. multiple hypotheses;
8. decision to run a reversible MSS-clamp test;
9. action record with approval and idempotency references;
10. post-action observations;
11. outcome and reusable lesson.

### 7.2 Required Incident Context

- affected resources;
- symptom timeline;
- size-class pattern;
- path/topology context;
- open, supported, contradicted, and rejected hypotheses;
- decision rationale;
- action record;
- pre/post measurements;
- outcome;
- final conclusion with evidence links;
- applicability constraints for reuse.

### 7.3 Required Similar-Incident Result

A query describing “small uploads succeed, large uploads fail after network path change” must rank the fixture incident above unrelated incidents using deterministic filters plus FTS5.

### 7.4 Pass Criteria

- no hypothesis is confirmed merely because it appears in an operator note;
- action recording performs no external execution;
- outcome success cites post-action measurement evidence;
- pre/post values are correct;
- reusable lesson includes the topology/symptom constraints under which it applies;
- similar incident is retrieved within the configured top results.

## 8. Golden Comparison Rules

Golden files compare semantic content rather than unstable implementation details.

Ignored or normalized fields may include:

- generated IDs when fixtures do not pin them;
- ingestion timestamps generated by a fixed test clock;
- internal SQLite row order not present in the contract;
- FTS floating-point rank when only relative order is contractual.

Never ignore:

- evidence references;
- observed/validity time;
- conflict state;
- stale flags;
- source type/ref;
- status/confidence;
- selected values;
- topology direction/state;
- incident hypothesis/conclusion status.

Golden updates require review of the semantic diff. Tests must not automatically accept new output.

## 9. Core Query Evaluation

### 9.1 `GetResourceState`

Checks:

- resource identity;
- selected values;
- intended/observed scope;
- conflicts;
- stale status;
- supporting evidence;
- projector version;
- last change.

### 9.2 `GetResourceTimeline`

Checks:

- bounded window;
- stable cursor pagination;
- operational order;
- ingestion delay visibility;
- kind/status filters;
- evidence links.

### 9.3 `TraceEvidence`

Checks:

- inbound/outbound links;
- supports/contradicts/caused-by/retracts/supersedes semantics;
- cycle protection;
- bounded depth and result count.

### 9.4 `QueryTopology`

Checks:

- point-in-time validity;
- direction;
- relation type;
- active/inactive/uncertain state;
- evidence references;
- missing-resource semantics.

### 9.5 `GetRecentChanges`

Checks:

- before/after values;
- effective time;
- correlation;
- confidence;
- evidence references;
- deterministic replay.

### 9.6 `ExplainResource`

Checks structured sections:

- summary inputs;
- current state;
- conflicts;
- recent changes;
- timeline evidence;
- topology;
- incidents;
- similar incidents;
- hypotheses;
- stale/data-gap warnings;
- all evidence references resolvable.

## 10. Adapter Contract Evaluation

### 10.1 CLI

- JSON fields use stable snake_case names;
- JSON-valued fields are embedded values, not double-encoded strings;
- output contains query window and projector version metadata;
- text output does not change core semantics;
- typed errors map to documented non-zero exit codes;
- list queries with no matches return empty arrays.

### 10.2 MCP

- initialize negotiation is valid;
- notification requests produce no response;
- parse errors use JSON-RPC `id: null`;
- server version is build-derived;
- invalid arguments return correct tool/protocol error classification;
- structured content matches core contracts;
- result limits are bounded;
- all explanation tools include evidence references;
- write tools are absent or explicitly disabled by default.

## 11. Agent-Facing Evaluation

Agent evaluation uses fixed questions over scenario snapshots.

Example questions:

- “What changed before Frankfurt RTT increased?”
- “Is the route change the confirmed root cause?”
- “Why are large uploads failing while small uploads work?”
- “Which mitigation was tested and what evidence shows the result?”
- “Does the current observed IP match intended state?”
- “Which facts are stale or conflicting?”

Score dimensions:

| Dimension | Required behavior |
|---|---|
| Evidence accuracy | cited IDs exist and support the claim |
| Temporal accuracy | chronology and validity are correct |
| Uncertainty | correlation, hypothesis, and confirmation are distinguished |
| Conflict disclosure | incompatible valid evidence is exposed |
| Freshness | stale information is labeled |
| Completeness | important topology/change/incident context is included |
| Restraint | no unsupported root cause or action outcome is invented |
| Reproducibility | same snapshot yields equivalent structured context |

Optional natural-language summaries fail if they introduce a claim not represented by structured context.

## 12. Retrieval Evaluation

Initial similar-incident retrieval uses deterministic filtering and FTS5.

Metrics:

- Recall@5 for relevant fixture incidents;
- Precision@5 for unrelated incident rejection;
- rank of the expected closest incident;
- evidence/subject reference resolvability;
- temporal/applicability filter correctness.

Vector retrieval is considered only after a representative corpus shows a documented lexical failure.

## 13. Migration and Rebuild Evaluation

For every released V2 schema:

- migrate empty database to latest;
- migrate prior release snapshot to latest;
- verify foreign keys and integrity;
- rebuild each projection;
- compare semantic outputs;
- verify raw evidence hashes and counts unchanged;
- test interrupted rebuild recovery;
- test backup before destructive repair.

## 14. Collector Evaluation

Recorded source fixtures verify:

- source timestamp preservation;
- source-reference stability;
- cursor advancement only after commit;
- overlapping window dedupe;
- delayed data;
- retry behavior;
- field redaction;
- secret exclusion;
- stale collector health;
- bounded batches.

Collectors are rejected if they write storage directly or bypass typed evidence validation.

## 15. Concurrency and Failure Evaluation

Tests cover:

- concurrent duplicate appends;
- SQLite busy/timeout behavior;
- canceled context before and during transaction;
- partial batch validation failures;
- projection failure before commit;
- projection rebuild interruption;
- corrupted/invalid JSON input;
- disk or read-only database errors where practical;
- clean process shutdown.

## 16. Reference Performance Dataset

Baseline dataset:

- 10,000 resources;
- 1,000,000 evidence records;
- 100,000 active/historical relationships;
- representative changes, incidents, and FTS documents.

Provisional targets on a documented machine:

- batch append plus synchronous projection: at least 200 evidence records/second;
- current resource state p95: below 100 ms;
- 24-hour resource timeline p95: below 250 ms;
- one-hop topology p95: below 200 ms;
- deterministic explanation context p95: below 500 ms, excluding optional LLM summarization;
- bounded memory during projection rebuild and JSONL import.

Benchmarks report database size, WAL behavior, machine details, Go version, and SQLite driver configuration.

Targets change only with recorded benchmark evidence and an updated design decision.

## 17. CI Gates by Milestone

### Milestone 1

- unit tests;
- SQLite integration;
- migration tests;
- architecture fitness test;
- Scenario C;
- rebuild equivalence;
- race tests for append/dedupe where supported.

### Milestone 2

Adds:

- topology/change tests;
- MCP protocol contracts;
- Scenario A.

### Milestone 3

Adds:

- collector contract fixtures;
- retry/cursor tests;
- real-source smoke tests where secrets are unnecessary.

### Milestone 4

Adds:

- incident/hypothesis/action/outcome tests;
- retrieval evaluation;
- Scenario B.

### Milestone 5

Adds:

- full agent question set;
- backup/restore/export/import;
- benchmark report;
- packaging/install smoke tests;
- compatibility contract checks.

## 18. Release Acceptance

A V2 alpha release is blocked unless:

- all three canonical scenarios pass;
- every explanatory structured result cites evidence;
- stale and conflicting data are represented;
- projection rebuild is verified;
- backup/restore works;
- MCP protocol defects from the demo are resolved;
- performance baseline is recorded;
- compatibility and schema policies are documented;
- no action execution is enabled by default.
