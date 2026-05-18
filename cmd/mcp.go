// Copyright (c) 2026 John Dewey
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// mcpRunner is the narrow consumer-seam interface this cobra command depends
// on, declared per the osapi-io pattern. Concrete *mcp.Server satisfies it
// structurally — the compiler verifies at the assignment site in mcp_start.go.
type mcpRunner interface {
	Run(ctx context.Context) error
}

// mcpCmd is the parent for `mlb-mcp mcp` — the Model Context Protocol server
// surface. Spawned per agent session (Claude Code / Cursor / any MCP-aware
// host), it serves MLB Stats API data over stdio JSON-RPC.
//
// Subcommands:
//
//	start    open an MCP server over stdio (the typical agent-spawn lifecycle)
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run a Model Context Protocol server for the MLB Stats API",
	Long: `Exposes MLB Stats API data as MCP tools over stdio JSON-RPC.

The MCP server is spawned by an agent (Claude Code, Cursor, …) per session
— when the agent disconnects the process exits. No authentication required;
MLB's Stats API is public.

Configure your agent (example for Claude Code) to spawn:

  mlb-mcp mcp start

…and it gets tools for schedule, standings, person, team, stats_leaders,
and linescore.`,
}

func init() {
	mcpCmd.AddCommand(mcpStartCmd)
	rootCmd.AddCommand(mcpCmd)
}
