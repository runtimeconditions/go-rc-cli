package rc

import "github.com/spf13/cobra"

func bindingsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bindings <command> [args...]",
		Short: "Generate and verify extension bindings",
		Long: `Runs the bindings tool and passes all arguments to it unchanged.

The tool is the executable named by RC_BINDINGS_BIN, else rc-bindings on PATH. It is built from
runtimeconditions/extensions (tooling/extension-bindings/orchestrator, cmd/rc-bindings).`,
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			path, err := findExecutable("RC_BINDINGS_BIN", "rc-bindings")
			if err != nil {
				return err
			}
			return runPlugin(c, path, args)
		},
	}
}
