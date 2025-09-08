package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/jlrickert/mcp-filesystem/mcpfs/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Normalize argv: cobra-style CLI runners expect args without the program name.
	// Some task runners accidentally inject the root command name into argv (e.g.
	// ["/path/to/mcpfs", "mcpfs", "completion", "zsh"]). Detect and remove any
	// leading occurrence(s) of the program base name so the CLI sees only the
	// intended subcommand args.
	raw := os.Args
	var args []string
	if len(raw) > 0 {
		args = raw[1:]
		progBase := filepath.Base(raw[0])
		for len(args) > 0 && (args[0] == progBase || args[0] == raw[0]) {
			args = args[1:]
		}
	}

	cli := cmd.Cli{}
	err := cli.Run(ctx, args)

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
