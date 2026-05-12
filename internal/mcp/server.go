// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// Package mcp is the Model Context Protocol server for mlb-mcp — a typed
// Go wrapper around MLB's public Stats API exposed as MCP tools that any
// MCP-compatible LLM agent can call.
//
// Architecturally: this package owns no state beyond the MLB SDK client it
// constructs on startup. It wires each mlb.Client method into an mcp.Tool.
// Stdio transport is the default — an agent (Claude Code, Cursor, …) spawns
// `mlb-mcp mcp start` as a subprocess and pipes JSON-RPC over stdin/stdout;
// when the agent disconnects the process exits cleanly.
//
// All diagnostic output goes to stderr; stdout is the JSON-RPC wire.
package mcp

import (
	"context"
	"log/slog"
	"os"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/retr0h/mlb-sdk/pkg/mlb"
)

// version is stamped into mcp.Implementation so MCP clients see who they're
// talking to.
const version = "0.1.0"

// Config bundles the runtime inputs for the mlb-mcp MCP server. Logger is
// for the server's own diagnostic chatter — never written to stdout, which
// belongs to the JSON-RPC wire. No auth is required; MLB's Stats API is public.
// Driver is optional — when nil, New() creates a real mlb.Client; tests
// inject a fake here.
type Config struct {
	Logger *slog.Logger
	Driver Driver
}

// Server is the mlb-mcp MCP server. Holds a Driver (the narrow consumer
// surface, see session.go) used by every tool handler, plus the underlying
// mcpsdk.Server. Constructed via New; the wire is driven by Run.
//
// client is typed as the Driver interface, not concrete *mlb.Client — this
// file holds the only concrete-type reference (in New, at the assignment site
// where the compiler verifies the structural fit).
type Server struct {
	mcp    *mcpsdk.Server
	client Driver
	logger *slog.Logger
}

// New constructs an MCP server backed by a fresh mlb.Client pointed at the
// public MLB Stats API. Every tool registration happens here so the returned
// Server is fully formed and ready to Run; later mutation is not supported.
func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	client := cfg.Driver
	if client == nil {
		client = mlb.New()
	}

	mcpSrv := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "mlb-mcp",
			Version: version,
		},
		&mcpsdk.ServerOptions{
			Instructions: instructions,
		},
	)

	s := &Server{
		mcp:    mcpSrv,
		client: client,
		logger: cfg.Logger.With(slog.String("subsystem", "mcp")),
	}
	s.registerTools()
	return s
}

// Run wires the MCP server to stdin/stdout and blocks until the transport
// closes (the spawning agent disconnects) or ctx cancels. Returns nil on
// clean shutdown.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info(
		"running",
		slog.String("transport", "stdio"),
	)
	return s.mcp.Run(ctx, &mcpsdk.StdioTransport{})
}

// instructions is the MCP server's self-description, surfaced to the agent
// on initialize. Keep it short and concrete; agents read this to understand
// what the server can do without paging through every tool's description.
const instructions = `mlb-mcp — MLB Stats API over MCP.

This server exposes MLB's public Stats API as composable, intent-focused
tools. Each tool answers one user question directly. No authentication is
required; all data is publicly available.

## Tools at a glance

- scores              — scores and results for any date (defaults to today)
- standings          — division standings for AL, NL, or both leagues
- player_bio         — biographical info for a player by MLB person ID
- team_info          — rich team metadata (venue, league, division) by team ID
- team_roster        — active roster for a team by team ID
- league_leaders     — top players in a stat category (HR, AVG, ERA, …)
- game_detail        — boxscore (team batting/pitching stats) for a gamePk
- game_linescore     — inning-by-inning line for a gamePk
- recent_transactions — recent signings, trades, DFAs; defaults to today
- free_agents        — free-agent declarations and signings for a season
- postseason_schedule — postseason game schedule for a season

## Key concepts

- Team IDs: use MLB's canonical numeric IDs (e.g. 119 = Los Angeles Dodgers,
  147 = New York Yankees, 111 = Boston Red Sox).
- Person IDs: use MLB's canonical numeric player IDs (e.g. 660271 = Shohei Ohtani).
- Game PKs: the unique game identifier returned by scores and
  postseason_schedule; used by game_detail and game_linescore.
- Stat categories for league_leaders: 'homeRuns', 'battingAverage',
  'strikeOuts', 'era', 'wins', 'saves', 'rbi', 'stolenBases'.
- Season: all season-bearing tools default to the current year when omitted.`
