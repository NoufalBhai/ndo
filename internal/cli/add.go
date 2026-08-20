package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/config"
	"github.com/green-threads/ndo/internal/recipe"
)

func newAddCmd(deps Deps) *cobra.Command {
	var params []string
	var dependsOn []string
	var local bool
	var description string

	cmd := &cobra.Command{
		Use:   "add <name> <command>",
		Short: "Add a new recipe",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(deps, cmd.OutOrStdout(), args[0], args[1], params, dependsOn, description, local)
		},
	}

	cmd.Flags().StringArrayVar(&params, "param", nil, "declared parameter name, in positional order (repeatable)")
	cmd.Flags().StringArrayVar(&dependsOn, "depends", nil, "recipe to run first, in order (repeatable); see docs/recipe-format.md#dependencies")
	cmd.Flags().BoolVar(&local, "local", false, "write to the nearest local .ndo.toml instead of the central file")
	cmd.Flags().StringVar(&description, "desc", "", "recipe description, shown in `ndo list`")

	return cmd
}

func runAdd(deps Deps, stdout io.Writer, name, command string, params, dependsOn []string, description string, local bool) error {
	targetPath, err := targetFilePath(deps, local)
	if err != nil {
		return err
	}

	rf, err := config.LoadRecipeFile(targetPath)
	if err != nil {
		return err
	}
	if rf.Recipes == nil {
		rf.Recipes = map[string]recipe.Recipe{}
	}
	if _, exists := rf.Recipes[name]; exists {
		return fmt.Errorf("recipe %q already exists in %s (use `ndo edit` to change it)", name, targetPath)
	}

	rf.Recipes[name] = recipe.Recipe{
		Command:     command,
		Params:      params,
		Depends:     dependsOn,
		Description: description,
	}

	if err := config.SaveRecipeFile(targetPath, rf); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "added %q to %s\n", name, targetPath)
	return nil
}

// targetFilePath resolves where `add`/`remove`/`edit` should read/write:
// the nearest local file (creating one if --local and none exists), or
// the central file.
func targetFilePath(deps Deps, local bool) (string, error) {
	if local {
		return config.EnsureLocalFile(deps.Cwd)
	}
	if err := config.EnsureNDOHome(deps.NDOHome); err != nil {
		return "", err
	}
	return centralFilePath(deps.NDOHome), nil
}
