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
go install github.com/retr0h/mlb-mcp@latest
```

## Usage

Run the MCP server over stdio:

```bash
mlb-mcp mcp start
```

### Claude Code

This repo ships a `.claude/mcp.json` that Claude Code auto-discovers when
your working directory is the repo root. Clone and open:

```bash
git clone https://github.com/retr0h/mlb-mcp.git
cd mlb-mcp
claude   # tools are available immediately
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mlb-mcp": {
      "command": "mlb-mcp",
      "args": ["mcp", "start"]
    }
  }
}
```

### Other MCP clients

Any client that supports stdio transport can launch `mlb-mcp mcp start` as
a subprocess. See [`.claude/mcp.json`](.claude/mcp.json) for the canonical
config.

## Tools

| Tool                  | Description                                                    |
| --------------------- | -------------------------------------------------------------- |
| `today_scores`        | Today's scores and game status across MLB                      |
| `standings`           | Division standings for AL, NL, or both leagues                 |
| `player_bio`          | Player biographical info by MLB person ID                      |
| `team_info`           | Team metadata (venue, league, division) by team ID             |
| `team_roster`         | Active roster for a team                                       |
| `league_leaders`      | Stat leaders (HR, AVG, ERA, ...) with sensible defaults        |
| `game_detail`         | Boxscore (team batting/pitching totals) for a game             |
| `game_linescore`      | Inning-by-inning breakdown for a game                          |
| `recent_transactions` | Recent trades, signings, DFAs (defaults to today)              |
| `free_agents`         | Free-agent declarations and signings for a season              |
| `postseason_schedule` | Postseason game schedule for a season                          |

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
