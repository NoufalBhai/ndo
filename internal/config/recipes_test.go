package config

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/green-threads/ndo/internal/recipe"
)

func TestRecipeFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "central.toml")

	want := RecipeFile{
		Recipes: map[string]recipe.Recipe{
			"open": {Command: "code {{file}}", Params: []string{"file"}},
			"test": {Command: "go test ./...", Description: "run tests"},
		},
		Vars: map[string]map[string]string{
			"folder": {"work": "C:\\Users\\dev\\projects"},
		},
	}

	if err := SaveRecipeFile(path, want); err != nil {
		t.Fatalf("SaveRecipeFile() error: %v", err)
	}
	got, err := LoadRecipeFile(path)
	if err != nil {
		t.Fatalf("LoadRecipeFile() error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadRecipeFile() = %+v, want %+v", got, want)
	}
}

func TestLoadRecipeFileMissingFileIsNotError(t *testing.T) {
	dir := t.TempDir()
	got, err := LoadRecipeFile(filepath.Join(dir, "nope.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Recipes) != 0 || len(got.Vars) != 0 {
		t.Fatalf("LoadRecipeFile() = %+v, want empty", got)
	}
}
