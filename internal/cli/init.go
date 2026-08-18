package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/config"
)

func newInitCmd(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create an empty .ndo.toml in the current directory, if none exists in the tree",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(deps, cmd.OutOrStdout())
		},
	}
	return cmd
}

func runInit(deps Deps, stdout io.Writer) error {
	if existing, ok := config.SearchUpward(deps.Cwd, config.LocalFileName()); ok {
		return fmt.Errorf("a local recipe file already exists: %s", existing)
	}

	path, err := config.EnsureLocalFile(deps.Cwd)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "created %s\n", path)
	return nil
}
