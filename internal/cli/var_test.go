package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/green-threads/ndo/internal/config"
)

func TestVarAddCreatesEntry(t *testing.T) {
	h := newTestRoot(t, "var", "add", "folder", "work", "C:\\Users\\dev\\projects", "--local")
	if err := h.Execute(); err != nil {
		t.Fatalf("var add failed: %v", err)
	}

	path, ok := config.SearchUpward(h.deps.Cwd, config.LocalFileName())
	if !ok {
		t.Fatalf("expected local file to be created")
	}
	rf, err := config.LoadRecipeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := rf.Vars["folder"]["work"]; got != "C:\\Users\\dev\\projects" {
		t.Fatalf("vars.folder.work = %q, want %q", got, "C:\\Users\\dev\\projects")
	}
}

func TestVarAddOverwritesExistingKey(t *testing.T) {
	h := newTestRoot(t)

	first := NewRootCmd("test", h.deps)
	first.SetArgs([]string{"var", "add", "folder", "work", "D:\\old", "--local"})
	if err := first.Execute(); err != nil {
		t.Fatalf("first var add failed: %v", err)
	}

	second := NewRootCmd("test", h.deps)
	second.SetArgs([]string{"var", "add", "folder", "work", "D:\\new", "--local"})
	if err := second.Execute(); err != nil {
		t.Fatalf("second var add failed: %v", err)
	}

	path, _ := config.SearchUpward(h.deps.Cwd, config.LocalFileName())
	rf, err := config.LoadRecipeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := rf.Vars["folder"]["work"]; got != "D:\\new" {
		t.Fatalf("vars.folder.work = %q, want overwritten value %q", got, "D:\\new")
	}
}

func TestVarRemove(t *testing.T) {
	h := newTestRoot(t)

	add := NewRootCmd("test", h.deps)
	add.SetArgs([]string{"var", "add", "folder", "work", "C:\\Users\\dev\\projects", "--local"})
	if err := add.Execute(); err != nil {
		t.Fatalf("var add failed: %v", err)
	}

	rm := NewRootCmd("test", h.deps)
	rm.SetArgs([]string{"var", "remove", "folder", "work", "--local"})
	if err := rm.Execute(); err != nil {
		t.Fatalf("var remove failed: %v", err)
	}

	path, _ := config.SearchUpward(h.deps.Cwd, config.LocalFileName())
	rf, err := config.LoadRecipeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := rf.Vars["folder"]["work"]; exists {
		t.Fatalf("expected vars.folder.work to be removed")
	}
}

func TestVarListGroupedOutput(t *testing.T) {
	h := newTestRoot(t)

	add1 := NewRootCmd("test", h.deps)
	add1.SetArgs([]string{"var", "add", "folder", "work", "C:\\Users\\dev\\projects", "--local"})
	if err := add1.Execute(); err != nil {
		t.Fatalf("var add failed: %v", err)
	}
	add2 := NewRootCmd("test", h.deps)
	add2.SetArgs([]string{"var", "add", "folder", "ndo", "C:\\Users\\dev\\projects\\ndo", "--local"})
	if err := add2.Execute(); err != nil {
		t.Fatalf("var add failed: %v", err)
	}

	list := NewRootCmd("test", h.deps)
	var out bytes.Buffer
	list.SetOut(&out)
	list.SetArgs([]string{"var", "list"})
	if err := list.Execute(); err != nil {
		t.Fatalf("var list failed: %v", err)
	}

	want := "folder:\n    ndo: C:\\Users\\dev\\projects\\ndo\n    work: C:\\Users\\dev\\projects\n"
	if got := out.String(); got != want {
		t.Fatalf("var list output = %q, want %q", got, want)
	}
}

func TestVarRemoveWholeGroup(t *testing.T) {
	h := newTestRoot(t)

	add1 := NewRootCmd("test", h.deps)
	add1.SetArgs([]string{"var", "add", "folder", "work", "C:\\Users\\dev\\projects", "--local"})
	if err := add1.Execute(); err != nil {
		t.Fatalf("var add failed: %v", err)
	}
	add2 := NewRootCmd("test", h.deps)
	add2.SetArgs([]string{"var", "add", "folder", "ndo", "C:\\Users\\dev\\projects\\ndo", "--local"})
	if err := add2.Execute(); err != nil {
		t.Fatalf("var add failed: %v", err)
	}

	rm := NewRootCmd("test", h.deps)
	rm.SetArgs([]string{"var", "remove", "folder", "--local"})
	if err := rm.Execute(); err != nil {
		t.Fatalf("var remove (group) failed: %v", err)
	}

	path, _ := config.SearchUpward(h.deps.Cwd, config.LocalFileName())
	rf, err := config.LoadRecipeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := rf.Vars["folder"]; exists {
		t.Fatalf("expected the whole %q group to be removed, got %+v", "folder", rf.Vars["folder"])
	}
}

func TestVarRemoveWholeGroupNotFoundErrors(t *testing.T) {
	h := newTestRoot(t, "var", "remove", "nope")
	if err := h.Execute(); err == nil {
		t.Fatal("expected error, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v, want mention of 'not found'", err)
	}
}

func TestVarRemoveNotFoundErrors(t *testing.T) {
	h := newTestRoot(t, "var", "remove", "folder", "nope")
	if err := h.Execute(); err == nil {
		t.Fatal("expected error, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v, want mention of 'not found'", err)
	}
}
