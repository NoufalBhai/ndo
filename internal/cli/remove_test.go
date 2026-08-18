package cli

import (
	"strings"
	"testing"

	"github.com/green-threads/ndo/internal/config"
)

func TestRemoveLocal(t *testing.T) {
	h := newTestRoot(t)
	deps := h.deps

	add := NewRootCmd("test", deps)
	add.SetArgs([]string{"add", "hello", "echo hi", "--local"})
	if err := add.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	rm := NewRootCmd("test", deps)
	rm.SetArgs([]string{"remove", "hello", "--local"})
	if err := rm.Execute(); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	path, ok := config.SearchUpward(deps.Cwd, config.LocalFileName())
	if !ok {
		t.Fatalf("expected local file to still exist")
	}
	rf, err := config.LoadRecipeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := rf.Recipes["hello"]; exists {
		t.Fatalf("expected recipe to be removed")
	}
}

func TestRemoveNotFoundErrors(t *testing.T) {
	h := newTestRoot(t, "remove", "nope")
	if err := h.Execute(); err == nil {
		t.Fatal("expected error, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v, want mention of 'not found'", err)
	}
}
