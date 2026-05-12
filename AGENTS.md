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

- **Binary project.** Has `main.go` at the root. Produces the `mlb-mcp`
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

## Tool design

Tools are composable answers, not raw endpoint wrappers. Each tool
should answer a user question in one call — "who won today", "show me
Ohtani's stats", "what are the standings" — not expose a bare SDK
method.

1. **One tool = one user intent.** If the user has to call two tools
   to get an answer, merge them into one tool that does both calls.
2. **10–20 tools max.** LLMs pick the right tool faster with fewer
   choices. Prefer fewer, richer tools over many thin ones.
3. **Tools call 1–3 SDK methods internally.** The composition logic
   (filter today's games, enrich with scores, etc.) lives in the
   handler, not in the LLM's reasoning.
4. **Return complete answers.** Format results so the LLM can relay
   them directly — don't return raw IDs that need a follow-up lookup.
5. **Name tools as user intents.** `scores` not `get_schedule`.
   `player_bio` not `get_person`. The name tells the LLM when to
   pick it.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/). Scopes:
`cmd`, `mcp`, `docs`, `chore`, `test`. Subject ≤ 50 chars. Body wraps
at 72.

When the agent is Claude, end every commit with:

```
🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```
