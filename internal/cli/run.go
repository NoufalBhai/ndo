package cli

import (
	"fmt"
	"io"

	"github.com/NoufalBhai/ndo/internal/recipe"
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
// binds args positionally, and executes the result — or just prints it if
// --dry-run was passed.
func runRecipe(deps Deps, stdout, stderr io.Writer, name string, args []string) error {
	res, err := loadResolved(deps)
	if err != nil {
		return err
	}

	r, ok := res.Merged.Recipes[name]
	if !ok {
		return fmt.Errorf("no such recipe: %s (see `ndo list`)", name)
	}

	command, err := r.Bind(args)
	if err != nil {
		return fmt.Errorf("recipe %s: %w", name, err)
	}

	if flagVerbose {
		source := "central"
		if _, isLocal := res.Local.Recipes[name]; isLocal {
			source = res.LocalPath
		}
		fmt.Fprintf(stderr, "# %s (from %s)\n", name, source)
	}

	if flagDryRun {
		fmt.Fprintln(stdout, command)
		return nil
	}

	shell := recipe.DefaultShell()
	code, err := recipe.Execute(shell, command, deps.Stdin, stdout, stderr)
	if err != nil {
		return fmt.Errorf("running recipe %s: %w", name, err)
	}
	if code != 0 {
		return &ExitError{Code: code}
	}
	return nil
}
