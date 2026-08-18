package recipe

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteExitCode(t *testing.T) {
	shell := DefaultShell()

	code, err := Execute(shell, "exit 0", nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	code, err = Execute(shell, "exit 3", nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
}

func TestExecuteStdoutCapture(t *testing.T) {
	var stdout bytes.Buffer
	_, err := Execute(DefaultShell(), "echo hello", nil, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "hello" {
		t.Fatalf("stdout = %q, want %q", got, "hello")
	}
}
