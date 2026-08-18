//go:build windows

package recipe

import "testing"

func TestQuoteArg(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", `""`},
		{"simple", "main.go", "main.go"},
		{"spaces", "hello world", `"hello world"`},
		{"embedded double quote", `say "hi"`, `"say ""hi"""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QuoteArg(tt.input); got != tt.want {
				t.Fatalf("QuoteArg(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultShell(t *testing.T) {
	got := DefaultShell()
	if len(got) != 2 || got[0] != "cmd" || got[1] != "/C" {
		t.Fatalf("DefaultShell() = %v, want [cmd /C]", got)
	}
}
