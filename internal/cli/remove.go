package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/config"
)

func newRemoveCmd(deps Deps) *cobra.Command {
	var local bool

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a recipe from a specific file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(deps, cmd.OutOrStdout(), args[0], local)
		},
	}

	cmd.Flags().BoolVar(&local, "local", false, "remove from the nearest local .ndo.toml instead of the central file")

	return cmd
}

func runRemove(deps Deps, stdout io.Writer, name string, local bool) error {
	var targetPath string
	if local {
		path, ok := config.SearchUpward(deps.Cwd, config.LocalFileName())
		if !ok {
			return fmt.Errorf("no local recipe file found (searched upward from %s)", deps.Cwd)
		}
		targetPath = path
	} else {
		targetPath = centralFilePath(deps.NDOHome)
	}

	rf, err := config.LoadRecipeFile(targetPath)
	if err != nil {
		return err
	}
	if _, exists := rf.Recipes[name]; !exists {
		return fmt.Errorf("recipe %q not found in %s", name, targetPath)
	}

	delete(rf.Recipes, name)
	if err := config.SaveRecipeFile(targetPath, rf); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "removed %q from %s\n", name, targetPath)
	return nil
}
