// Package cli wires ndo's cobra commands to internal/config and
// internal/recipe. Config is loaded once by the caller (main.go) and
// threaded through explicitly — no package-level globals.
package cli

import (
	"io"
	"path/filepath"
	"sort"

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" || cmd.Name() == "__complete" || cmd.Name() == "__completeNoDesc" || isUnderCompletionCmd(cmd) {
				return nil
			}
			maybeOfferCompletionSetup(cmd.Root(), deps, cmd.ErrOrStderr())
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runRecipe(deps, cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], args[1:])
		},
		ValidArgsFunction: recipeCompletionFunc(deps),
	}

	root.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "print extra diagnostic output, including recipe source file")
	root.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print the resolved, interpolated command without executing")

	// ndo provides its own "completion" command tree (with install/
	// uninstall alongside the per-shell generators), so cobra's default
	// one is disabled to avoid a naming collision.
	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(
		newAddCmd(deps),
		newListCmd(deps),
		newEditCmd(deps),
		newRemoveCmd(deps),
		newInitCmd(deps),
		newVarCmd(deps),
		newCompletionCmd(deps),
	)

	return root
}

// isUnderCompletionCmd reports whether cmd is the "completion" command
// itself or one of its descendants (bash/zsh/fish/powershell/install/
// uninstall). Used to skip the one-time completion-setup prompt for
// anything already related to completion — including a generator
// subcommand like `ndo completion bash`, which runs as a real subprocess
// every time a shell sources it, and must never itself trigger a prompt.
func isUnderCompletionCmd(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "completion" {
			return true
		}
	}
	return false
}

// recipeCompletionFunc drives shell tab-completion for the bare
// `ndo <name> [args...]` invocation:
//   - at the recipe-name position, it offers every resolved recipe name;
//   - at a param position, if that param has a matching `vars` group, it
//     offers that group's keys alongside normal file completion (vars are
//     an additive shortcut, so completion never hides the option to type
//     a literal value instead).
func recipeCompletionFunc(deps Deps) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		res, err := loadResolved(deps)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		if len(args) == 0 {
			names := make([]string, 0, len(res.Merged.Recipes))
			for name := range res.Merged.Recipes {
				names = append(names, name)
			}
			sort.Strings(names)
			return names, cobra.ShellCompDirectiveNoFileComp
		}

		r, ok := res.Merged.Recipes[args[0]]
		if !ok {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		paramIdx := len(args) - 1
		if paramIdx >= len(r.Params) {
			return nil, cobra.ShellCompDirectiveDefault
		}
		group := res.Merged.Vars[r.Params[paramIdx]]
		if len(group) == 0 {
			return nil, cobra.ShellCompDirectiveDefault
		}
		keys := make([]string, 0, len(group))
		for k := range group {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys, cobra.ShellCompDirectiveDefault
	}
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
