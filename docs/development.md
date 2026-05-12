# Development Guide

## Prerequisites

Install tools using [mise](https://mise.jdx.dev/):

```bash
mise install
```

- **[Go](https://go.dev)** >= 1.25
- **[just](https://just.systems)** — task runner

## Quick Start

```bash
git clone https://github.com/retr0h/mlb-mcp.git
cd mlb-mcp
go build ./...
go test ./...
```

## Layout

```
main.go                    Entry point; delegates to cmd.Execute()
cmd/root.go                Cobra root command
cmd/mcp.go                 mcp parent command
cmd/mcp_start.go           mcp start — runs the MCP server over stdio
internal/mcp/server.go     Server struct, Config, New(), Run()
internal/mcp/session.go    Driver interface (consumer-side abstraction)
internal/mcp/tools.go      Tool registration + handlers
internal/mcp/json.go       JSON marshaling helper
```

This is a binary project — `main.go` at the root produces the `mlb-mcp`
executable. All MLB data access goes through the mlb-sdk library; this project
never calls `statsapi.mlb.com` directly.

## Common tasks

```bash
go build ./...             # build
go test ./...              # run tests
go vet ./...               # vet
gofmt -l .                 # find unformatted files
```

Or via just:

```bash
just deps
just test
just ready          # fmt + vet + lint
```

## Adding a new tool

There are two kinds of tools. Pick the right workflow:

### Adding a composed tool (hand-written, intent-focused)

A composed tool answers a **user intent** in one call — not a bare SDK method.
See `AGENTS.md § Tool design` for the design rules. Each tool calls 1–3 SDK
methods internally and returns a complete answer.

1. **`internal/mcp/session.go`** — if the tool needs an SDK method not yet on
   the `Driver` interface, add it. The concrete `*mlb.Client` must already
   satisfy the new method structurally.
2. **`internal/mcp/tools.go`** — register the tool via `mcpsdk.AddTool` in
   `registerTools()`. Write the args struct (with `json` + `jsonschema` tags)
   and handler func. Wrap errors as `fmt.Errorf("mlb-mcp: <tool>: %w", err)`.
3. **`internal/mcp/server_test.go`** — add rows to `TestServer_CallTool` for the
   new tool: happy path, missing required args, SDK error propagation. Use the
   `fakeDriver`. Coverage must stay at 100.0%.
4. **`internal/mcp/server.go`** — update the `instructions` const to mention the
   new tool under Tier 1.
5. **`internal/mcp/mcpgen/main.go`** — add the tool's `operationId` to
   `composedOps` so the generator skips it.
6. **`README.md`** — add a row to the **Composed tools** table.
7. **`go generate ./internal/mcp/`** — regenerate `tools_gen.go` so the raw tool
   is removed (now covered by the composed version).
8. **`just ready`** — final gate. fmt + vet + lint + 100% coverage all green
   before committing.

### Adding a raw tool (auto-generated from OpenAPI spec)

When a new endpoint is added to [mlb-sdk][]'s OpenAPI spec, run:

```bash
go generate ./internal/mcp/
```

The `mcpgen` tool reads the embedded spec from `mlb-sdk/pkg/api`, generates
typed args structs and HTTP handlers for every non-composed operation, and
writes `internal/mcp/tools_gen.go`. Update the raw tools table in `README.md` to
match.

[mlb-sdk]: https://github.com/retr0h/mlb-sdk

## Public surface authoring

This project does not expose a Go library API. Its public surface is the set of
MCP tools it registers and the JSON schemas those tools accept and return.

### Error handling

- Handler functions return `(content, error)`.
- Wrap errors with the tool name: `fmt.Errorf("mlb-mcp: <tool>: %w", err)`.
- Propagate mlb-sdk errors as-is; callers can `errors.Is` against
  `mlb.ErrNotFound`.
- Handler functions never panic; they always return errors.

### Functional options

The server constructor and any configurable component use functional options:

```go
type Option func(*config)
type config struct { /* internal */ }

func WithThing(v T) Option { return func(c *config) { c.thing = v } }
```

## Testing conventions

**Every public function and method MUST have a table-driven test.** One table
per function, with rows covering both the happy path and every failure mode the
function can produce. Failure rows belong in the same table as the happy row —
not in a separate test.

> **Anti-pattern (do not do this):** writing a separate one-off test function
> for a failure scenario. Each public function gets exactly **one** `Test*`
> function in the codebase. Reviewers should reject PRs that introduce
> additional one-off tests for the same function.

### File naming (non-negotiable)

**One test file per production file.** `tools.go` is tested by `tools_test.go`.
If tests outgrow one file, split the production file first.

Two exceptions: shared fixtures in `helpers_test.go` / `fakes_test.go`, and
`main_test.go` for `TestMain`.

### Table shape

For tool handlers, each row injects a fake `Driver` and asserts on either the
JSON result or the error:

```go
func TestTodayScores(t *testing.T) {
    cases := []struct {
        name    string
        driver  Driver       // fake implementation
        wantErr string       // error substring; "" means expect nil
        want    string       // JSON substring in result
    }{
        {name: "happy path", driver: fakeWithGames(2), want: `"gamePk"`},
        {name: "sdk error", driver: fakeWithError(errBoom), wantErr: "today_scores"},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) { ... })
    }
}
```

### Required failure rows for tool handlers

Every tool handler's table MUST include rows covering:

| Row                  | Setup                                                          |
| -------------------- | -------------------------------------------------------------- |
| Happy path           | Fake Driver returns valid data                                 |
| Missing required arg | Args struct missing a required field → expect validation error |
| SDK error            | Fake Driver returns an error → expect wrapped error            |
| Empty result         | Fake Driver returns empty data → expect graceful zero values   |

### Branchy args need branch-per-row coverage

When a handler conditionally sets query fields (`if args.Season != 0`), the
happy-path row only exercises one branch. Add rows that exercise each branch so
coverage stays at 100.0%.

### Test naming

- Tool handlers: `TestToolTodayScores`, `TestToolStandings`.
- Pure functions: `TestFunctionName`.
- Methods on a type: `TestType_Method`.

### Coverage gate

**Coverage is 100.0% of statements.** `main.go` is excluded via `.coverignore`.
Run `just go::test` to confirm.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- **Subject line**: max 50 characters, imperative mood, capitalized, no period
- **Body**: wrap at 72 characters, separated from subject by a blank line
- **Format**: `type(scope): description`
- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`
- **Scopes**: `cmd`, `mcp`, `docs`
