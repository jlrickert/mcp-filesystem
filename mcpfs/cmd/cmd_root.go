package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jlrickert/go-std/pkg"
	"github.com/jlrickert/mcp-filesystem/mcpfs"
	"github.com/spf13/cobra"
)

// NewRootCmd constructs the cobra root command wired with a minimal serve and
// version subcommands. It accepts a pointer to a Config and a logger. The
// function configures persistent flags and a PersistentPreRunE that will
// reload config from disk (if --config provided), apply flag overrides, and
// recreate the application instance used by subcommands.
//
// Notes:
//   - The root command's default Out/Err are set to bytes.Buffer so tests can
//     capture output by default.
func (cli *Cli) newRootCmd() *cobra.Command {
	flags := globalFlags{}

	s := &state{cli: cli, teardown: func() error { return nil }}
	root := &cobra.Command{
		Version: mcpfs.Version,
		Use:     "mcpfs",
		Short:   "mcpfs CLI (minimal root command)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// base defaults
			cfg := &mcpfs.Config{
				LogLevel: "info",
			}

			// If a config file path is provided, load it (propagate errors).
			if flags.cfgPath != "" {
				loaded, err := mcpfs.ReadAndParseConfig(flags.cfgPath)
				if err != nil {
					return err
				}
				// Use loaded file as base
				cfg = loaded
			}

			var logfile string
			if flags.logFile != "" {
				logfile = flags.logFile
			} else {
				path, err := std.UserStatePath(mcpfs.AppName, cli.Services.Env)
				if err == nil {
					logfile = filepath.Join(path, "log.json")
				}
			}

			// Default to discard; may be replaced with a file or stderr fallback.
			logR := io.Discard
			if logfile != "" {
				// Ensure the directory for the logfile exists before attempting to open it.
				dir := filepath.Dir(logfile)
				if dir != "" {
					if err := os.MkdirAll(dir, 0o755); err != nil {
						// If we can't create the directory, fall back to stderr so logging still works.
						logR = os.Stderr
					} else {
						// Try to open (or create) the requested log file for append.
						f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
						if err != nil {
							// On failure, fall back to stderr so the caller still has a valid writer.
							logR = os.Stderr
						} else {
							// Use the opened file as the log writer and ensure it's closed at teardown.
							logR = f
							td := s.teardown
							s.teardown = func() error {
								var errs []error
								if td != nil {
									if err := td(); err != nil {
										errs = append(errs, err)
									}
								}
								if err := f.Close(); err != nil {
									errs = append(errs, err)
								}
								if len(errs) == 0 {
									return nil
								}
								return errors.Join(errs...)
							}
						}
					}
				}
			}

			logger := std.NewLogger(std.LoggerConfig{
				Version: mcpfs.Version,
				Out:     logR,
				Level:   parseLevel(flags.logLevel),
				JSON:    true,
			})
			logger.Info("initialized")
			cmd.SetContext(std.ContextWithLogger(cmd.Context(), logger))

			s.logger = logger
			s.cli = cli
			s.app = mcpfs.NewApp(cfg, logger, cli.Services)
			cmd.SetIn(s.InOrStdin())
			cmd.SetOut(s.OutOrStdout())
			cmd.SetErr(s.ErrOrStderr())
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "WAWER")
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return s.teardown()
		},
	}

	root.RegisterFlagCompletionFunc("log-level", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		choices := []cobra.Completion{
			"debug",
			"info",
			"warn",
			"error",
		}
		if toComplete == "" {
			return choices, cobra.ShellCompDirectiveNoFileComp
		}
		var out []cobra.Completion
		for _, c := range choices {
			// manual prefix check to avoid adding a new import
			if len(toComplete) <= len(c) && c[:len(toComplete)] == toComplete {
				out = append(out, c)
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	})

	// persistent flags
	root.PersistentFlags().StringVarP(&flags.cfgPath, "config", "c", "", "optional JSON config file path")
	root.PersistentFlags().StringVar(&flags.logLevel, "log-level", "info", "override log level (debug/info/warn/error)")
	root.PersistentFlags().StringVar(&flags.logFile, "logfile", "", "description")

	// Add subcommands
	root.AddCommand(s.newStdioCmd())
	root.AddCommand(s.newStdio2Cmd())

	return root
}

func (s *state) newStdioCmd() *cobra.Command {
	return &cobra.Command{
		Use: "stdio",
		// Aliases: []string{"-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// w := cmd.OutOrStdout()
			// fmt.Fprintln(w, "hello world")
			return nil
		},
	}
}

func (s *state) newStdio2Cmd() *cobra.Command {
	return &cobra.Command{
		Use: "stdio2",
		// Aliases: []string{"-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// w := cmd.OutOrStdout()
			// fmt.Fprintln(w, "hello world")
			return nil
		},
	}
}
