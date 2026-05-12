// Copyright (c) 2026 John Dewey
//
// SPDX-License-Identifier: MIT

// Package cli provides the themed CLI banner for mlb-mcp.
package cli

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	top = "█▀▄▀█ █░░ █▄▄   █▀▄▀█ █▀▀ █▀█"
	bot = "█░▀░█ █▄▄ █▄█   █░▀░█ █▄▄ █▀▀"
)

// Banner returns the two-line block-text banner in MLB blue (#002D72).
// Falls back to uncolored text when the writer is not a TTY.
func Banner(w io.Writer) string {
	if !isTTY(w) {
		return top + "\n" + bot + "\n"
	}
	// MLB blue: RGB(0, 45, 114) = #002D72
	blue := "\033[38;2;0;45;114m"
	reset := "\033[0m"
	return fmt.Sprintf("%s%s%s\n%s%s%s\n", blue, top, reset, blue, bot, reset)
}

func isTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
