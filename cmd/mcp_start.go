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
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	mcppkg "github.com/retr0h/mlb-mcp/internal/mcp"
)

var mcpStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Run the mlb-mcp MCP server over stdio",
	Long: `Speaks Model Context Protocol on stdin/stdout — the transport
agents (Claude Code, Cursor, …) expect when they spawn a server as a
subprocess. Blocks until the agent disconnects (the typical MCP lifecycle).

Logs go to stderr only — stdout is the JSON-RPC wire and writing anything
else there would corrupt the protocol. No authentication required; MLB's
Stats API is public.

  mlb-mcp mcp start`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		log := logger.With(slog.String("subsystem", "mcp.start"))
		log.Info("starting")

		ctx, cancel := signal.NotifyContext(
			cmd.Context(),
			syscall.SIGINT,
			syscall.SIGTERM,
		)
		defer cancel()

		s := mcppkg.New(mcppkg.Config{
			Logger: logger,
		})
		var srv mcpRunner = s
		if err := srv.Run(ctx); err != nil {
			return fmt.Errorf("mcp start: %w", err)
		}
		return nil
	},
}
