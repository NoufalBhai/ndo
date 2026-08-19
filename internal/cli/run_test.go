package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/green-threads/ndo/internal/config"
	"github.com/green-threads/ndo/internal/recipe"
)

func TestRunResolvesAndExecutesRecipe(t *testing.T) {
	h := newTestRoot(t)
	local := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"hello": {Command: "echo {{name}}", Params: []string{"name"}},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), local); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"hello", "world"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "world" {
		t.Fatalf("stdout = %q, want %q", got, "world")
	}
}

func TestRunUnknownRecipeErrors(t *testing.T) {
	h := newTestRoot(t, "nope")
	if err := h.Execute(); err == nil {
		t.Fatal("expected error, got nil")
	} else if !strings.Contains(err.Error(), "no such recipe") {
		t.Fatalf("error = %v, want mention of 'no such recipe'", err)
	}
}

func TestRunMissingParamErrors(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"open": {Command: "code {{file}}", Params: []string{"file"}},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	root.SetArgs([]string{"open"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error, got nil")
	} else if !strings.Contains(err.Error(), "missing required argument") {
		t.Fatalf("error = %v, want mention of missing argument", err)
	}
}

func TestRunDryRunPrintsWithoutExecuting(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"hello": {Command: "echo {{name}}", Params: []string{"name"}},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--dry-run", "hello", "world"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got := strings.TrimSpace(out.String()); !strings.Contains(got, "world") {
		t.Fatalf("dry-run output = %q, want it to contain the resolved command", got)
	}
}

func TestRunPropagatesNonZeroExitCode(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"fail": {Command: "exit 7"},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	root.SetArgs([]string{"fail"})
	err := root.Execute()
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %v, want *ExitError", err)
	}
	if exitErr.Code != 7 {
		t.Fatalf("exit code = %d, want 7", exitErr.Code)
	}
}

func TestRunResolvesArgAgainstMatchingVarsGroup(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{
		Recipes: map[string]recipe.Recipe{
			"o": {Command: "echo {{folder}}", Params: []string{"folder"}},
		},
		Vars: map[string]map[string]string{
			"folder": {"work": "C:\\Users\\dev\\projects"},
		},
	}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"o", "work"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "C:\\Users\\dev\\projects" {
		t.Fatalf("stdout = %q, want %q", got, "C:\\Users\\dev\\projects")
	}
}

func TestRunFallsBackToLiteralWhenNoVarsMatch(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{
		Recipes: map[string]recipe.Recipe{
			"o": {Command: "echo {{folder}}", Params: []string{"folder"}},
		},
		Vars: map[string]map[string]string{
			"folder": {"work": "C:\\Users\\dev\\projects"},
		},
	}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"o", "D:\\some\\other\\path"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "D:\\some\\other\\path" {
		t.Fatalf("stdout = %q, want unmatched literal passed through", got)
	}
}

func centralFilePathForTest(h *rootHarness) string {
	return centralFilePath(h.deps.NDOHome)
}
