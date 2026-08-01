# MemInfra Current Status

> Live execution checkpoint. Read this immediately after `PROJECT_MEMORY.md` when resuming work.

Last updated: 2026-08-02
Branch: `refactor/memory-core-v2`
Pull request: #10 — `docs: redesign MemInfra around evidence-based operational memory`
Stage: Milestone 0 design/reset

## Durable Checkpoint

The initial anti-loss checkpoint was pushed first:

```text
2d9c43792f69334041698f28794df4963e9994b2
docs: checkpoint MemInfra memory-first redesign
```

The branch now contains the full V2 design package.

## Completed

- [x] create `refactor/memory-core-v2` from `main` at `0426749be268a40754ff062341b2087a9cb04f9a`;
- [x] commit and push authoritative project memory before other work;
- [x] redefine README and product boundary;
- [x] add focused product contract;
- [x] replace architecture with core/adapter/projection design;
- [x] define immutable evidence and V2 projection data model;
- [x] add milestone roadmap with slice-level acceptance criteria;
- [x] define deterministic canonical scenarios and evaluation gates;
- [x] accept ADR 0001: memory-first breaking redesign;
- [x] accept ADR 0002: append-only evidence;
- [x] accept ADR 0003: SQLite local-first storage;
- [x] accept ADR 0004: rebuildable versioned projections;
- [x] retire the contradictory V1 session-progress document;
- [x] catalogue V1 structure, workflow baseline, and unresolved review defects;
- [x] open Draft PR #10 against `main`;
- [x] trigger pull-request CI and CodeQL.

## Validation

Confirmed through GitHub repository operations:

- every listed file exists on the redesign branch;
- PR #10 contains 13 changed files at its first opened head;
- branch commits were successfully pushed and fetched;
- existing source and prior PR #1 review threads were inspected.

Local checks not run:

```text
go test
go build
installer smoke test
SQLite runtime tests
```

Reason: the available local execution environment could not resolve `github.com`, so repository clone and module retrieval failed. Do not reinterpret this as a successful local validation.

Remote PR checks at the time this status file was created:

```text
CI: queued
CodeQL: queued
```

Fetch the current PR head and workflow runs before relying on these statuses.

## Known V1 Review Concerns to Preserve as V2 Requirements

- stable snake_case CLI/MCP JSON DTOs;
- JSON-RPC parse errors include `id: null`;
- MCP server version is build-derived;
- install documentation states actual toolchain prerequisites;
- numeric/ID inputs are bounded before narrowing conversions.

These are correctness requirements, not V1 compatibility promises.

## Current Design Decision

No product code has been changed in the redesign PR. The next implementation must start at the memory kernel, not at collectors, MCP expansion, HTTP, UI, or actions.

## Exact Next Action

1. Inspect CI and CodeQL for the latest PR head.
2. Fix any documentation/CI issue introduced by the redesign.
3. Mark PR #10 ready for review when checks are green or accurately documented.
4. Merge the design/reset PR after review.
5. Start Milestone 1 Slice 1.1: package-boundary skeleton and architecture fitness test.
6. Update this file with the new branch, commit, checks, and next action before ending the next work session.

## Resume Rules

After any restart:

1. read `docs/PROJECT_MEMORY.md`;
2. read this file;
3. inspect PR #10 and the latest five commits;
4. verify the actual branch head and check status;
5. continue the first incomplete exact-next-action item;
6. never assume planned code exists until confirmed in the repository.
