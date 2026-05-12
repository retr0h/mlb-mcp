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

// Package cmd contains the mlb-mcp cobra command tree.
package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// logger is the package-level slog logger. Targets stderr so it never
// contaminates the JSON-RPC wire on stdout.
var logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

var rootCmd = &cobra.Command{
	Use:   "mlb-mcp",
	Short: "MLB Stats API MCP server",
	Long: `mlb-mcp is a Model Context Protocol server for the MLB Stats API.

It exposes MLB schedule, standings, player, team, stat-leaders, and linescore
data as MCP tools that any MCP-compatible LLM agent (Claude Code, Cursor, …)
can call directly. No API key required — MLB's Stats API is public.

Run the MCP server over stdio:

  mlb-mcp mcp start`,
	RunE: func(c *cobra.Command, _ []string) error {
		return c.Help()
	},
}

// Execute runs the root command; invoked by main. SilenceUsage suppresses
// the help-text dump on runtime failures where it would just be noise.
func Execute() {
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
