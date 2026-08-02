# MemInfra Current Status

> Live execution checkpoint. Read this immediately after `PROJECT_MEMORY.md` when resuming work.

Last updated: 2026-08-02
Branch: `refactor/memory-kernel-v2-s1-1`
Pull request: #12 — `refactor(core): establish V2 application boundary`
Execution tracker: #11 — `V2 implementation tracker: evidence-based operational memory`
Stage: Milestone 1 / Slice 1.1 implemented and validated; awaiting final PR review/merge

## Durable Design Baseline

The memory-first redesign was merged through PR #10:

```text
d03c6b5f22fba30fae1a3d6e20e5c3d53805d6fe
Merge pull request #10 from mss-boot-io/refactor/memory-core-v2
```

The original anti-loss design checkpoint remains:

```text
2d9c43792f69334041698f28794df4963e9994b2
docs: checkpoint MemInfra memory-first redesign
```

Read `docs/PROJECT_MEMORY.md` and ADRs 0001–0004 before changing V2 semantics.

## Completed in Slice 1.1

- [x] merge the V2 design/reset into `main`;
- [x] create `refactor/memory-kernel-v2-s1-1` from the merged design baseline;
- [x] move the V1 store-forwarding service from `internal/core` to `internal/legacy/core`;
- [x] preserve the existing CLI and MCP demo by routing them through the legacy package;
- [x] create a new persistence- and protocol-independent `internal/core`;
- [x] define foundational `Clock`, `IDGenerator`, and `TransactionManager` ports;
- [x] add `core.Application` construction and validated port delegation;
- [x] add stable typed application errors independent of adapters;
- [x] preserve typed adapter/application errors through core boundaries;
- [x] add `internal/app/bootstrap` as the composition root;
- [x] close adapter resources once and in reverse construction order;
- [x] establish adapter, ingestion, and projection package namespaces;
- [x] add an architecture fitness test for the V2 core boundary;
- [x] reject every repository-local import outside `internal/core` from the core;
- [x] explicitly reject CLI, HTTP, MCP, database, GORM, SQLite, and gRPC dependencies from the core;
- [x] test the dependency policy itself against allowed and forbidden examples;
- [x] open PR #12 and complete full CI and CodeQL validation.

## Commit Sequence

```text
ee47791f64289f9cc5c78e527909401343bf9cdd
refactor: isolate V1 demo service behind legacy package

9387d352fb9a2e6b5b2dc71987063ef52b240491
feat(core): add V2 application boundary skeleton

27c812ced750091f2e8e3f5bd96378d4a2a5ae66
test(architecture): reject all outward core dependencies
```

The status-only commit containing this file is newer than the validated code head and changes no Go behavior.

## Validation

### Focused Local Validation

The dependency-free V2 packages were formatted and tested in an isolated local workspace:

```text
go test ./internal/core \
  ./internal/app/bootstrap \
  ./internal/architecture \
  ./internal/adapters \
  ./internal/ingest \
  ./internal/projection
```

Result: passed.

The strengthened architecture policy was rerun after adding the repository-wide outward-import rule and its policy test. Result: passed.

### Full GitHub Actions Validation

Validated code head:

```text
27c812ced750091f2e8e3f5bd96378d4a2a5ae66
```

CI:

```text
CI #21 / Validate: success
```

Successful checks include:

- checkout and Go setup;
- module download;
- formatting check;
- module metadata/tidy check;
- installer shell syntax;
- installer configuration safety test;
- all Go tests, including the new architecture and core tests;
- both CLI and MCP binary builds;
- installer smoke test;
- binary artifact upload.

Security:

```text
CodeQL #32 / Analyze Go: success
```

The earlier implementation head `9387d352fb9a2e6b5b2dc71987063ef52b240491` also passed CI #20 and CodeQL #31 before the architecture policy was strengthened.

### Local Environment Limitation

A complete local clone remains unavailable in the execution container because DNS resolution for `github.com` fails. Therefore the full dependency graph was not independently retested locally. GitHub Actions is the authoritative full-repository validation; the focused dependency-free packages were independently tested locally as recorded above.

## Architecture Boundary Now Enforced

`internal/core` may import:

- the Go standard library, except concrete adapter/protocol packages explicitly forbidden by policy;
- its own `internal/core/...` subpackages;
- future third-party domain-only libraries after review.

It may not import:

- any other package in this repository;
- `database/sql`;
- CLI frameworks or `flag`;
- HTTP, gRPC, or MCP protocol packages;
- GORM;
- concrete SQLite drivers;
- V1 model, store, index, legacy, CLI, or MCP code.

This is executable policy in `internal/architecture/dependencies_test.go`, not documentation only.

## Preserved V1 Behavior

The current demo remains buildable and tested through:

```text
internal/legacy/core
```

This is a temporary migration boundary, not a V2 contract. V1 code must be removed incrementally when equivalent V2 adapters and use cases replace it. New product behavior must never be added to the legacy service.

## Scope Still Not Implemented

- evidence/domain identifiers;
- resource-key rules;
- observed, ingested, and validity time types;
- evidence status and confidence semantics;
- typed evidence payloads;
- SQLite V2 migrations;
- append-evidence commands and repositories;
- projection implementations;
- collectors;
- V2 CLI/MCP product tools;
- incident learning;
- action execution.

## Exact Next Action

1. Verify CI and CodeQL on the final status-only PR head.
2. Mark PR #12 ready for review.
3. Review and merge PR #12 with its full commit history.
4. Update issue #11 to mark Slice 1.1 complete.
5. Create a new branch from updated `main` for Milestone 1 / Slice 1.2.
6. Implement domain primitives only:
   - opaque typed IDs for workspace, resource, evidence, incident, correlation, and change;
   - resource-key parsing and normalization;
   - evidence kind and status enums;
   - confidence validation;
   - observed/ingested/validity time types;
   - source/provenance value types;
   - JSON/text round-trip and invariant tests.
7. Do not add SQLite schema, collectors, MCP tools, HTTP/UI, vector retrieval, or action execution in Slice 1.2.
8. Update this file again before ending the next substantial work session.

## Resume Rules

After any restart:

1. read `docs/PROJECT_MEMORY.md`;
2. read this file;
3. inspect PR #12, issue #11, and the latest five commits;
4. verify the actual branch head and workflow status;
5. continue the first incomplete item in `Exact Next Action`;
6. never assume planned code exists until confirmed in the repository.
