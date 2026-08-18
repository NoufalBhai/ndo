package config

import "github.com/NoufalBhai/ndo/internal/recipe"

// Merge combines central and local recipe sets. On a name collision, the
// local entry wholesale-replaces the central one — no field-level
// patching (DESIGN.md §7: partial overrides are out of scope for v1).
func Merge(central, local RecipeFile) RecipeFile {
	merged := make(map[string]recipe.Recipe, len(central.Recipes)+len(local.Recipes))
	for name, r := range central.Recipes {
		merged[name] = r
	}
	for name, r := range local.Recipes {
		merged[name] = r
	}
	return RecipeFile{Recipes: merged}
}
