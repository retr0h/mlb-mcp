// Copyright (c) 2026 John Dewey
//
// SPDX-License-Identifier: MIT

// Package cmd contains the mlb-mcp cobra command tree.
package cmd

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// logger is the package-level slog logger, populated from initLogger
// after cobra parses persistent flags. MCP subcommands pass it to the
// server via Config so all diagnostic output targets stderr — stdout is
// reserved for the JSON-RPC wire.
var (
	logger     = slog.New(slog.NewTextHandler(os.Stderr, nil))
	jsonOutput bool
)

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

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "emit logs as JSON")

	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	viper.SetEnvPrefix("mlb_mcp")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func initLogger() {
	level := slog.LevelInfo
	if viper.GetBool("debug") {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if jsonOutput {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
			NoColor:    !term.IsTerminal(int(os.Stderr.Fd())),
		})
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
}
