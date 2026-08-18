package cli

import (
	"strings"
	"testing"

	"github.com/NoufalBhai/ndo/internal/config"
)

func TestInitCreatesLocalFile(t *testing.T) {
	h := newTestRoot(t, "init")
	if err := h.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, ok := config.SearchUpward(h.deps.Cwd, config.LocalFileName()); !ok {
		t.Fatalf("expected .ndo.toml to be created in %s", h.deps.Cwd)
	}
}

func TestInitErrorsIfLocalFileAlreadyExists(t *testing.T) {
	h := newTestRoot(t)

	first := NewRootCmd("test", h.deps)
	first.SetArgs([]string{"init"})
	if err := first.Execute(); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	second := NewRootCmd("test", h.deps)
	second.SetArgs([]string{"init"})
	if err := second.Execute(); err == nil {
		t.Fatal("expected error on second init, got nil")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want mention of 'already exists'", err)
	}
}
