package cli

import (
	"testing"

	"github.com/NoufalBhai/ndo/internal/config"
)

func TestTargetFilePathCentral(t *testing.T) {
	h := newTestRoot(t)
	path, err := targetFilePath(h.deps, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != centralFilePath(h.deps.NDOHome) {
		t.Fatalf("targetFilePath() = %q, want central path", path)
	}
}

func TestTargetFilePathLocalCreatesIfMissing(t *testing.T) {
	h := newTestRoot(t)
	path, err := targetFilePath(h.deps, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := config.SearchUpward(h.deps.Cwd, config.LocalFileName()); !ok {
		t.Fatalf("expected local file to exist at %s", path)
	}
}

func TestResolveEditorPrefersSettingsOverEnv(t *testing.T) {
	ndoHome := t.TempDir()
	t.Setenv("EDITOR", "nano")
	if err := config.SaveSettings(ndoHome, config.Settings{Editor: "vim"}); err != nil {
		t.Fatal(err)
	}
	got, err := resolveEditor(ndoHome)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "vim" {
		t.Fatalf("resolveEditor() = %q, want %q", got, "vim")
	}
}

func TestResolveEditorFallsBackToEnv(t *testing.T) {
	ndoHome := t.TempDir()
	t.Setenv("EDITOR", "nano")
	got, err := resolveEditor(ndoHome)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "nano" {
		t.Fatalf("resolveEditor() = %q, want %q", got, "nano")
	}
}

func TestResolveEditorErrorsIfNoneConfigured(t *testing.T) {
	ndoHome := t.TempDir()
	t.Setenv("EDITOR", "")
	if _, err := resolveEditor(ndoHome); err == nil {
		t.Fatal("expected error when no editor is configured")
	}
}
