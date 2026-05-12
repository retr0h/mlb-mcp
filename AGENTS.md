# Agent Instructions

Canonical guidance for any AI coding agent (Claude Code, Cursor, Aider, …)
working in this repository. Where Claude-Code-specific rules diverge they live
in [`CLAUDE.md`](CLAUDE.md); everything in this file applies regardless of
which agent is driving.

> **Read [`docs/development.md`](docs/development.md) before writing code.**
> It is the source of truth for code organization, naming, error handling,
> the functional-options pattern, the table-driven testing rules every public
> function must follow, and the recipe for adding a new MCP tool.

## Project shape

- **Binary project.** Has `cmd/mlb-mcp/main.go`. Produces the `mlb-mcp`
  executable.
- **MCP server.** Implements the [Model Context Protocol][mcp] so LLMs can
  call MLB Stats API endpoints as tools.
- **Wraps mlb-sdk.** All MLB data access goes through
  `github.com/retr0h/mlb-sdk/pkg/mlb`. This project never calls the MLB
  Stats API directly.
- **PR workflow.** Feature branches + pull requests into `main`.
  Branch protection is enabled; direct pushes to `main` are blocked.

[mcp]: https://modelcontextprotocol.io

## Tasks you'll be asked to do

| Task                   | Where to look                                                          |
| ---------------------- | ---------------------------------------------------------------------- |
| Add a new MCP tool     | [`docs/development.md` → Adding a new tool](docs/development.md#adding-a-new-tool) |
| Write a test           | [`docs/development.md` → Testing conventions](docs/development.md#testing-conventions) |

## Hard rules

1. **Never import `internal/` packages from mlb-sdk.** Only ever import
   `github.com/retr0h/mlb-sdk/pkg/mlb`.
2. **Every public function gets exactly one table-driven test.** Failure rows
   live in the same table as the happy row — no separate one-off failure
   tests.
3. **Coverage is 100.0% of statements.** A change that drops coverage is a
   regression and must be brought back to 100% before commit.
4. **`just ready` is the gate before committing.** Runs fmt, vet, and lint.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/). Scopes:
`cmd`, `mcp`, `docs`, `chore`, `test`. Subject ≤ 50 chars. Body wraps
at 72.

When the agent is Claude, end every commit with:

```
🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```
