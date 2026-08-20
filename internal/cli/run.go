package cli

import (
	"fmt"
	"io"

	"github.com/green-threads/ndo/internal/recipe"
)

// ExitError signals that a recipe's command ran to completion but exited
// non-zero. main.go translates this into the process's own exit code
// instead of printing it as a generic error.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("recipe exited with code %d", e.Code)
}

// runRecipe resolves name against the merged central+local recipe set,
// resolves its dependency chain (name's depends, deduped and topologically
// ordered, per internal/recipe.ResolveDependencies), and runs each recipe
// in order — binding CLI args only against the top-level recipe, since
// dependencies never receive arguments. Stops at the first non-zero exit
// or error rather than running the rest of the chain. If --dry-run was
// passed, prints every resolved command in the chain instead of running
// any of them.
func runRecipe(deps Deps, stdout, stderr io.Writer, name string, args []string) error {
	res, err := loadResolved(deps)
	if err != nil {
		return err
	}

	if _, ok := res.Merged.Recipes[name]; !ok {
		return fmt.Errorf("no such recipe: %s (see `ndo list`)", name)
	}

	order, err := recipe.ResolveDependencies(res.Merged.Recipes, name)
	if err != nil {
		return err
	}

	for _, n := range order {
		r := res.Merged.Recipes[n]

		var recipeArgs []string
		if n == name {
			recipeArgs = resolveVars(res.Merged.Vars, r.Params, args)
		}

		command, err := r.Bind(recipeArgs)
		if err != nil {
			return fmt.Errorf("recipe %s: %w", n, err)
		}

		if flagVerbose {
			source := "central"
			if _, isLocal := res.Local.Recipes[n]; isLocal {
				source = res.LocalPath
			}
			fmt.Fprintf(stderr, "# %s (from %s)\n", n, source)
		}

		if flagDryRun {
			fmt.Fprintln(stdout, command)
			continue
		}

		shell := recipe.DefaultShell()
		code, err := recipe.Execute(shell, command, deps.Stdin, stdout, stderr)
		if err != nil {
			return fmt.Errorf("running recipe %s: %w", n, err)
		}
		if code != 0 {
			return &ExitError{Code: code}
		}
	}
	return nil
}

// resolveVars expands args positionally against vars: if the arg at index
// i matches a key in the vars group named after params[i], it's replaced
// with the looked-up value. Args with no matching param, or no matching
// group/key, pass through unchanged — this makes the lookup an opt-in
// convenience, never a breaking requirement.
func resolveVars(vars map[string]map[string]string, params, args []string) []string {
	resolved := make([]string, len(args))
	copy(resolved, args)
	for i, paramName := range params {
		if i >= len(resolved) {
			break
		}
		if value, ok := vars[paramName][resolved[i]]; ok {
			resolved[i] = value
		}
	}
	return resolved
}
