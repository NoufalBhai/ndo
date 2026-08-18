package config

import (
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/NoufalBhai/ndo/internal/recipe"
)

// RecipeFile is the schema shared by central.toml and .ndo.toml.
//
// Vars are named lookup tables (e.g. vars.folder.work = "D:\learn\work")
// that recipe param binding checks by param name: if an arg matches a key
// in the vars group named after its param, the looked-up value is
// substituted before interpolation. See internal/cli's runRecipe.
type RecipeFile struct {
	Recipes map[string]recipe.Recipe     `toml:"recipes"`
	Vars    map[string]map[string]string `toml:"vars,omitempty"`
}

// LoadRecipeFile reads a recipe file at path. A missing file is not an
// error — it returns an empty RecipeFile.
func LoadRecipeFile(path string) (RecipeFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RecipeFile{Recipes: map[string]recipe.Recipe{}, Vars: map[string]map[string]string{}}, nil
	}
	if err != nil {
		return RecipeFile{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var rf RecipeFile
	if err := toml.Unmarshal(data, &rf); err != nil {
		return RecipeFile{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	if rf.Recipes == nil {
		rf.Recipes = map[string]recipe.Recipe{}
	}
	if rf.Vars == nil {
		rf.Vars = map[string]map[string]string{}
	}
	return rf, nil
}

// SaveRecipeFile writes rf to path, creating the parent directory if
// needed.
func SaveRecipeFile(path string, rf RecipeFile) error {
	data, err := toml.Marshal(rf)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
