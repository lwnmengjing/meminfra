# ADR 0005: Opaque Domain Value Objects at the Core Boundary

- Status: Accepted
- Date: 2026-08-02
- Depends on: ADR 0001

## Context

MemInfra V2 must preserve evidence identity, time, provenance, and confidence consistently across CLI, MCP, collectors, SQLite, exports, and future adapters.

Using unconstrained strings, floats, and `time.Time` values throughout the application would permit:

- wrong identifier prefixes;
- malformed or non-canonical ULIDs;
- ambiguous Go zero values;
- invalid resource keys;
- unknown evidence kinds or statuses;
- NaN, infinity, or out-of-range confidence;
- zero or unordered validity times;
- silent JSON `null` acceptance;
- adapter-specific normalization rules.

Such invalid values would reach persistence and projections before being rejected.

## Decision

MemInfra V2 represents foundational domain concepts as opaque value objects under `internal/core/domain`.

Initial value objects include:

- typed IDs for workspaces, resources, evidence, incidents, correlations, and changes;
- resource keys;
- evidence kinds and statuses;
- confidence;
- observed and ingested instants;
- validity intervals;
- source types and source references.

Rules:

1. Values are constructed through explicit constructors or parsers.
2. Internal representation is not directly mutable by adapters.
3. Invalid or uninitialized values fail serialization where absence would be ambiguous.
4. JSON and text decoding apply the same domain validation as programmatic construction.
5. Canonical normalization occurs only where the product contract explicitly permits it.
6. Strict values such as enums and IDs are not silently normalized.
7. Persistence and protocol adapters map to and from these value objects rather than redefining invariants.

## Consequences

### Positive

- Invalid states are rejected before application commands and persistence.
- Adapters share one normalization and validation contract.
- Identifier categories cannot be accidentally interchanged at compile time.
- Valid confidence `0` is distinguishable from an unset Go zero value.
- Time and interval semantics become explicit and testable.
- Serialization behavior is deterministic.

### Negative

- Mapping code is required at adapter boundaries.
- The core contains more small types and constructors.
- Zero-value behavior must be designed deliberately for each type.
- Strict parsing can reject previously tolerated input, requiring adapters to report useful errors.

## Alternatives Rejected

### Primitive Type Aliases Only

Rejected because exported string/float aliases remain freely constructible and can bypass invariants.

### Validate Only in CLI or MCP

Rejected because collectors, tests, imports, and future APIs would implement inconsistent policy.

### Validate Only in SQLite Constraints

Rejected because errors arrive too late, persistence-specific rules leak into the product, and many semantics cannot be expressed reliably as SQL constraints.

### Normalize Every String Input

Rejected because silently changing IDs, enum values, or source references can collapse distinct identities or hide producer defects. Normalization is limited to explicitly canonicalized values such as resource keys and source-type slugs.

## Verification

- malformed values are covered by table-driven tests;
- JSON and text round trips use constructors/parsers;
- JSON `null` and unknown fields are rejected where invalid;
- zero values cannot serialize as valid required values;
- `internal/core/domain` has no persistence, protocol, CLI, collector, or source-SDK imports;
- repository-wide architecture tests remain green.
