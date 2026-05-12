// Copyright (c) 2026 John Dewey
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/retr0h/mlb-mcp/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of mlb-mcp",
	Run: func(_ *cobra.Command, _ []string) {
		info := version.BuildInfo()
		jsonOut, _ := info.JSONString()
		fmt.Println(jsonOut)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
