// Copyright (c) 2026 John Dewey
//
// SPDX-License-Identifier: MIT

// Package version is the single source of build identity for
// `mlb-mcp version`. goreleaser stamps the ldflag targets at link
// time; `go run` / `go build` of a working tree leaves them empty
// (caarlos0/go-version backfills devel defaults from
// debug.ReadBuildInfo).
package version

import goversion "github.com/caarlos0/go-version"

//nolint:revive
var (
	Version   = ""
	Commit    = ""
	TreeState = ""
	Date      = ""
	BuiltBy   = ""
)

// BuildInfo returns the build identity for `mlb-mcp version`.
func BuildInfo() goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails(
			"mlb-mcp",
			"MCP server for the MLB Stats API.\n",
			"https://github.com/retr0h/mlb-mcp",
		),
		func(i *goversion.Info) {
			if Commit != "" {
				i.GitCommit = Commit
			}
			if TreeState != "" {
				i.GitTreeState = TreeState
			}
			if Date != "" {
				i.BuildDate = Date
			}
			if Version != "" {
				i.GitVersion = Version
			}
			if BuiltBy != "" {
				i.BuiltBy = BuiltBy
			}
		},
	)
}
