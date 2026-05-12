# Contributing to mlb-mcp

Thanks for your interest in contributing.

## Before you start

- **Check existing work** — search open issues and PRs to avoid duplicating
  effort.
- **Start small** — focused PRs are easier to review than sweeping changes.

## Making changes

### Code style

- Run `gofmt -l .` and `go vet ./...` before pushing.
- Tests: `go test ./...`. Add tests for any non-trivial behavior.
- Idiomatic Go — see [development.md](development.md).

### Adding a new MCP tool

1. Implement the handler in `internal/tools/<resource>.go`.
2. Register the tool in `cmd/mlb-mcp/main.go`.
3. Add tests in `internal/tools/<resource>_test.go` — 100% coverage required.
4. Update the `## Tools` table in `README.md`.

### Documentation

- Update docs alongside code changes when tool behavior changes.

## Submitting a PR

1. Create a feature branch from `main` (`type/short-description` — `feat/`,
   `fix/`, `docs/`, `refactor/`, `chore/`).
2. Commit messages:
   [Conventional Commits](https://www.conventionalcommits.org/), see
   [development.md](development.md#commit-messages).
3. PR description: what changed, why, and any follow-ups.
4. Open as draft if you want early feedback before final review.
5. One logical change per PR — split unrelated changes.
