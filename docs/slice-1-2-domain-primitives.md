# Slice 1.2 — V2 Domain Primitives

Last updated: 2026-08-02
Status: implementation checkpoint on `refactor/memory-kernel-v2-s1-2`

## Scope

This slice establishes persistence- and protocol-independent value types required by the immutable evidence kernel.

Implemented under `internal/core/domain`:

- typed opaque identifiers for workspaces, resources, evidence, incidents, correlations, and derived changes;
- canonical prefixed ULID parsing and serialization;
- normalized resource keys;
- evidence kind and evidence status enums;
- initialized confidence values in `[0, 1]`;
- observed time, ingestion time, and half-open validity intervals;
- extensible source-type slugs and source-local references;
- typed validation errors;
- strict text and JSON decoding;
- table-driven invariant and round-trip tests.

## Deliberate Exclusions

This slice does not add:

- evidence envelopes or typed evidence payload schemas;
- SQLite schema, repositories, migrations, or projections;
- collectors/importers;
- CLI or MCP product capabilities;
- HTTP, UI, vector retrieval, or infrastructure execution.

Those remain ordered by `docs/roadmap-v2.md`.

## Identifier Contract

Canonical text forms:

```text
ws_<26-character uppercase canonical ULID>
res_<26-character uppercase canonical ULID>
evd_<26-character uppercase canonical ULID>
inc_<26-character uppercase canonical ULID>
cor_<26-character uppercase canonical ULID>
chg_<26-character uppercase canonical ULID>
```

Parsing is strict:

- exact prefix and length;
- uppercase canonical Crockford Base32;
- forbidden ambiguous characters rejected;
- ULID timestamp overflow rejected;
- outer whitespace and lowercase variants rejected;
- zero values cannot be serialized as valid identifiers.

## Resource-Key Contract

Canonical structure:

```text
<kind>/<identity-segment>[/<identity-segment>...]
```

Examples:

```text
node/frankfurt-01
service/s3-put_gateway/us-mia-1
bucket/ubiasnap-us
```

Rules:

- input is trimmed and normalized to lowercase;
- at least one identity segment is required;
- forward slashes delimit segments;
- kind starts with an ASCII letter and supports lowercase letters, digits, hyphens, and underscores;
- identity segments start and end with a lowercase letter or digit and may additionally contain `.`, `:`, and `@`;
- empty, dot-path, Unicode, control/space, and backslash forms are rejected;
- total and per-segment sizes are bounded.

## Evidence Classification

Kinds:

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

Statuses:

```text
observed
asserted
inferred
confirmed
contradicted
retracted
```

Enums use strict canonical values; adapters cannot silently normalize unknown values.

## Confidence Contract

`Confidence`:

- must be explicitly constructed or decoded;
- accepts only finite values in the inclusive range `[0, 1]`;
- serializes as a JSON number;
- rejects `null`, unset zero values, NaN, infinity, and out-of-range values.

The additional initialized flag distinguishes valid confidence `0` from an uninitialized Go zero value.

## Time Contract

`ObservedAt` and `IngestedAt`:

- reject zero time;
- normalize to UTC;
- strip monotonic-clock data;
- use RFC3339Nano-compatible text and JSON strings.

`ValidityInterval`:

- is half-open: `[valid_from, valid_to)`;
- supports an open-ended interval;
- requires `valid_to > valid_from` when closed;
- supports containment and overlap queries;
- rejects zero, malformed, null, unknown-field, and ambiguous trailing-separator inputs.

## Source Contract

`SourceType`:

- is normalized to lowercase;
- starts with an ASCII letter and ends with an alphanumeric character;
- supports lowercase letters, digits, hyphens, underscores, and dots;
- is bounded to 64 bytes.

`SourceReference`:

- trims only outer whitespace;
- preserves case and internal source-specific structure;
- requires valid UTF-8;
- rejects control characters;
- is bounded to 1024 bytes.

## Validation

Local isolated package validation performed before the branch commit:

```text
go test ./internal/core/domain
go vet ./internal/core/domain
go test -race ./internal/core/domain
```

All passed in the isolated local module used to develop this package.

Repository-wide validation is delegated to the pull-request CI and CodeQL workflows because the local execution environment could not resolve `github.com` to clone and download the repository dependencies.

## Next Slice

After this slice is reviewed and merged, Milestone 1 Slice 1.3 implements versioned typed evidence payload schemas, canonical payload encoding, validation, and content hashing. It must build on these domain primitives without introducing persistence or protocol coupling.
