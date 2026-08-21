package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/green-threads/ndo/internal/config"
)

func TestAddCreatesLocalRecipe(t *testing.T) {
	h := newTestRoot(t, "add", "hello", "echo {{name}}", "--param", "name", "--local")
	if err := h.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	path, ok := config.SearchUpward(h.deps.Cwd, config.LocalFileName())
	if !ok {
		t.Fatalf("expected local file to be created")
	}
	rf, err := config.LoadRecipeFile(path)
	if err != nil {
		t.Fatalf("LoadRecipeFile: %v", err)
	}
	r, ok := rf.Recipes["hello"]
	if !ok {
		t.Fatalf("expected recipe %q to exist", "hello")
	}
	if r.Command != "echo {{name}}" || len(r.Params) != 1 || r.Params[0] != "name" {
		t.Fatalf("unexpected recipe: %+v", r)
	}
}

func TestAddWithDependsFlagRepeatable(t *testing.T) {
	h := newTestRoot(t, "add", "deploy", "./deploy.sh", "--depends", "build", "--depends", "lint", "--local")
	if err := h.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	path, ok := config.SearchUpward(h.deps.Cwd, config.LocalFileName())
	if !ok {
		t.Fatalf("expected local file to be created")
	}
	rf, err := config.LoadRecipeFile(path)
	if err != nil {
		t.Fatalf("LoadRecipeFile: %v", err)
	}
	r, ok := rf.Recipes["deploy"]
	if !ok {
		t.Fatalf("expected recipe %q to exist", "deploy")
	}
	if want := []string{"build", "lint"}; !reflect.DeepEqual(r.Depends, want) {
		t.Fatalf("Depends = %v, want %v", r.Depends, want)
	}
}

func TestAddDuplicateInTargetFileErrors(t *testing.T) {
	h := newTestRoot(t, "add", "hello", "echo hi", "--local")
	if err := h.Execute(); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	// Re-run against the same deps (same cwd -> same local file) so the
	// second add collides with the first.
	root := NewRootCmd("test", h.deps)
	root.SetArgs([]string{"add", "hello", "echo hi again", "--local"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error on duplicate recipe name, got nil")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want mention of 'already exists'", err)
	}
}

func TestAddRejectsBuiltinCommandName(t *testing.T) {
	for _, name := range []string{"list", "add", "edit", "remove", "init", "var", "completion", "update"} {
		t.Run(name, func(t *testing.T) {
			h := newTestRoot(t, "add", name, "echo hi", "--local")
			err := h.Execute()
			if err == nil {
				t.Fatalf("expected error adding a recipe named %q, got nil", name)
			}
			if !strings.Contains(err.Error(), "built-in ndo command") {
				t.Fatalf("error = %v, want mention of 'built-in ndo command'", err)
			}

			path, ok := config.SearchUpward(h.deps.Cwd, config.LocalFileName())
			if ok {
				rf, loadErr := config.LoadRecipeFile(path)
				if loadErr == nil {
					if _, exists := rf.Recipes[name]; exists {
						t.Fatalf("recipe %q was written despite the error", name)
					}
				}
			}
		})
	}
}

func TestAddDefaultsToCentral(t *testing.T) {
	h := newTestRoot(t, "add", "hello", "echo hi")
	if err := h.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	rf, err := config.LoadRecipeFile(centralFilePath(h.deps.NDOHome))
	if err != nil {
		t.Fatalf("LoadRecipeFile: %v", err)
	}
	if _, ok := rf.Recipes["hello"]; !ok {
		t.Fatalf("expected recipe to land in central.toml")
	}
}
