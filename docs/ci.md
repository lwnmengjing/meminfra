# MemInfra CI/CD

The public repository is:

```text
https://github.com/mss-boot-io/meminfra
```

## Workflows

### CI

File:

```text
.github/workflows/ci.yml
```

Runs on:

- pull requests to `main`
- pushes to `main`
- manual dispatch

Checks:

- `gofmt`
- `go mod tidy` cleanliness
- installer shell syntax
- `go test -tags sqlite_fts5 ./...`
- build `bin/meminfra`
- build `bin/meminfra-mcp`
- installer smoke test
- upload CI binaries as workflow artifacts

### Release

File:

```text
.github/workflows/release.yml
```

Runs on:

- tags matching `v*`
- manual dispatch

Release tag example:

```zsh
git tag v0.1.0
git push origin v0.1.0
```

Artifacts:

- `meminfra_<version>_linux_amd64.tar.gz`
- `checksums.txt`

The release workflow uploads artifacts to the workflow run and, for tag builds, publishes them to the GitHub Release.

### CodeQL

File:

```text
.github/workflows/codeql.yml
```

Runs on:

- pull requests to `main`
- pushes to `main`
- weekly schedule
- manual dispatch

### Dependabot

File:

```text
.github/dependabot.yml
```

Checks weekly for:

- Go module updates
- GitHub Actions updates

## Recommended Branch Protection

In GitHub repository settings, protect `main` and require:

- `CI / Validate`
- `CodeQL / Analyze Go`

Recommended settings:

- Require pull request before merging.
- Require status checks to pass before merging.
- Require branches to be up to date before merging.
- Require conversation resolution before merging.
- Do not allow force pushes.
- Do not allow deletions.

## Current Artifact Scope

The first release pipeline builds Linux amd64 artifacts. This keeps CGO + SQLite predictable for the initial public release. Multi-platform release artifacts can be added later with platform-specific CGO toolchains.
