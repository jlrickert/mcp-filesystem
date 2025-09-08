package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"log/slog"

	"github.com/jlrickert/go-std/pkg"
	"github.com/jlrickert/mcp-filesystem/mcpfs"
)

// // RunCommand runs a cobra command with controlled env, args and stdin and
// // returns captured stdout/stderr and error. It restores any modified environment
// // after completion.
// func RunCommand(root *cobra.Command, env map[string]string, args []string, stdin string) (string, string, error) {
// 	// Save and restore environment
// 	origEnv := make(map[string]*string)
// 	for k := range env {
// 		if v, ok := os.LookupEnv(k); ok {
// 			ov := v
// 			origEnv[k] = &ov
// 		} else {
// 			origEnv[k] = nil
// 		}
// 		os.Setenv(k, env[k])
// 	}
// 	defer func() {
// 		for k, v := range origEnv {
// 			if v == nil {
// 				os.Unsetenv(k)
// 			} else {
// 				os.Setenv(k, *v)
// 			}
// 		}
// 	}()
//
// 	// Capture output
// 	outBuf := &bytes.Buffer{}
// 	errBuf := &bytes.Buffer{}
// 	root.SetOut(outBuf)
// 	root.SetErr(errBuf)
//
// 	// Set stdin if provided
// 	if stdin != "" {
// 		root.SetIn(strings.NewReader(stdin))
// 	} else {
// 		root.SetIn(io.NopCloser(strings.NewReader("")))
// 	}
//
// 	// Set args and execute
// 	root.SetArgs(args)
// 	_, err := root.ExecuteC()
//
// 	return outBuf.String(), errBuf.String(), err
// }

// Fixture provides a convenient test fixture for command integration tests.
type Fixture struct {
	T        *testing.T
	TempDir  string
	Cfg      *mcpfs.Config
	Logger   *slog.Logger
	Services *mcpfs.Services
}

// NewFixture constructs a fixture with t.TempDir(), base config, a discard logger,
// and a root command wired for testing. Callers should defer f.Teardown().
func NewFixture(t *testing.T) *Fixture {
	td := t.TempDir()
	cfg := &mcpfs.Config{
		LogLevel: "debug",
	}
	logger := std.NewDiscardLogger()

	return &Fixture{
		T:       t,
		TempDir: td,
		Cfg:     cfg,
		Logger:  logger,
		Services: &mcpfs.Services{
			Env:   std.NewTestEnv(filepath.Join(td, "testuser"), "testuser"),
			Clock: &std.TestClock{},
		},
	}
}

// WithEnv sets an environment variable in the fixture (applied when Run is called).
func (f *Fixture) WithEnv(key, val string) *Fixture {
	f.Services.Env.Set(key, val)
	return f
}

// WithConfigFile writes content to a temp file in the fixture TempDir and returns its path.
func (f *Fixture) WithConfigFile(content string) (string, error) {
	// Ensure dir exists (t.TempDir ensures it).
	fn := filepath.Join(f.T.TempDir(), mcpfs.DefaultConfigFilename)
	if err := os.WriteFile(fn, []byte(content), 0o600); err != nil {
		return "", err
	}
	return fn, nil
}

// // Run runs the root command with fixture.Env merged into env and returns captured output.
// func (f *Fixture) Run(args []string, stdin string) (string, string, error) {
// 	return RunCommand(f.Root, f.Env, args, stdin)
// }

// Teardown is a placeholder to satisfy the interface. No-op currently.
func (f *Fixture) Teardown() {}
