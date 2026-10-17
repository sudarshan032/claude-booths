# Contributing

Thanks for helping. Bugs and feature ideas go in [issues](https://github.com/sudarshan032/claude-booths/issues/new/choose),
questions in [discussions](https://github.com/sudarshan032/claude-booths/discussions), and security reports
privately (see [SECURITY.md](SECURITY.md)). Everyone taking part follows the
[Code of Conduct](CODE_OF_CONDUCT.md).

## Setup

```bash
git clone https://github.com/sudarshan032/claude-booths.git && cd claude-booths
go mod download
go build -o ~/.local/bin/claude-booths .
```

Requires Go (version pinned by the `toolchain` line in `go.mod`), `tmux` and `git`.

## Sources of truth

| What | Where |
|---|---|
| Go version | `toolchain` line in `go.mod` (CI and release read it from there) |
| App version | `version` in `main.go`; release tags must match (`release.yml` checks) |
| User-facing changes | `CHANGELOG.md` (`## Unreleased` until a release is cut) |
| Config/state location | `config.GetConfigDir()` → `~/.claude-booths` |
| tmux session naming | `tmux.TmuxPrefix` (`claudebooths_`) |
| Colors | `ui/theme` palettes only; styles are built in `theme.OnChange` hooks (no hard-coded colors elsewhere) |
| CI | `.github/workflows/ci.yml` (test, lint, build); releases: `release.yml` + `.goreleaser.yaml` |

## Branches

`main` is always releasable. Work on short-lived branches named `<type>/<short-kebab-description>`:

`feat/…` · `fix/…` · `perf/…` · `refactor/…` · `test/…` · `docs/…` · `ci/…` · `chore/…`

e.g. `fix/branch-picker-in-use`, `perf/stream-pane-output`. Merge via PR (squash), then delete the branch.

## Commits

[Conventional Commits](https://www.conventionalcommits.org/): `type(scope): summary`, imperative, ≤72 chars.
Scopes are package names: `app`, `ui`, `tmux`, `git`, `config`, `session`, `daemon`, `ci`.

```
fix(git): don't offer branches already checked out in another worktree

Body explains why (what broke, for whom), not what the diff says.
```

`feat` and `fix` commits also get a line in `CHANGELOG.md` under `## Unreleased`.

## Tests

- Every fix gets a test that fails without it; every feature gets tests for its logic.
- Tests must never touch your real state: each package's `TestMain` calls
  `internal/testutil.Isolate()`, which points `HOME` and `TMUX_TMPDIR` at a temp dir.
  New test packages must do the same.
- Before pushing:

```bash
gofmt -l .        # must print nothing
go vet ./...
go test -race -count=1 ./...
golangci-lint run # v2.5.0, same as CI (config: .golangci.yml)
```

## Releases

1. Move `## [Unreleased]` entries in `CHANGELOG.md` under `## [X.Y.Z] – YYYY-MM-DD` and add its
   compare link at the bottom.
2. Bump `version` in `main.go`, commit `chore(release): vX.Y.Z`.
3. `git tag vX.Y.Z && git push origin main vX.Y.Z` — `release.yml` builds a draft release.
