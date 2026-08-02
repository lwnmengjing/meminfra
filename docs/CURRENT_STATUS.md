# MemInfra Current Status

> Live execution checkpoint. Read this immediately after `PROJECT_MEMORY.md` when resuming work.

Last updated: 2026-08-02
Branch: `refactor/memory-core-v2`
Pull request: #10 — `docs: redesign MemInfra around evidence-based operational memory`
Execution tracker: #11 — `V2 implementation tracker: evidence-based operational memory`
Stage: Milestone 0 design/reset complete; awaiting review/merge

## Durable Checkpoint

The initial anti-loss checkpoint was pushed before all other redesign work:

```text
2d9c43792f69334041698f28794df4963e9994b2
docs: checkpoint MemInfra memory-first redesign
```

That commit contains the complete product definition, principles, architecture, data model, roadmap, safety model, acceptance scenarios, and restart procedure needed to recover after a process restart.

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
- [x] open PR #10 against `main`;
- [x] create umbrella implementation tracker #11;
- [x] run pull-request CI and CodeQL successfully on the complete design content.

## Validation

### GitHub Actions Baseline

Validated content head:

```text
4d4c78c1b0a79fdb02f854582bb638b0bf33a6db
```

CI run:

```text
CI #16 / Validate: success
```

Successful steps:

- checkout;
- Go setup;
- module download;
- formatting check;
- module metadata/tidy check;
- installer shell syntax;
- installer configuration safety test;
- Go tests;
- Go builds;
- installer smoke test;
- binary artifact upload.

Security run:

```text
CodeQL #27 / Analyze Go: success
```

The current status-only commit is newer than the validated content head. It changes no product code or design semantics and triggers its own PR checks; GitHub remains the authoritative source for the final head status.

### Repository Checks

Confirmed through GitHub repository operations:

- all design and status files exist on the redesign branch;
- PR #10 is open and mergeable;
- branch commits were successfully pushed and fetched;
- existing source and prior PR #1 review threads were inspected;
- the V2 tracker exists as issue #11.

### Local Environment Limitation

A direct local clone was attempted but not available because the execution environment could not resolve `github.com`.

Therefore these checks were not independently rerun in the local container:

```text
go test
go build
installer smoke test
SQLite runtime tests
```

They did run successfully in GitHub Actions as recorded above. Do not claim a separate local validation.

## Known V1 Review Concerns Preserved as V2 Requirements

- stable snake_case CLI/MCP JSON DTOs;
- JSON-RPC parse errors include `id: null`;
- MCP server version is build-derived;
- install documentation states actual toolchain prerequisites;
- numeric/ID inputs are bounded before narrowing conversions.

These are correctness requirements, not V1 compatibility promises.

## Current Design Decision

No product code has been changed in PR #10. The redesign deliberately stops after establishing the product and engineering contract.

The next implementation must start at the memory kernel. It must not begin with collectors, MCP expansion, HTTP, UI, vector retrieval, incident feature growth, or action execution.

## Exact Next Action

1. Verify the final PR-head CI and CodeQL checks after this status-only commit.
2. Mark PR #10 ready for review.
3. Review and merge the design/reset PR.
4. Create the first implementation branch from the updated `main`.
5. Implement Milestone 1 Slice 1.1 only:
   - package/dependency skeleton;
   - core ports;
   - typed application errors;
   - clock and ID generator interfaces;
   - bootstrap composition root;
   - architecture dependency-boundary test.
6. Update this file with the new branch, commit, checks, risks, and exact next action before ending that work session.

## Resume Rules

After any restart:

1. read `docs/PROJECT_MEMORY.md`;
2. read this file;
3. inspect PR #10, issue #11, and the latest five commits;
4. verify the actual branch head and check status;
5. continue the first incomplete exact-next-action item;
6. never assume planned code exists until confirmed in the repository.
