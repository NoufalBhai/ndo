package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveNDOHome(t *testing.T) {
	t.Run("NDO_HOME set", func(t *testing.T) {
		t.Setenv("NDO_HOME", filepath.Join("custom", "path"))
		got, err := ResolveNDOHome()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != filepath.Join("custom", "path") {
			t.Fatalf("ResolveNDOHome() = %q, want %q", got, filepath.Join("custom", "path"))
		}
	})

	t.Run("NDO_HOME unset falls back to ~/.ndo", func(t *testing.T) {
		t.Setenv("NDO_HOME", "")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no home dir available: %v", err)
		}
		got, err := ResolveNDOHome()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(home, ".ndo")
		if got != want {
			t.Fatalf("ResolveNDOHome() = %q, want %q", got, want)
		}
	})
}

func TestSearchUpward(t *testing.T) {
	t.Run("file in cwd", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, ".ndo.toml")
		if err := os.WriteFile(target, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		got, ok := SearchUpward(dir, ".ndo.toml")
		if !ok || got != target {
			t.Fatalf("SearchUpward() = (%q, %v), want (%q, true)", got, ok, target)
		}
	})

	t.Run("file several levels up", func(t *testing.T) {
		root := t.TempDir()
		target := filepath.Join(root, ".ndo.toml")
		if err := os.WriteFile(target, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		nested := filepath.Join(root, "a", "b", "c")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		got, ok := SearchUpward(nested, ".ndo.toml")
		if !ok || got != target {
			t.Fatalf("SearchUpward() = (%q, %v), want (%q, true)", got, ok, target)
		}
	})

	t.Run("no local file anywhere", func(t *testing.T) {
		dir := t.TempDir()
		_, ok := SearchUpward(dir, ".ndo.toml")
		if ok {
			t.Fatalf("SearchUpward() found a file that shouldn't exist")
		}
	})
}

func TestLocalFileName(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("NDO_LOCAL_FILE", "")
		if got := LocalFileName(); got != DefaultLocalFileName {
			t.Fatalf("LocalFileName() = %q, want %q", got, DefaultLocalFileName)
		}
	})
	t.Run("override", func(t *testing.T) {
		t.Setenv("NDO_LOCAL_FILE", "recipes.toml")
		if got := LocalFileName(); got != "recipes.toml" {
			t.Fatalf("LocalFileName() = %q, want %q", got, "recipes.toml")
		}
	})
}

func TestEnsureLocalFile(t *testing.T) {
	dir := t.TempDir()
	path, err := EnsureLocalFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to be created: %v", err)
	}

	// second call should find the same file, not create a duplicate.
	path2, err := EnsureLocalFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path2 != path {
		t.Fatalf("EnsureLocalFile() second call = %q, want %q", path2, path)
	}
}
