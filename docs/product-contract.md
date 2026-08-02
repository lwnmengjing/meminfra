# MemInfra V2 Product Contract

Last updated: 2026-08-02
Status: accepted design contract for `refactor/memory-core-v2`

## 1. Product Statement

MemInfra is an AI-native operational memory for infrastructure.

It ingests evidence from infrastructure systems and operators, preserves time and provenance, derives current state and changes, and provides explainable context to AI agents and human operators.

MemInfra is not the source system for metrics, logs, cloud inventory, tickets, or automation. It is the memory layer that relates those sources into a traceable operational history.

## 2. Primary User

The primary user is an AI agent operating with an infrastructure engineer.

Examples include Codex, Claude Code, OpenCode, Cursor, and MCP clients. The agent must be able to retrieve structured context without scraping dashboards or trusting unqualified prose.

Human operators use the same core through CLI commands, exports, and inspectable SQLite data.

## 3. Jobs to Be Done

### 3.1 Understand Current State

Given a resource, return:

- current known values;
- freshness and validity;
- source and confidence;
- conflicts;
- supporting evidence IDs;
- active topology relationships;
- last material change.

### 3.2 Explain Recent Change

Given a resource and time window, return:

- ordered timeline;
- before and after state;
- events and topology changes;
- correlated incidents and actions;
- supporting and contradicting evidence;
- gaps that prevent a stronger conclusion.

### 3.3 Reuse Operational Experience

Given a symptom or incident, return:

- similar historical incidents;
- affected resource and topology similarities;
- hypotheses considered;
- decisions and actions;
- measured outcomes;
- constraints that determine whether the old lesson applies.

### 3.4 Preserve Decisions and Outcomes

Record:

- what question was decided;
- alternatives considered;
- selected option and rationale;
- approval and action references;
- pre-action evidence;
- post-action evidence;
- measured result.

The initial product records this loop but does not execute infrastructure changes.

## 4. Required Product Properties

### 4.1 Traceable

Every derived state, change, explanation, incident conclusion, and lesson must reference stable evidence IDs.

### 4.2 Temporal

The system distinguishes observed time, ingestion time, and validity time. Query defaults are bounded and explicit.

### 4.3 Honest About Uncertainty

The system exposes stale, conflicting, missing, inferred, contradicted, or retracted evidence. It does not silently collapse uncertainty into one value.

### 4.4 Idempotent

Repeated ingestion of the same source record does not create duplicate semantic evidence.

### 4.5 Rebuildable

Current state, topology, changes, search documents, and summaries can be reconstructed from durable evidence.

### 4.6 Inspectable

A local operator can inspect, back up, export, and verify the SQLite database without a remote service.

### 4.7 Safe

Collectors are read-only by default. Secrets are not stored. Infrastructure action execution is deferred until explicit approval, policy, idempotency, audit, and outcome requirements exist.

## 5. Canonical Questions

A complete read/explain milestone must answer these questions with structured output:

1. What do we currently know about `resource_key`?
2. Which evidence establishes each selected value?
3. How old is the selected evidence?
4. Are there valid conflicting values?
5. What changed between time A and time B?
6. What event or relationship change occurred immediately before the symptom?
7. Which evidence supports and contradicts a hypothesis?
8. Which similar incident had a confirmed outcome?
9. What information is missing or stale?
10. Can the result be reproduced from the same database snapshot?

## 6. Canonical Scenarios

### 6.1 RTT Spike

A tunnel or route changes, followed by an RTT increase. MemInfra must correlate the timeline without claiming causation until evidence confirms it.

### 6.2 Large S3 Upload Failure

Small uploads succeed while large objects fail after a network-path change. MemInfra must assemble size-class failures, MTU/MSS evidence, topology, hypotheses, test actions, and outcomes.

### 6.3 Resource Drift

Intended state conflicts with observed state. MemInfra must show both values, their sources and freshness, and mark the projection as drifted or uncertain.

These scenarios are acceptance tests, not documentation examples only.

## 7. Product Boundaries

MemInfra does not aim to own:

- metric visualization and alert-rule management;
- raw log retention and querying at log-platform scale;
- cloud resource lifecycle management;
- ticket queues and general workflow;
- configuration deployment;
- secret management;
- arbitrary graph analytics;
- autonomous remediation;
- a traditional Web UI.

A proposed feature is in scope only when it improves evidence preservation, state/change projection, contextual retrieval, incident learning, or safe action outcome memory.

## 8. Initial Deployment Contract

The first complete deployment is:

- one local MemInfra process;
- one SQLite database;
- FTS5 lexical retrieval;
- CLI and MCP read access;
- read-only collectors;
- explicit backup/export;
- no mandatory network service;
- no multi-tenant control plane.

## 9. Compatibility Contract

The existing V1 demo has no production compatibility guarantee.

For V2 development:

- the schema may be replaced;
- command names may change;
- MCP tool names and outputs may change;
- existing demo databases need not migrate;
- repository history is the archive of V1 behavior.

Compatibility begins only after a V2 alpha contract is explicitly tagged and documented.

## 10. Success Criteria

V2 is useful when an AI agent can answer the canonical questions using real collected evidence and every answer includes:

- time context;
- source/provenance;
- supporting evidence IDs;
- conflicts and uncertainty;
- topology and recent changes where relevant;
- related incident/action/outcome history;
- stale-data and data-gap warnings.

The number of tables, commands, adapters, or MCP tools is not a success metric.

## 11. Feature Decision Test

Before adding a feature, answer:

1. Which canonical question or scenario does it improve?
2. What durable evidence or projection does it introduce?
3. How is the result traced to evidence?
4. How does it handle time, freshness, conflict, and retraction?
5. Can the projection be rebuilt?
6. Does it duplicate a source system?
7. What is the end-to-end acceptance test?

Features without satisfactory answers do not enter the roadmap.
