package cli

import (
	"reflect"
	"testing"

	"github.com/green-threads/ndo/internal/config"
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

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []string
		wantErr bool
	}{
		{name: "simple", in: "vim", want: []string{"vim"}},
		{name: "with args", in: "code --wait", want: []string{"code", "--wait"}},
		{
			name: "quoted path with spaces",
			in:   `"C:\Program Files\Editor\editor.exe" --wait`,
			want: []string{`C:\Program Files\Editor\editor.exe`, "--wait"},
		},
		{name: "empty", in: "", wantErr: true},
		{name: "whitespace only", in: "   ", wantErr: true},
		{name: "unterminated quote", in: `"C:\no closing quote`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitCommand(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("splitCommand(%q) = %v, want an error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitCommand(%q) unexpected error: %v", tt.in, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitCommand(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
