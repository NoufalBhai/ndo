package config

import "github.com/NoufalBhai/ndo/internal/recipe"

// Merge combines central and local recipe sets. On a name collision, the
// local entry wholesale-replaces the central one — no field-level
// patching (DESIGN.md §7: partial overrides are out of scope for v1).
//
// Vars merge at key level instead: central provides defaults, local can
// add or override individual keys within a group without losing the rest
// of that group's central entries.
func Merge(central, local RecipeFile) RecipeFile {
	recipes := make(map[string]recipe.Recipe, len(central.Recipes)+len(local.Recipes))
	for name, r := range central.Recipes {
		recipes[name] = r
	}
	for name, r := range local.Recipes {
		recipes[name] = r
	}

	vars := make(map[string]map[string]string, len(central.Vars)+len(local.Vars))
	for group, entries := range central.Vars {
		merged := make(map[string]string, len(entries))
		for k, v := range entries {
			merged[k] = v
		}
		vars[group] = merged
	}
	for group, entries := range local.Vars {
		if vars[group] == nil {
			vars[group] = make(map[string]string, len(entries))
		}
		for k, v := range entries {
			vars[group][k] = v
		}
	}

	return RecipeFile{Recipes: recipes, Vars: vars}
}
