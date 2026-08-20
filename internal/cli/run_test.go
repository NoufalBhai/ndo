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

func TestRunExecutesDependenciesBeforeRecipe(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"deploy": {Command: "echo deploy", Depends: []string{"build"}},
		"build":  {Command: "echo build"},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"deploy"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	want := "build\ndeploy\n"
	if got := strings.ReplaceAll(out.String(), "\r\n", "\n"); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunDryRunPrintsWholeDependencyChain(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"deploy": {Command: "echo deploy", Depends: []string{"build"}},
		"build":  {Command: "echo build"},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--dry-run", "deploy"})
	if err := root.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	want := "echo build\necho deploy\n"
	if got := out.String(); got != want {
		t.Fatalf("dry-run stdout = %q, want %q", got, want)
	}
}

func TestRunStopsChainOnDependencyFailure(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"deploy": {Command: "echo deploy", Depends: []string{"build"}},
		"build":  {Command: "exit 3"},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"deploy"})
	err := root.Execute()

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %v, want *ExitError", err)
	}
	if exitErr.Code != 3 {
		t.Fatalf("exit code = %d, want 3", exitErr.Code)
	}
	if got := out.String(); got != "" {
		t.Fatalf("stdout = %q, want empty — deploy should never run after build fails", got)
	}
}

func TestRunDependencyGraphErrorSurfacesBeforeAnyExecution(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"deploy": {Command: "echo deploy", Depends: []string{"missing"}},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"deploy"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %v, want mention of the missing dependency", err)
	}
	if got := out.String(); got != "" {
		t.Fatalf("stdout = %q, want empty — nothing should run", got)
	}
}

func centralFilePathForTest(h *rootHarness) string {
	return centralFilePath(h.deps.NDOHome)
}
