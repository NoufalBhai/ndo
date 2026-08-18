package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/NoufalBhai/ndo/internal/config"
	"github.com/NoufalBhai/ndo/internal/recipe"
)

func TestListMergedGolden(t *testing.T) {
	h := newTestRoot(t)

	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"build": {Command: "go build ./..."},
		"test":  {Command: "go test ./...", Description: "run tests"},
	}}
	if err := config.SaveRecipeFile(centralFilePath(h.deps.NDOHome), central); err != nil {
		t.Fatal(err)
	}

	local := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"test": {Command: "go test ./... -race", Description: "run tests with race detector"},
	}}
	if err := config.SaveRecipeFile(filepath.Join(h.deps.Cwd, ".ndo.toml"), local); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("list failed: %v", err)
	}

	golden := filepath.Join("..", "..", "testdata", "cli", "list", "merged.golden")
	assertGolden(t, golden, out.String())
}

func assertGolden(t *testing.T, path, got string) {
	t.Helper()
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden file %s: %v (run with UPDATE_GOLDEN=1 to create it)", path, err)
	}
	if got != string(want) {
		t.Fatalf("output mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, string(want))
	}
}
