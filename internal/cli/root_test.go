package cli

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/config"
	"github.com/green-threads/ndo/internal/recipe"
)

func TestIsUnderCompletionCmd(t *testing.T) {
	h := newTestRoot(t)

	// Regression case: cmd.Name() alone (e.g. "bash") doesn't match
	// "completion" — the check must walk parents, or `ndo completion bash`
	// (which runs as a real subprocess every time a shell sources it)
	// would itself trigger the one-time completion-setup prompt.
	completionCmd, _, err := h.cmd.Find([]string{"completion"})
	if err != nil {
		t.Fatal(err)
	}
	bashCmd, _, err := h.cmd.Find([]string{"completion", "bash"})
	if err != nil {
		t.Fatal(err)
	}
	addCmd, _, err := h.cmd.Find([]string{"add"})
	if err != nil {
		t.Fatal(err)
	}

	if !isUnderCompletionCmd(completionCmd) {
		t.Error("isUnderCompletionCmd(completion) = false, want true")
	}
	if !isUnderCompletionCmd(bashCmd) {
		t.Error("isUnderCompletionCmd(completion bash) = false, want true")
	}
	if isUnderCompletionCmd(addCmd) {
		t.Error("isUnderCompletionCmd(add) = true, want false")
	}
}

func TestCompletionOffersRecipeNamesAtFirstPosition(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"pf":   {Command: "cat {{file}}", Params: []string{"file"}},
		"open": {Command: "code {{path}}", Params: []string{"path"}},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	got, directive := h.cmd.ValidArgsFunction(h.cmd, []string{}, "")
	want := []string{"open", "pf"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("completions = %v, want %v", got, want)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want ShellCompDirectiveNoFileComp", directive)
	}
}

func TestCompletionOffersVarsGroupForDeclaredParam(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{
		Recipes: map[string]recipe.Recipe{
			"pf": {Command: "cat {{file}}", Params: []string{"file"}},
		},
		Vars: map[string]map[string]string{
			"file": {"file1.txt": "path/to/file1.txt", "file2.txt": "path/to/file2.txt"},
		},
	}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	got, directive := h.cmd.ValidArgsFunction(h.cmd, []string{"pf"}, "")
	want := []string{"file1.txt", "file2.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("completions = %v, want %v", got, want)
	}
	if directive != cobra.ShellCompDirectiveDefault {
		t.Fatalf("directive = %v, want ShellCompDirectiveDefault (vars are additive, file completion stays available)", directive)
	}
}

func TestCompletionFallsBackToFileCompWhenNoVarsGroupMatches(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{Recipes: map[string]recipe.Recipe{
		"open": {Command: "code {{path}}", Params: []string{"path"}},
	}}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	got, directive := h.cmd.ValidArgsFunction(h.cmd, []string{"open"}, "")
	if got != nil {
		t.Fatalf("completions = %v, want nil", got)
	}
	if directive != cobra.ShellCompDirectiveDefault {
		t.Fatalf("directive = %v, want ShellCompDirectiveDefault", directive)
	}
}

func TestCompletionUnknownRecipeReturnsNoFileComp(t *testing.T) {
	h := newTestRoot(t)

	got, directive := h.cmd.ValidArgsFunction(h.cmd, []string{"nope"}, "")
	if got != nil {
		t.Fatalf("completions = %v, want nil", got)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want ShellCompDirectiveNoFileComp", directive)
	}
}

func TestCompletionBeyondDeclaredParamsFallsBackToDefault(t *testing.T) {
	h := newTestRoot(t)
	central := config.RecipeFile{
		Recipes: map[string]recipe.Recipe{
			"pf": {Command: "cat {{file}}", Params: []string{"file"}},
		},
		Vars: map[string]map[string]string{
			"file": {"file1.txt": "path/to/file1.txt"},
		},
	}
	if err := config.SaveRecipeFile(centralFilePathForTest(h), central); err != nil {
		t.Fatal(err)
	}

	// "pf" only declares one param, so completing the second positional
	// arg has nothing to match against — should fall back to plain file
	// completion rather than re-offering vars.file.
	got, directive := h.cmd.ValidArgsFunction(h.cmd, []string{"pf", "file1.txt"}, "")
	if got != nil {
		t.Fatalf("completions = %v, want nil", got)
	}
	if directive != cobra.ShellCompDirectiveDefault {
		t.Fatalf("directive = %v, want ShellCompDirectiveDefault", directive)
	}
}
