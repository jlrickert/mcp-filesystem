package cmd

import (
	"context"
	"io"
	"os"

	"log/slog"

	"github.com/jlrickert/go-std/pkg"
	"github.com/jlrickert/mcp-filesystem/mcpfs"
)

type Cli struct {
	Services *mcpfs.Services

	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type globalFlags struct {
	cfgPath   string
	logFile   string
	logLevel  string
	flagDebug bool
}

type state struct {
	cli *Cli

	// // Internal state
	teardown func() error

	app    *mcpfs.App
	logger *slog.Logger

	flags globalFlags
}

func (s *state) Clock() std.Clock {
	if s.cli.Services.Clock != nil {
		return s.cli.Services.Clock
	}
	return &std.OsClock{}
}

func (s *state) Env() std.Env {
	if s.cli.Services.Env != nil {
		return s.cli.Services.Env
	}
	return &std.OsEnv{}
}

func (s *state) InOrStdin() io.Reader {
	if s.cli.In != nil {
		return s.cli.In
	}
	return os.Stdin
}

func (s *state) OutOrStdout() io.Writer {
	if s.cli.Out != nil {
		return s.cli.Out
	}
	return os.Stdout
}

func (s *state) ErrOrStderr() io.Writer {
	if s.cli.Err != nil {
		return s.cli.Err
	}
	return os.Stderr
}

// parseLevel maps common textual log level names to slog.Level. Defaults to info.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug", "DEBUG", "Debug":
		return slog.LevelDebug
	case "warn", "warning", "WARN", "Warning":
		return slog.LevelWarn
	case "error", "ERROR", "Error":
		return slog.LevelError
	case "info", "INFO", "Info", "":
		return slog.LevelInfo
	default:
		// Unknown -> default to info
		return slog.LevelInfo
	}
}

func (c *Cli) Run(ctx context.Context, args []string) error {
	if c.Services == nil {
		c.Services = mcpfs.NewDefaultServices()
	}
	cmd := c.newRootCmd()
	cmd.SetArgs(args)
	cmd.SetIn(c.In)
	cmd.SetOut(c.Out)
	cmd.SetErr(c.Err)
	return cmd.ExecuteContext(ctx)
}
