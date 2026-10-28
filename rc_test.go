package rc

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeProfiler(t *testing.T, name string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\necho \"args: $*\"\nexit ${FAKE_EXIT:-0}\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func rc(args ...string) (string, string, error) {
	var out, errout bytes.Buffer
	err := Run(context.Background(), "1.2.3", args, &out, &errout)
	return out.String(), errout.String(), err
}

func TestRootListsProfileAndReportsItsVersion(t *testing.T) {
	out, _, err := rc()
	if err != nil || !strings.Contains(out, "profile") {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, _, _ = rc("--version"); out != "rc 1.2.3\n" {
		t.Fatalf("version output %q", out)
	}
}

func TestProfilePassesArgumentsToThePathExecutable(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	out, _, err := rc("profile", "ruby", "generate", "--name", "orders")
	if err != nil || out != "args: generate --name orders\n" {
		t.Fatalf("%v\n%q", err, out)
	}
}

func TestProfileFindsPublishedExecutableNamesAndEnvironmentOverrides(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "runtimeconditions-typescript-profiler"))
	if out, _, err := rc("profile", "typescript", "--version"); err != nil || out != "args: --version\n" {
		t.Fatalf("%v\n%q", err, out)
	}
	dir := fakeProfiler(t, "custom")
	t.Setenv("RC_PYTHON_PROFILER_BIN", filepath.Join(dir, "custom"))
	if out, _, err := rc("profile", "python", "generate"); err != nil || out != "args: generate\n" {
		t.Fatalf("%v\n%q", err, out)
	}
}

func TestProfileReturnsTheProfilerExitStatus(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	t.Setenv("FAKE_EXIT", "3")
	var exit *ExitError
	if _, _, err := rc("profile", "ruby"); !errors.As(err, &exit) || exit.Code != 3 {
		t.Fatalf("exit status not propagated: %v", err)
	}
}

func TestProfileReportsAMissingOrInvalidProfiler(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, _, err := rc("profile", "ruby"); err == nil || !strings.Contains(err.Error(), "rc-profiler-ruby") {
		t.Fatalf("missing profiler not explained: %v", err)
	}
	if _, _, err := rc("profile", "../ruby"); err == nil || !strings.Contains(err.Error(), "invalid language") {
		t.Fatalf("path-like language accepted: %v", err)
	}
}
