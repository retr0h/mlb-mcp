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
cmd/mlb-mcp/main.go   Entry point; starts the MCP server
```

This is a binary project — `cmd/mlb-mcp/main.go` produces the `mlb-mcp`
executable. All MLB data access goes through the mlb-sdk library; this
project never calls `statsapi.mlb.com` directly.

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

An MCP tool exposes one MLB Stats API call as a function an LLM can invoke.
Each tool maps one-to-one onto a method on `mlb.Client`.

When adding a new tool, touch every one of these in order:

1. **`cmd/mlb-mcp/main.go`** — register the new tool with the MCP server,
   declaring its name, description, and input schema.
2. **A handler file** (e.g. `internal/tools/<resource>.go`) — implement
   the handler function: parse arguments, call the mlb-sdk method, marshal
   the result to JSON, and return it as MCP tool output. Wrap errors as
   `fmt.Errorf("mlb-mcp: <tool>: %w", err)`.
3. **A test file** (`internal/tools/<resource>_test.go`) — one table-driven
   test per handler, covering: happy path, empty/missing fields, mlb-sdk
   error propagation, argument parse failure. Coverage must stay at 100.0%.
4. **`README.md`** — add a row to the `## Tools` table.
5. **`just ready`** — final gate. fmt + vet + lint + 100% coverage all
   green before committing.

## Public surface authoring

This project does not expose a Go library API. Its public surface is the set
of MCP tools it registers and the JSON schemas those tools accept and return.

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

**Every public function and method MUST have a table-driven test.** One
table per function, with rows covering both the happy path and every failure
mode the function can produce. Failure rows belong in the same table as the
happy row — not in a separate test.

### File naming (non-negotiable)

**One test file per production file.** `tools/schedule.go` is tested by
`tools/schedule_test.go`. If tests outgrow one file, split the production
file first.

### Table shape

```go
func TestScheduleTool(t *testing.T) {
    cases := []struct {
        name    string
        args    map[string]any
        want    string // JSON substring expected in output
        wantErr string // error substring; "" means expect nil
    }{
        {name: "happy path", args: map[string]any{"date": "2024-04-01"}, want: `"gamePk"`},
        {name: "missing date", args: map[string]any{}, wantErr: "date"},
        {name: "sdk error", args: map[string]any{"date": "bad"}, wantErr: "schedule"},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) { ... })
    }
}
```

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- **Subject line**: max 50 characters, imperative mood, capitalized, no period
- **Body**: wrap at 72 characters, separated from subject by a blank line
- **Format**: `type(scope): description`
- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`
- **Scopes**: `cmd`, `mcp`, `docs`
