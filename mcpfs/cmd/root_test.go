package cmd_test

import ()

// func TestServeCommand_PrintStarted_NoStdin(t *testing.T) {
// 	f := NewFixture(t)
// 	defer f.Teardown()
//
// 	// Optionally set an env var
// 	f.WithEnv("HOME", "/tmp")
//
// 	stdout, stderr, err := f.Run([]string{"serve"}, "")
// 	if err != nil {
// 		t.Fatalf("serve returned error: %v, stderr=%s", err, stderr)
// 	}
// 	if !strings.Contains(stdout, "started") {
// 		t.Fatalf("expected stdout to contain 'started', got: %q", stdout)
// 	}
// }
//
// func TestServeCommand_ReadsStdinAndPrints(t *testing.T) {
// 	f := NewFixture(t)
// 	defer f.Teardown()
//
// 	stdin := "hello-from-stdin\n"
// 	stdout, stderr, err := f.Run([]string{"serve"}, stdin)
// 	if err != nil {
// 		t.Fatalf("serve returned error: %v, stderr=%s", err, stderr)
// 	}
// 	if !strings.Contains(stdout, "stdin: hello-from-stdin") {
// 		t.Fatalf("expected stdout to contain stdin echo, got: %q", stdout)
// 	}
// 	if !strings.Contains(stdout, "started") {
// 		t.Fatalf("expected stdout to contain 'started', got: %q", stdout)
// 	}
// }
//
// func TestFlagsOverrideConfig(t *testing.T) {
// 	f := NewFixture(t)
// 	defer f.Teardown()
//
// 	stdout, stderr, err := f.Run([]string{"--foo", "overridden", "--log-level", "debug", "version"}, "")
// 	if err != nil {
// 		t.Fatalf("version command returned error: %v, stderr=%s", err, stderr)
// 	}
// 	if !strings.Contains(stdout, "foo: overridden") {
// 		t.Fatalf("expected stdout to show overridden foo, got: %q", stdout)
// 	}
// 	if !strings.Contains(stdout, "log_level: debug") {
// 		t.Fatalf("expected stdout to show log_level debug, got: %q", stdout)
// 	}
// }
//
// func TestConfigFlagLoadsFile(t *testing.T) {
// 	f := NewFixture(t)
// 	defer f.Teardown()
//
// 	cfgPath, err := f.WithConfigFile(`{"log_level":"debug","foo":"fromfile"}`)
// 	if err != nil {
// 		t.Fatalf("WithConfigFile failed: %v", err)
// 	}
//
// 	stdout, stderr, err := f.Run([]string{"--config", cfgPath, "version"}, "")
// 	if err != nil {
// 		t.Fatalf("version command returned error: %v, stderr=%s", err, stderr)
// 	}
// 	if !strings.Contains(stdout, "foo: fromfile") {
// 		t.Fatalf("expected stdout to show foo from file, got: %q", stdout)
// 	}
// 	if !strings.Contains(stdout, "log_level: debug") {
// 		t.Fatalf("expected stdout to show log_level debug, got: %q", stdout)
// 	}
// }
