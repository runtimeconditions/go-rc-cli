package rc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// CoreSchema is the core profile schema version this release of rc targets.
const CoreSchema = "0.4.0"

var languageName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

var publishedProfilers = map[string]string{"go": "go-rc-profiler", "python": "runtimeconditions-python-profiler", "typescript": "runtimeconditions-typescript-profiler"}

func profileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "profile <language> [args...]",
		Short: "Profile an application with an installed language profiler",
		Long: `Runs the profiler for <language> and passes the remaining arguments to it unchanged.

The profiler is the executable named by RC_<LANGUAGE>_PROFILER_BIN, else rc-profiler-<language> on PATH.
The go, python and typescript profilers are also found under their own published executable names.`,
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
				return c.Help()
			}
			path, err := profilerPath(args[0])
			if err != nil {
				return err
			}
			if runsProfiler(args[1:]) {
				if err = checkCoreSchema(c.Context(), path, args[0], c.ErrOrStderr()); err != nil {
					return err
				}
			}
			cmd := exec.CommandContext(c.Context(), path, args[1:]...)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, c.OutOrStdout(), c.ErrOrStderr()
			var exit *exec.ExitError
			if err = cmd.Run(); errors.As(err, &exit) {
				return &ExitError{exit.ExitCode()}
			}
			return err
		},
	}
}

func profilerPath(language string) (string, error) {
	if !languageName.MatchString(language) {
		return "", fmt.Errorf("invalid language %q", language)
	}
	env := "RC_" + strings.ToUpper(strings.ReplaceAll(language, "-", "_")) + "_PROFILER_BIN"
	names := []string{"rc-profiler-" + language}
	if published, ok := publishedProfilers[language]; ok {
		names = append(names, published)
	}
	if override := os.Getenv(env); override != "" {
		names = []string{override}
	}
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no %s profiler found: install rc-profiler-%s on PATH or set %s", language, language, env)
}

func runsProfiler(args []string) bool {
	return len(args) > 0 && !strings.HasPrefix(args[0], "-") && args[0] != "version"
}

func checkCoreSchema(ctx context.Context, path, language string, stderr io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var info struct {
		Version     string   `json:"version"`
		CoreSchemas []string `json:"coreSchemas"`
	}
	out, err := exec.CommandContext(ctx, path, "--version", "--json").Output()
	if err != nil || json.Unmarshal(out, &info) != nil || len(info.CoreSchemas) == 0 {
		fmt.Fprintf(stderr, "note: the %s profiler does not report its core schemas; skipping the compatibility check\n", language)
		return nil
	}
	if !slices.Contains(info.CoreSchemas, CoreSchema) {
		return fmt.Errorf("the %s profiler %s supports core schema %s, but this rc targets %s; install a compatible profiler", language, info.Version, strings.Join(info.CoreSchemas, ", "), CoreSchema)
	}
	return nil
}
