# MemInfra Current Status

> Live execution checkpoint. Read this immediately after `PROJECT_MEMORY.md` when resuming work.

Last updated: 2026-08-02
Active integration branch: `main`
Last completed pull request: #12 — `refactor(core): establish V2 application boundary`
Execution tracker: #11 — `V2 implementation tracker: evidence-based operational memory`
Stage: Milestone 1 / Slice 1.1 merged; Slice 1.2 is the next implementation unit

## Durable Baselines

Memory-first design/reset:

```text
d03c6b5f22fba30fae1a3d6e20e5c3d53805d6fe
Merge pull request #10 from mss-boot-io/refactor/memory-core-v2
```

V2 application boundary:

```text
54d8c4e58d203a7a673b6932c688ddf99f36d7d7
Merge pull request #12 from mss-boot-io/refactor/memory-kernel-v2-s1-1
```

Original anti-loss design checkpoint:

```text
2d9c43792f69334041698f28794df4963e9994b2
docs: checkpoint MemInfra memory-first redesign
```

Read `docs/PROJECT_MEMORY.md`, issue #11, and ADRs 0001–0004 before changing V2 semantics.

## Slice 1.1 — Completed and Merged

PR #12 completed the first implementation slice:

- [x] isolated the V1 store-forwarding service under `internal/legacy/core`;
- [x] preserved existing CLI and MCP demo behavior through the legacy boundary;
- [x] replaced `internal/core` with a persistence- and protocol-independent V2 core;
- [x] introduced foundational `Clock`, `IDGenerator`, and `TransactionManager` ports;
- [x] added `core.Application` construction and validated port delegation;
- [x] added typed application errors independent of adapters;
- [x] added `internal/app/bootstrap` as the composition root;
- [x] made adapter shutdown idempotent and reverse-ordered;
- [x] established adapter, ingestion, and projection package namespaces;
- [x] added executable architecture dependency rules;
- [x] prohibited every repository-local outward import from `internal/core`;
- [x] explicitly prohibited CLI, HTTP, MCP, gRPC, database, GORM, and SQLite dependencies from the core;
- [x] added policy self-tests for the architecture guard;
- [x] added full-repository `go vet` to Makefile and pull-request CI.

Important implementation commits retained in merge history:

```text
ee47791f64289f9cc5c78e527909401343bf9cdd
refactor: isolate V1 demo service behind legacy package

9387d352fb9a2e6b5b2dc71987063ef52b240491
feat(core): add V2 application boundary skeleton

27c812ced750091f2e8e3f5bd96378d4a2a5ae66
test(architecture): reject all outward core dependencies

21e0a8f776296a6bc75a6a14659088d1a96ea94f
ci: add Go vet validation target

a4e1c31d35a879fa841c978784c2680b7f72c195
ci: enforce Go vet on pull requests
```

## Validation of Slice 1.1

Final PR head:

```text
692ba54106a194806aae5c708062a7f07e533d23
```

Remote validation on that exact head:

```text
CI #25 / Validate: success
CodeQL #36 / Analyze Go: success
```

The successful CI gates included:

- formatting;
- module metadata/tidy verification;
- full-repository `go vet` with SQLite FTS5 build tags;
- installer shell syntax and configuration safety;
- all Go tests;
- both CLI and MCP binary builds;
- installer smoke test;
- binary artifact upload.

Focused dependency-free V2 packages were also formatted and tested in an isolated local workspace:

```text
go test ./internal/core \
  ./internal/app/bootstrap \
  ./internal/architecture \
  ./internal/adapters \
  ./internal/ingest \
  ./internal/projection
```

Result: passed.

A complete local repository clone was not available because the execution container could not resolve `github.com`. The full dependency graph was therefore validated by GitHub Actions; no separate full local run is claimed.

## Enforced Architecture Rule

`internal/core` may import:

- ordinary domain-safe Go standard-library packages;
- its own `internal/core/...` subpackages;
- future third-party domain-only libraries after explicit review.

It may not import:

- any other package in this repository;
- `database/sql`;
- CLI frameworks or `flag`;
- HTTP, gRPC, or MCP protocol packages;
- GORM;
- concrete SQLite drivers;
- V1 model, store, index, legacy, CLI, or MCP code.

The rule is executable in `internal/architecture/dependencies_test.go`.

## Temporary V1 Boundary

The existing demonstration remains under:

```text
internal/legacy/core
```

It is temporary and is not a V2 contract. Do not add new product behavior to it. Remove legacy paths only when an equivalent V2 vertical slice replaces and validates the behavior.

## Slice 1.2 — Exact Scope

The next implementation branch must implement domain primitives only:

- opaque typed IDs for workspace, resource, evidence, incident, correlation, and change;
- stable textual prefixes and strict parsing/validation;
- resource-key parsing, normalization, and invariants;
- evidence kind enum;
- evidence status enum;
- confidence value validation in `[0, 1]`;
- observed, ingested, and validity time value types;
- source type and source reference value types;
- JSON and text round-trip behavior;
- table-driven invariant and malformed-input tests.

Slice 1.2 must not introduce:

- evidence payload schemas;
- SQLite V2 migrations or repositories;
- collectors;
- new product MCP tools;
- HTTP or Web UI;
- vector retrieval;
- incident feature expansion;
- infrastructure action execution.

## Exact Next Action

1. Verify the push CI and CodeQL status for this merged-main checkpoint commit.
2. Create `refactor/memory-kernel-v2-s1-2` from the resulting current `main`.
3. Implement only the Slice 1.2 domain primitives listed above.
4. Keep all new domain code inside `internal/core/...` so the architecture guard applies.
5. Run formatting, full-repository vet/test/build/installer checks, CodeQL, and focused invariant tests.
6. Update issue #11 and this file with actual commits, validation, risks, and the next slice before ending that work session.

## Resume Rules

After any restart:

1. read `docs/PROJECT_MEMORY.md`;
2. read this file;
3. inspect issue #11 and the latest five commits on `main`;
4. verify actual workflow status rather than relying only on recorded status;
5. continue the first incomplete item in `Exact Next Action`;
6. never infer planned implementation from documentation alone.
