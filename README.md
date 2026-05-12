[![go report card](https://goreportcard.com/badge/github.com/retr0h/mlb-mcp?style=for-the-badge)](https://goreportcard.com/report/github.com/retr0h/mlb-mcp)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![build](https://img.shields.io/github/actions/workflow/status/retr0h/mlb-mcp/go.yml?style=for-the-badge)](https://github.com/retr0h/mlb-mcp/actions/workflows/go.yml)
[![conventional commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![built with just](https://img.shields.io/badge/Built_with-Just-black?style=for-the-badge&logo=just&logoColor=white)](https://just.systems)
![github commit activity](https://img.shields.io/github/commit-activity/m/retr0h/mlb-mcp?style=for-the-badge)
[![MLB](https://img.shields.io/badge/MLB-002D72?style=for-the-badge&logo=mlb&logoColor=white)](https://mlb.com)

# mlb-mcp

MCP server for the MLB Stats API.

[Model Context Protocol (MCP)][mcp] is an open standard that lets LLMs
call external tools and data sources in a structured, type-safe way. This
server wraps the [mlb-sdk][] Go library and exposes MLB Stats API endpoints
as MCP tools that any MCP-compatible LLM (Claude, GPT-4, etc.) can call
directly.

## Install

```bash
go install github.com/retr0h/mlb-mcp/cmd/mlb-mcp@latest
```

## Usage

Run the server (stdio transport, compatible with any MCP client):

```bash
mlb-mcp
```

Configure your MCP client to launch `mlb-mcp` as a subprocess. For Claude
Desktop, add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mlb": {
      "command": "mlb-mcp"
    }
  }
}
```

## Tools

| Tool          | MLB Stats API endpoint                   | Description                        |
| ------------- | ---------------------------------------- | ---------------------------------- |
| *(none yet — add rows here as tools are implemented)* | | |

## Features

| Feature           | Description                                               |
| ----------------- | --------------------------------------------------------- |
| MCP stdio server  | Works with any MCP-compatible client out of the box       |
| Typed responses   | All MLB data surfaced as structured JSON via mlb-sdk      |
| Idiomatic Go      | Functional options, context propagation, wrapped errors   |

## ⚖️ Copyright notice

This package and its author are not affiliated with MLB or any MLB team. This
module is a typed Go client for MLB's public Stats API. Use of MLB data is
subject to the notice posted at
<http://gdx.mlb.com/components/copyright.txt>.

## License

The [MIT][] License.

[mcp]: https://modelcontextprotocol.io
[mlb-sdk]: https://github.com/retr0h/mlb-sdk
[MIT]: LICENSE
[Development]: docs/development.md
[Contributing]: docs/contributing.md
