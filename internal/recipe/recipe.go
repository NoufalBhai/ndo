package recipe

import (
	"fmt"
	"regexp"
	"strings"
)

// Recipe is a named, shell-executable command with optional parameters.
//
// Depends names other recipes to run first — see ResolveDependencies in
// deps.go for ordering, dedup, and cycle-detection rules. Recipes used
// purely as dependencies (never invoked directly) can't declare required
// Params, since dependencies never receive CLI arguments.
type Recipe struct {
	Command     string   `toml:"command"`
	Params      []string `toml:"params,omitempty"`
	Depends     []string `toml:"depends,omitempty"`
	Description string   `toml:"description,omitempty"`
}

var tokenPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*(\|\s*raw)?\s*\}\}`)

// Bind resolves positional args against the recipe's declared params and
// interpolates them into Command, producing the final shell command line.
//
// Missing required params are a hard error (never silently rendered
// empty). Args beyond the declared params are appended raw (quoted) to the
// end of the command, covering ad-hoc passthrough like `ndo test -- -run
// TestFoo`.
func (r Recipe) Bind(args []string) (string, error) {
	bound := make(map[string]string, len(r.Params))
	for i, name := range r.Params {
		if i >= len(args) {
			return "", fmt.Errorf("missing required argument: %s", name)
		}
		bound[name] = args[i]
	}

	var interpErr error
	interpolated := tokenPattern.ReplaceAllStringFunc(r.Command, func(match string) string {
		groups := tokenPattern.FindStringSubmatch(match)
		name, raw := groups[1], groups[2] != ""

		value, ok := bound[name]
		if !ok {
			interpErr = fmt.Errorf("recipe references undeclared parameter: %s", name)
			return match
		}
		if raw {
			return value
		}
		return QuoteArg(value)
	})
	if interpErr != nil {
		return "", interpErr
	}

	if leftover := args[len(r.Params):]; len(leftover) > 0 {
		quoted := make([]string, len(leftover))
		for i, v := range leftover {
			quoted[i] = QuoteArg(v)
		}
		interpolated = interpolated + " " + strings.Join(quoted, " ")
	}

	return interpolated, nil
}
