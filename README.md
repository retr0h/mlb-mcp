[![go report card](https://goreportcard.com/badge/github.com/retr0h/mlb-mcp?style=for-the-badge)](https://goreportcard.com/report/github.com/retr0h/mlb-mcp)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![build](https://img.shields.io/github/actions/workflow/status/retr0h/mlb-mcp/go.yml?style=for-the-badge)](https://github.com/retr0h/mlb-mcp/actions/workflows/go.yml)
[![conventional commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![built with just](https://img.shields.io/badge/Built_with-Just-black?style=for-the-badge&logo=just&logoColor=white)](https://just.systems)
![github commit activity](https://img.shields.io/github/commit-activity/m/retr0h/mlb-mcp?style=for-the-badge)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=for-the-badge)](https://pkg.go.dev/github.com/retr0h/mlb-mcp)
[![MLB](https://img.shields.io/badge/MLB-002D72?style=for-the-badge&logo=mlb&logoColor=white)](https://mlb.com)

# mlb-mcp

⚾ MCP server for the MLB Stats API.

[Model Context Protocol (MCP)][mcp] is an open standard that lets LLMs
call external tools and data sources in a structured, type-safe way. This
server wraps the [mlb-sdk][] Go library and exposes MLB Stats API endpoints
as MCP tools that any MCP-compatible LLM (Claude, GPT-4, etc.) can call
directly. No API key required — MLB's Stats API is public.

## 📦 Install

```bash
go install github.com/retr0h/mlb-mcp@latest
```

## 🚀 Usage

Run the MCP server over stdio:

```bash
mlb-mcp mcp start
```

### Claude Code

This repo ships a [`.mcp.json`](.mcp.json) that Claude Code auto-discovers
when your working directory is the repo root. Clone and open:

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
a subprocess. See [`.mcp.json`](.mcp.json) for the canonical config.

## ⚙️ Tools

The server exposes **65 tools** in two tiers. Prefer the composed tools for
common questions; the raw tools cover everything else.

### Composed tools

Hand-written, intent-focused wrappers that answer common questions directly.
They accept friendly arguments and return clean, structured JSON from the
typed [mlb-sdk][] Go library.

| Tool                  | Description                                                    |
| --------------------- | -------------------------------------------------------------- |
| `scores`              | Scores and results for any date (defaults to today)            |
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

### Auto-generated raw tools (`mlb_*`)

Every other endpoint from the [mlb-sdk][] OpenAPI spec is auto-generated at
build time by `mcpgen` (run via `go generate ./internal/mcp/`). These tools
pass parameters directly to `statsapi.mlb.com` and return raw JSON. When a
new endpoint is added to mlb-sdk, regenerating picks it up automatically.

<details>
<summary>54 raw tools (click to expand)</summary>

| Tool | Description |
| ---- | ----------- |
| `mlb_get_all_seasons` | All historical seasons |
| `mlb_get_all_star_ballot` | All-Star ballot |
| `mlb_get_all_star_final_vote` | All-Star final vote |
| `mlb_get_all_star_write_ins` | All-Star write-ins |
| `mlb_get_attendance` | Attendance records |
| `mlb_get_award_recipients` | Award recipients (MVP, HOF, ...) |
| `mlb_get_conferences` | Conferences |
| `mlb_get_context_metrics` | Win probability for a game |
| `mlb_get_divisions` | Divisions |
| `mlb_get_draft` | Draft picks by year |
| `mlb_get_game_changes` | Recently changed games |
| `mlb_get_game_color` | Color commentary feed |
| `mlb_get_game_color_diff` | Color feed diff patch |
| `mlb_get_game_color_timestamps` | Color feed timestamps |
| `mlb_get_game_content` | Game highlights and editorial |
| `mlb_get_game_diff` | Live feed diff patch |
| `mlb_get_game_pace` | Pace-of-play stats |
| `mlb_get_game_timestamps` | Live feed timestamps |
| `mlb_get_game_uniforms` | Game uniform data |
| `mlb_get_game_win_probability` | Win probability per at-bat |
| `mlb_get_high_low` | Season high/low records |
| `mlb_get_home_run_derby` | Home Run Derby |
| `mlb_get_jobs` | Staff by job type |
| `mlb_get_jobs_datacasters` | Datacaster roster |
| `mlb_get_jobs_official_scorers` | Official scorer roster |
| `mlb_get_jobs_umpires` | Umpire roster |
| `mlb_get_leagues` | League details |
| `mlb_get_live_feed` | Full live game data feed |
| `mlb_get_meta` | API metadata (gameTypes, etc.) |
| `mlb_get_people` | Multiple players by ID |
| `mlb_get_people_changes` | Recent roster changes |
| `mlb_get_person_game_stats` | Player stats in a specific game |
| `mlb_get_play_by_play` | Play-by-play for a game |
| `mlb_get_schedule_postseason_series` | Postseason series |
| `mlb_get_schedule_postseason_tune_in` | Postseason tune-in info |
| `mlb_get_schedule_tied` | Tied/suspended games |
| `mlb_get_season` | Single season metadata |
| `mlb_get_seasons` | Seasons (filtered) |
| `mlb_get_sports` | All sports (MLB, AAA, ...) |
| `mlb_get_sports_players` | All players for a sport |
| `mlb_get_stats` | Individual player stats |
| `mlb_get_stats_streaks` | Hitting/pitching streaks |
| `mlb_get_team_alumni` | Team alumni |
| `mlb_get_team_coaches` | Coaching staff |
| `mlb_get_team_leaders` | Team stat leaders |
| `mlb_get_team_personnel` | Front-office personnel |
| `mlb_get_team_stats` | Team aggregate stats |
| `mlb_get_team_uniforms` | Team uniform catalog |
| `mlb_get_teams` | All teams |
| `mlb_get_teams_affiliates` | Minor league affiliates |
| `mlb_get_teams_history` | Historical team records |
| `mlb_get_teams_stats` | League-wide team stats |
| `mlb_get_umpire_games` | Umpire game assignments |
| `mlb_get_venue` | Venue details |

</details>

## ✨ Features

| Feature              | Description                                               |
| -------------------- | --------------------------------------------------------- |
| MCP stdio server     | Works with any MCP-compatible client out of the box       |
| Two-tier tools       | 11 composed (intent-focused) + 54 auto-generated (raw)    |
| Auto-generated       | `mcpgen` reads the mlb-sdk OpenAPI spec at build time     |
| Typed responses      | Composed tools return structured JSON via [mlb-sdk][]     |
| Full API coverage    | Every MLB Stats API endpoint is available as a tool       |
| Idiomatic Go         | Cobra CLI, functional options, context propagation        |

## 💡 How it works

This server wraps the [mlb-sdk][] Go library, which in turn wraps MLB's
public Stats API (`statsapi.mlb.com`). The composed tools call the typed
SDK through a `Driver` interface; the raw tools call the API directly with
parameters derived from the embedded OpenAPI spec.

## 📖 Documentation

See the [Development][] guide for prerequisites, setup, and conventions.
See the [Contributing][] guide before submitting a PR.

## ⚖️ Copyright notice

This package and its author are not affiliated with MLB or any MLB team. This
module is a typed Go client for MLB's public Stats API. Use of MLB data is
subject to the notice posted at
<http://gdx.mlb.com/components/copyright.txt>.

## 📄 License

The [MIT][] License.

[mcp]: https://modelcontextprotocol.io
[mlb-sdk]: https://github.com/retr0h/mlb-sdk
[MIT]: LICENSE
[Development]: docs/development.md
[Contributing]: docs/contributing.md
