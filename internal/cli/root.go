// Package cli wires ndo's cobra commands to internal/config and
// internal/recipe. Config is loaded once by the caller (main.go) and
// threaded through explicitly — no package-level globals.
package cli

import (
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/config"
	"github.com/green-threads/ndo/internal/recipe"
)

// Deps bundles what commands need to resolve and run recipes. Built once
// in main.go from the real environment; tests construct their own with
// temp dirs and in-memory streams.
type Deps struct {
	NDOHome string
	Cwd     string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

// Global flags shared across commands.
var (
	flagVerbose bool
	flagDryRun  bool
)

// NewRootCmd builds the root command and registers all subcommands.
func NewRootCmd(version string, deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "ndo",
		Short:         "ndo — a CLI-first, centrally+locally layered task runner",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
		// Default invocation: `ndo <recipe-name> [args...]`. Cobra falls
		// back to the root's own RunE whenever args[0] doesn't match a
		// registered subcommand name, which is what makes plain recipe
		// names like `ndo open ./file.txt` work without a `run` prefix.
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runRecipe(deps, cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], args[1:])
		},
	}

	root.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "print extra diagnostic output, including recipe source file")
	root.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print the resolved, interpolated command without executing")

	root.AddCommand(
		newAddCmd(deps),
		newListCmd(deps),
		newEditCmd(deps),
		newRemoveCmd(deps),
		newInitCmd(deps),
		newVarCmd(deps),
	)

	return root
}

func centralFilePath(ndoHome string) string {
	return filepath.Join(ndoHome, "central.toml")
}

// resolved bundles the three views over recipes a command typically needs:
// the raw central set, the raw local set, and their merge. localPath is
// "" if no local file was found.
type resolved struct {
	Central   config.RecipeFile
	Local     config.RecipeFile
	Merged    config.RecipeFile
	LocalPath string
}

func loadResolved(deps Deps) (resolved, error) {
	central, err := config.LoadRecipeFile(centralFilePath(deps.NDOHome))
	if err != nil {
		return resolved{}, err
	}

	localPath, found := config.SearchUpward(deps.Cwd, config.LocalFileName())
	local := config.RecipeFile{Recipes: map[string]recipe.Recipe{}}
	if found {
		local, err = config.LoadRecipeFile(localPath)
		if err != nil {
			return resolved{}, err
		}
	} else {
		localPath = ""
	}

	return resolved{
		Central:   central,
		Local:     local,
		Merged:    config.Merge(central, local),
		LocalPath: localPath,
	}, nil
}
