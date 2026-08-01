# MemInfra V1 Demo Baseline

Last updated: 2026-08-02
Status: historical implementation baseline for V2 planning

## 1. Baseline

V2 redesign branch:

```text
refactor/memory-core-v2
```

Created from `main` at:

```text
0426749be268a40754ff062341b2087a9cb04f9a
```

The baseline contains no production deployment or data that must be preserved.

## 2. Existing Demo Capabilities

The demo includes:

- Go module targeting Go 1.26;
- SQLite through GORM and the current SQLite driver;
- FTS5 enabled by the `sqlite_fts5` build tag;
- CLI command families for resources, observations, events, incidents, relationships, topology, and search;
- text and JSON rendering;
- a minimal read-oriented MCP stdio server;
- installer scripts;
- GitHub Actions for CI, CodeQL, release packaging, and Dependabot.

## 3. Existing Demo Architecture

```text
CLI / MCP
    |
internal/core
    |
internal/store
    |
GORM + SQLite + FTS5
```

Observed structural issues:

- `internal/core` aliases input and option types from `internal/store`;
- core methods mostly forward directly to store methods;
- `internal/store` owns validation, defaults, source policy, transactions, persistence, limits, topology direction, search, and projection writes;
- GORM persistence models are also domain/output models;
- resource current properties are mutable columns rather than temporal evidence;
- incidents store root cause, solution, and result as unlinked text;
- memory documents concatenate entity fields into FTS strings;
- search documents do not provide a full evidence lineage model;
- no collector/discovery pipeline establishes observed truth;
- no projection rebuild contract exists.

These findings motivate the breaking V2 design. They are not a request to preserve the V1 abstractions.

## 4. Existing MCP Surface

The demo exposes:

```text
search_memory
list_resources
get_resource
query_topology
list_observations
list_events
list_incidents
```

The surface is primarily entity/list/query oriented. V2 replaces this with memory capabilities after the evidence kernel is implemented.

## 5. Known Review Defects From PR #1

Four unresolved review concerns remain relevant as correctness requirements if corresponding code is reused:

1. `search --output json` directly encoded Go model fields, producing inconsistent field names instead of stable snake_case DTOs.
2. JSON-RPC parse-error responses could omit `id`; compliant pre-ID errors must contain `id: null`.
3. MCP `initialize` reported a hard-coded `0.1.0` server version instead of build-derived version information.
4. Source-build installation documentation did not clearly state the C toolchain/CGO prerequisites of the selected demo driver.

A previous CodeQL integer-conversion finding in CLI ID parsing was marked resolved in the old PR, but V2 ID parsing and protocol inputs still require bounded validation.

## 6. Existing Workflow Baseline

The repository defines:

- CI validation on pushes to `main` and pull requests targeting `main`;
- format check;
- module tidy check;
- installer shell checks and smoke test;
- Go tests and builds;
- artifact upload;
- CodeQL analysis;
- tag-based release packaging;
- Dependabot updates.

V2 must extend this baseline with:

- architecture dependency tests;
- SQLite migration tests;
- projection rebuild tests;
- canonical scenario tests;
- MCP protocol contract tests;
- collector contract tests;
- backup/restore tests;
- benchmark reporting at release gates.

## 7. Validation Performed During Redesign Session

Confirmed through the GitHub repository interface:

- the base repository and permissions were accessible;
- branch `refactor/memory-core-v2` was created from `main`;
- the authoritative project-memory commit was pushed and fetched successfully;
- subsequent design commits were pushed to the same branch;
- old PR #1 review threads were inspected;
- current source files and documentation were read from the branch.

Not performed in the current execution environment:

- local clone;
- `go test`;
- `go build`;
- installer smoke test;
- SQLite runtime test.

Reason: the available local execution environment could not resolve `github.com`, so a clone and module retrieval could not be performed. This limitation must not be represented as a successful code validation.

The design changes are documentation-only. Opening a pull request will trigger the repository’s available pull-request workflows and provide the next remote validation signal.

## 8. V1 Disposition

V1 code can be:

- reused when it cleanly fits a V2 adapter;
- used as a behavior/example fixture;
- removed when the corresponding V2 vertical slice replaces it.

V1 code must not force:

- schema compatibility;
- CLI compatibility;
- MCP compatibility;
- GORM in the core;
- mutable current-state semantics;
- plain-text incident conclusions;
- FTS documents as durable truth.

## 9. First Implementation Baseline

After the design PR, the first code slice is limited to:

- V2 package/dependency skeleton;
- architecture fitness test;
- core IDs/time/provenance/status primitives;
- no collector expansion;
- no new MCP tool;
- no incident feature expansion;
- no action execution.

This ensures future progress begins at the memory core rather than adding another outer surface.
