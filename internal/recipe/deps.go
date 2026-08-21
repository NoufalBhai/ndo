package recipe

import (
	"fmt"
	"strings"
)

// ResolveDependencies returns the order recipes must run in to satisfy
// name's dependency graph: every recipe reachable through name's (and its
// dependencies') depends field, exactly once each, in depth-first
// declared order, followed by name itself.
//
// Dependencies never receive CLI arguments — only the top-level recipe
// being invoked does — so any recipe reachable purely as a dependency
// that declares required params is a hard configuration error, caught
// here before anything runs rather than surfacing as a confusing
// "missing required argument" at execution time.
func ResolveDependencies(recipes map[string]Recipe, name string) ([]string, error) {
	if _, ok := recipes[name]; !ok {
		return nil, fmt.Errorf("no such recipe: %s", name)
	}

	var order []string
	done := make(map[string]bool)
	onPath := make(map[string]bool)
	var path []string

	var visit func(n string, isRoot bool) error
	visit = func(n string, isRoot bool) error {
		if done[n] {
			return nil
		}
		if onPath[n] {
			return fmt.Errorf("dependency cycle: %s -> %s", strings.Join(path, " -> "), n)
		}

		r, ok := recipes[n]
		if !ok {
			return fmt.Errorf("recipe %s depends on %s, which does not exist", path[len(path)-1], n)
		}
		if !isRoot && len(r.Params) > 0 {
			return fmt.Errorf("recipe %s has required params, so it can't be used as a dependency (dependencies never receive arguments)", n)
		}

		onPath[n] = true
		path = append(path, n)
		for _, dep := range r.Depends {
			if err := visit(dep, false); err != nil {
				return err
			}
		}
		path = path[:len(path)-1]
		onPath[n] = false

		done[n] = true
		order = append(order, n)
		return nil
	}

	if err := visit(name, true); err != nil {
		return nil, err
	}
	return order, nil
}
