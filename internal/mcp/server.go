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
type Config struct {
	Logger *slog.Logger
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

	var client Driver = mlb.New()

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

This server exposes MLB's public Stats API as MCP tools. No authentication
is required; all data is publicly available.

## Getting started

1. Call schedule to fetch today's games or filter by team and date.
2. Call standings to see division standings for AL (103) or NL (104).
3. Call person to look up a player by their MLB person ID.
4. Call team to retrieve team metadata including venue, league, and division.
5. Call stats_leaders to find the current or historical stat leaders.
6. Call linescore to get the inning-by-inning breakdown for a specific game.

## Key concepts

- League IDs: 103 = American League, 104 = National League.
- Team IDs: use MLB's canonical numeric IDs (e.g. 119 = Los Angeles Dodgers).
- Person IDs: use MLB's canonical numeric player IDs (e.g. 660271 = Shohei Ohtani).
- Game PKs: the unique game identifier returned by schedule; used by linescore.
- Dates: YYYY-MM-DD format for all date parameters.
- Hydrate: a comma-separated string to expand sub-objects (e.g. "league,division").`
