package config

import (
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/NoufalBhai/ndo/internal/recipe"
)

// RecipeFile is the schema shared by central.toml and .ndo.toml.
type RecipeFile struct {
	Recipes map[string]recipe.Recipe `toml:"recipes"`
}

// LoadRecipeFile reads a recipe file at path. A missing file is not an
// error — it returns an empty RecipeFile.
func LoadRecipeFile(path string) (RecipeFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RecipeFile{Recipes: map[string]recipe.Recipe{}}, nil
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
