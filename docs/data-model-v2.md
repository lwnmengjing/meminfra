# MemInfra V2 Data Model

Last updated: 2026-08-02
Status: accepted logical model; physical schema may be refined by migration tests and benchmarks

## 1. Model Goals

The V2 model must preserve infrastructure memory rather than mirror source-system tables.

It must support:

- stable resource identity;
- immutable evidence;
- source provenance;
- observed, ingestion, and validity time;
- idempotent ingestion;
- retraction and supersession;
- causal and evidentiary links;
- rebuildable resource state, topology, changes, and search;
- incidents, hypotheses, decisions, actions, and outcomes;
- traceable explanations.

The V1 demo schema is not a compatibility constraint.

## 2. Identifier Strategy

Externally visible domain identifiers are opaque text IDs generated through a core `IDGenerator` port.

The initial implementation should use monotonic ULIDs:

- sortable by creation time;
- safe for local/offline generation;
- stable across export/import;
- easy to include in logs and MCP output.

Prefixes improve diagnosis without changing identity semantics:

```text
ws_<ulid>
res_<ulid>
evd_<ulid>
inc_<ulid>
chg_<ulid>
```

Database code must not derive behavior from prefixes.

Human-facing identities use stable keys such as:

```text
node/frankfurt-01
service/s3-put-gateway/us
bucket/ubiasnap-us
tunnel/frankfurt-01/london-01
```

## 3. Time Semantics

All persisted timestamps use UTC with sufficient precision and are serialized as RFC 3339.

Evidence distinguishes:

- `observed_at`: time at which the source observed or asserted the record;
- `ingested_at`: time at which MemInfra committed it;
- `valid_from`: beginning of semantic validity;
- `valid_to`: optional exclusive end of validity.

Rules:

- `observed_at` may be older than `ingested_at`;
- `valid_from` defaults to `observed_at` when omitted;
- `valid_to`, when present, must be greater than `valid_from`;
- ingestion order never replaces operational chronology;
- tie-breaking uses stable evidence ID after semantic ordering fields.

## 4. Durable Core Tables

### 4.1 `workspaces`

The first release creates one local workspace but preserves the boundary for export and future separation.

```text
id                TEXT PRIMARY KEY
workspace_key     TEXT NOT NULL UNIQUE
display_name      TEXT NOT NULL
created_at        TIMESTAMP NOT NULL
metadata_json     TEXT NOT NULL DEFAULT '{}'
```

Default key: `local`.

### 4.2 `resources`

A resource row is identity, not current state.

```text
id                TEXT PRIMARY KEY
workspace_id      TEXT NOT NULL
resource_key      TEXT NOT NULL
display_name      TEXT NOT NULL DEFAULT ''
kind              TEXT NOT NULL
identity_json     TEXT NOT NULL DEFAULT '{}'
created_at        TIMESTAMP NOT NULL
retired_at        TIMESTAMP NULL

UNIQUE(workspace_id, resource_key)
FOREIGN KEY(workspace_id) REFERENCES workspaces(id)
```

`identity_json` contains only source-correlation identifiers that must remain stable. Mutable properties such as IP address, provider, role, health, and region are evidence and projected state.

### 4.3 `resource_aliases`

Correlates source-specific identifiers to one resource.

```text
id                TEXT PRIMARY KEY
workspace_id      TEXT NOT NULL
resource_id       TEXT NOT NULL
source_type       TEXT NOT NULL
alias_type        TEXT NOT NULL
alias_value       TEXT NOT NULL
created_at        TIMESTAMP NOT NULL
retired_at        TIMESTAMP NULL

UNIQUE(workspace_id, source_type, alias_type, alias_value)
FOREIGN KEY(resource_id) REFERENCES resources(id)
```

Aliases are retired rather than silently reassigned. Reassignment requires an explicit audited command.

### 4.4 `evidence`

Immutable evidence envelope.

```text
id                TEXT PRIMARY KEY
workspace_id      TEXT NOT NULL
subject_key       TEXT NOT NULL
subject_resource_id TEXT NULL
kind              TEXT NOT NULL
schema_name       TEXT NOT NULL
schema_version    INTEGER NOT NULL
payload_json      TEXT NOT NULL
source_type       TEXT NOT NULL
source_ref        TEXT NOT NULL DEFAULT ''
collector_id      TEXT NOT NULL DEFAULT ''
observed_at       TIMESTAMP NOT NULL
ingested_at       TIMESTAMP NOT NULL
valid_from        TIMESTAMP NOT NULL
valid_to          TIMESTAMP NULL
status            TEXT NOT NULL
confidence        REAL NOT NULL
correlation_id    TEXT NOT NULL DEFAULT ''
causation_id      TEXT NOT NULL DEFAULT ''
supersedes_id     TEXT NULL
dedupe_key        TEXT NOT NULL
content_hash      TEXT NOT NULL
tags_json         TEXT NOT NULL DEFAULT '[]'
metadata_json     TEXT NOT NULL DEFAULT '{}'

UNIQUE(workspace_id, dedupe_key)
FOREIGN KEY(workspace_id) REFERENCES workspaces(id)
FOREIGN KEY(subject_resource_id) REFERENCES resources(id)
FOREIGN KEY(supersedes_id) REFERENCES evidence(id)
```

Allowed initial `kind` values:

```text
fact
observation
event
relationship
operator_note
hypothesis
decision
action
outcome
retraction
```

Allowed initial `status` values:

```text
observed
asserted
inferred
confirmed
contradicted
retracted
```

`confidence` is in `[0, 1]`. Status and confidence are separate: a human assertion can have high confidence without becoming a directly observed fact.

Evidence rows are not updated through normal application code.

### 4.5 `evidence_links`

Represents many-to-many semantic relationships between evidence records.

```text
id                TEXT PRIMARY KEY
workspace_id      TEXT NOT NULL
from_evidence_id  TEXT NOT NULL
to_evidence_id    TEXT NOT NULL
link_type         TEXT NOT NULL
created_at        TIMESTAMP NOT NULL
metadata_json     TEXT NOT NULL DEFAULT '{}'

UNIQUE(from_evidence_id, to_evidence_id, link_type)
FOREIGN KEY(from_evidence_id) REFERENCES evidence(id)
FOREIGN KEY(to_evidence_id) REFERENCES evidence(id)
```

Initial link types:

```text
supports
contradicts
caused_by
correlated_with
supersedes
retracts
decision_for
action_for
outcome_of
derived_from
```

A link records the relationship; it does not mutate either evidence row.

### 4.6 `incidents`

Incident identity and immutable opening metadata.

```text
id                TEXT PRIMARY KEY
workspace_id      TEXT NOT NULL
incident_key      TEXT NOT NULL
title             TEXT NOT NULL
created_at        TIMESTAMP NOT NULL
created_by_ref    TEXT NOT NULL DEFAULT ''
metadata_json     TEXT NOT NULL DEFAULT '{}'

UNIQUE(workspace_id, incident_key)
```

Status, severity, affected resources, hypotheses, conclusions, decisions, actions, and outcomes are represented by evidence addressed to `incident/<incident_key>` and linked through the tables below.

### 4.7 `incident_resources`

```text
incident_id       TEXT NOT NULL
resource_id       TEXT NOT NULL
role              TEXT NOT NULL DEFAULT 'affected'
linked_at         TIMESTAMP NOT NULL
source_evidence_id TEXT NULL

PRIMARY KEY(incident_id, resource_id, role)
```

### 4.8 `incident_evidence`

```text
incident_id       TEXT NOT NULL
evidence_id       TEXT NOT NULL
role              TEXT NOT NULL
linked_at         TIMESTAMP NOT NULL

PRIMARY KEY(incident_id, evidence_id, role)
```

Initial roles include `symptom`, `context`, `supports`, `contradicts`, `decision`, `action`, `outcome`, and `conclusion`.

## 5. Typed Evidence Payload Schemas

All known schemas are represented by Go structs, validators, examples, and versioned JSON Schema documents.

Unknown schemas may be stored only through an explicit opaque-import mode and are not projected until support exists.

### 5.1 `fact.v1`

```json
{
  "predicate": "network.ipv4",
  "value": "192.0.2.10",
  "value_type": "string",
  "scope": "observed"
}
```

`scope` initially supports `observed`, `intended`, and `computed`.

### 5.2 `observation.numeric.v1`

```json
{
  "metric": "network.rtt_ms",
  "value": 82.4,
  "unit": "ms",
  "dimensions": {
    "peer": "node/london-01"
  }
}
```

MemInfra stores memory-relevant observations, not every high-frequency sample. Collectors apply configured aggregation, anomaly, or incident-window policies.

### 5.3 `event.v1`

```json
{
  "event_type": "route.changed",
  "severity": "warning",
  "message": "preferred route changed from fra to lon",
  "data": {
    "before": "fra",
    "after": "lon"
  }
}
```

### 5.4 `relationship.v1`

```json
{
  "src_resource_key": "node/frankfurt-01",
  "dst_resource_key": "node/london-01",
  "relation_type": "wireguard_peer",
  "state": "active",
  "attributes": {
    "interface": "wg0"
  }
}
```

`state` supports `active`, `inactive`, and `uncertain`.

### 5.5 `operator_note.v1`

```json
{
  "title": "Packet-size pattern",
  "body": "Small uploads succeed while large uploads fail.",
  "author_ref": "operator/lwx"
}
```

### 5.6 `hypothesis.v1`

```json
{
  "statement": "A PMTU black hole causes large upload failures.",
  "hypothesis_status": "open",
  "rationale": "Failures correlate with payload size after ASN change."
}
```

Status supports `open`, `supported`, `contradicted`, `confirmed`, and `rejected`. Confirmation requires supporting evidence links and no unresolved blocking contradiction according to policy.

### 5.7 `decision.v1`

```json
{
  "question": "Which mitigation should be tested first?",
  "selected_option": "Clamp TCP MSS at the gateway",
  "alternatives": [
    "Reduce device MTU",
    "Move traffic to another edge"
  ],
  "rationale": "Gateway change is reversible and affects only the test path."
}
```

### 5.8 `action.v1`

```json
{
  "action_type": "network.set_mss_clamp",
  "target": "gateway/s3-us",
  "parameters": {
    "mss": 1360
  },
  "approval_ref": "approval/change-1234",
  "idempotency_key": "change-1234:set-mss-1360",
  "execution_status": "recorded"
}
```

The early product records actions; it does not execute them.

### 5.9 `outcome.v1`

```json
{
  "action_evidence_id": "evd_...",
  "success": true,
  "summary": "Large upload success rate recovered.",
  "measurements": [
    {
      "metric": "upload.success_rate.large",
      "before": 0.17,
      "after": 0.94,
      "unit": "ratio"
    }
  ]
}
```

### 5.10 `retraction.v1`

```json
{
  "target_evidence_id": "evd_...",
  "reason": "Collector timestamp was parsed in the wrong timezone."
}
```

The target row remains stored. Projectors exclude it after applying the retraction.

## 6. Idempotency and Canonicalization

Known payloads are decoded into typed structs and encoded in a canonical field order before hashing.

`content_hash` uses SHA-256 over canonical payload bytes.

The importer supplies a source-stable record identity when available. The core constructs `dedupe_key` from:

```text
workspace
source_type
source_ref or source_record_identity
schema_name + schema_version
observed_at
content_hash
```

A duplicate append returns the existing evidence ID and a duplicate disposition; it is not a failure.

Two source records with different source identities but identical content remain separate evidence because provenance differs.

## 7. Projection Tables

### 7.1 `projection_checkpoints`

```text
projection_name   TEXT NOT NULL
projection_version INTEGER NOT NULL
workspace_id      TEXT NOT NULL
last_evidence_id  TEXT NOT NULL DEFAULT ''
last_ingested_at  TIMESTAMP NULL
state             TEXT NOT NULL
updated_at        TIMESTAMP NOT NULL
metadata_json     TEXT NOT NULL DEFAULT '{}'

PRIMARY KEY(projection_name, projection_version, workspace_id)
```

### 7.2 `resource_state`

One row per resource and projector version.

```text
workspace_id      TEXT NOT NULL
resource_id       TEXT NOT NULL
projection_version INTEGER NOT NULL
state_json        TEXT NOT NULL
conflicts_json    TEXT NOT NULL DEFAULT '[]'
supporting_evidence_json TEXT NOT NULL
freshness_json    TEXT NOT NULL
last_change_id    TEXT NULL
updated_at        TIMESTAMP NOT NULL

PRIMARY KEY(workspace_id, resource_id, projection_version)
```

`state_json` is a map from predicate to selected value envelope, not an unqualified property bag.

A selected value envelope contains:

```json
{
  "value": "192.0.2.10",
  "value_type": "string",
  "scope": "observed",
  "status": "observed",
  "confidence": 1.0,
  "observed_at": "2026-08-02T00:00:00Z",
  "valid_from": "2026-08-02T00:00:00Z",
  "valid_to": null,
  "stale": false,
  "evidence_ids": ["evd_..."]
}
```

### 7.3 `relationship_state`

```text
workspace_id      TEXT NOT NULL
relationship_key  TEXT NOT NULL
projection_version INTEGER NOT NULL
src_resource_id   TEXT NOT NULL
dst_resource_id   TEXT NOT NULL
relation_type     TEXT NOT NULL
state             TEXT NOT NULL
attributes_json   TEXT NOT NULL
valid_from        TIMESTAMP NOT NULL
valid_to          TIMESTAMP NULL
supporting_evidence_json TEXT NOT NULL
confidence        REAL NOT NULL
updated_at        TIMESTAMP NOT NULL

PRIMARY KEY(workspace_id, relationship_key, projection_version)
```

`relationship_key` is deterministic from source, destination, relation type, and semantic identity attributes.

### 7.4 `derived_changes`

```text
id                TEXT PRIMARY KEY
workspace_id      TEXT NOT NULL
resource_id       TEXT NULL
change_type       TEXT NOT NULL
predicate         TEXT NOT NULL DEFAULT ''
before_json       TEXT NOT NULL
after_json        TEXT NOT NULL
effective_at      TIMESTAMP NOT NULL
detected_at       TIMESTAMP NOT NULL
confidence        REAL NOT NULL
correlation_id    TEXT NOT NULL DEFAULT ''
supporting_evidence_json TEXT NOT NULL
projection_version INTEGER NOT NULL
```

Changes are derived and can be rebuilt.

### 7.5 `incident_state`

```text
workspace_id      TEXT NOT NULL
incident_id       TEXT NOT NULL
projection_version INTEGER NOT NULL
status            TEXT NOT NULL
severity          TEXT NOT NULL
opened_at         TIMESTAMP NOT NULL
closed_at         TIMESTAMP NULL
affected_resources_json TEXT NOT NULL
open_hypotheses_json TEXT NOT NULL
confirmed_conclusion_evidence_id TEXT NULL
last_activity_at  TIMESTAMP NOT NULL
updated_at        TIMESTAMP NOT NULL

PRIMARY KEY(workspace_id, incident_id, projection_version)
```

### 7.6 `search_documents`

```text
id                TEXT PRIMARY KEY
workspace_id      TEXT NOT NULL
doc_type          TEXT NOT NULL
subject_ref       TEXT NOT NULL
title             TEXT NOT NULL
body              TEXT NOT NULL
tags              TEXT NOT NULL
valid_from        TIMESTAMP NULL
valid_to          TIMESTAMP NULL
source_evidence_json TEXT NOT NULL
projection_version INTEGER NOT NULL
updated_at        TIMESTAMP NOT NULL

UNIQUE(workspace_id, doc_type, subject_ref, projection_version)
```

An FTS5 virtual table indexes selected columns using `search_documents.id` as the stable content reference.

## 8. State Projection Algorithm

For each resource predicate:

1. load non-retracted fact evidence relevant to the projection time;
2. apply supersession links;
3. discard invalid validity intervals;
4. group semantically equal values;
5. score candidate groups using explicit policy:
   - status;
   - scope;
   - source priority;
   - confidence;
   - observed time;
   - freshness;
6. select a value only when policy establishes a clear winner;
7. otherwise write a conflict containing all candidate values and evidence IDs;
8. derive a change when the selected value or conflict state differs from the prior projection;
9. retain evidence references for every selected or conflicting value.

“Latest row wins” is prohibited as the only selection rule.

## 9. Retraction and Supersession

- supersession means newer evidence replaces the target for a specific semantic claim;
- retraction means the target should no longer participate in projections;
- contradiction means both records remain active evidence but disagree;
- none of these operations deletes the target row;
- projectors resolve link graphs deterministically and reject cycles.

## 10. Incident Semantics

An incident conclusion is evidence, not a mutable `root_cause` text column.

A confirmed conclusion requires:

- one or more `supports` links;
- affected resource linkage;
- explicit confirmation status;
- no policy-blocking unresolved contradiction;
- actor/source reference;
- observed or decision time.

Actions and outcomes are linked so the system can compare pre-action and post-action observations.

## 11. Indexes

Initial physical indexes should cover:

- resource key and aliases;
- evidence by workspace/subject/observed time;
- evidence by source and source reference;
- evidence by correlation and causation ID;
- evidence kind/schema/status;
- unique dedupe key;
- evidence-link inbound/outbound traversal;
- incident evidence/resource links;
- relationship source/destination/type/time;
- derived changes by resource/effective time;
- projection checkpoint state.

Indexes are validated against query plans and benchmarks rather than added speculatively.

## 12. Data Integrity Rules

- JSON columns contain valid JSON with canonical empty values (`{}` or `[]`), never ambiguous empty strings.
- confidence is within `[0, 1]`.
- valid intervals are well ordered.
- referenced workspace/resource/evidence/incident rows exist.
- evidence is immutable through repository APIs.
- dedupe uniqueness is enforced by SQLite, not only application checks.
- projection version is mandatory.
- source and subject fields are bounded and non-empty according to schema.
- payload and metadata size limits are enforced before persistence.

## 13. Export and Backup

A portable export includes:

- workspace and resource identities;
- aliases;
- evidence;
- evidence links;
- incident identities and links;
- schema versions;
- optional projections.

Projections may be omitted because they are rebuildable.

Export supports redaction rules for payload and metadata paths. Import preserves IDs and detects collisions.

## 14. V1 Disposition

V1 tables (`observations`, `events`, plain-text incidents, relationships, memory documents) may be removed when the first V2 vertical slice is implemented.

No production migration is required. During development, V2 should use an explicit schema version and may use a separate database path to prevent accidental interpretation of a V1 demo database.
