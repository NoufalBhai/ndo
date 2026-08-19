package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/green-threads/ndo/internal/config"
)

func TestAppendMarkedBlockCreatesFileAndBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "rc")

	if err := appendMarkedBlock(path, "source <(ndo completion bash)"); err != nil {
		t.Fatalf("appendMarkedBlock() error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{completionMarkerBegin, "source <(ndo completion bash)", completionMarkerEnd} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("file content = %q, want it to contain %q", got, want)
		}
	}
}

func TestAppendMarkedBlockIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rc")
	if err := os.WriteFile(path, []byte("existing content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := appendMarkedBlock(path, "line one"); err != nil {
		t.Fatal(err)
	}
	if err := appendMarkedBlock(path, "line one"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(got), completionMarkerBegin); n != 1 {
		t.Fatalf("marker appears %d times, want exactly 1 (re-running should be a no-op)", n)
	}
	if !strings.Contains(string(got), "existing content") {
		t.Fatalf("pre-existing content was clobbered: %q", got)
	}
}

func TestIsTerminalFileFalseForRegularFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "notatty")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if isTerminalFile(f) {
		t.Fatal("isTerminalFile() = true for a regular file, want false")
	}
}

// Regression test: os.DevNull (/dev/null, or NUL on Windows) is itself a
// character device, so an os.ModeCharDevice-based check false-positives
// as "terminal" for it — this previously caused maybeOfferCompletionSetup
// to run for real against redirected/empty stdin. isTerminalFile must use
// a real isatty check instead.
func TestIsTerminalFileFalseForDevNull(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if isTerminalFile(f) {
		t.Fatal("isTerminalFile() = true for os.DevNull, want false — this is the exact bug that let the completion prompt run against non-interactive stdin")
	}
}

func TestMaybeOfferCompletionSetupSkipsWhenAlreadyAnswered(t *testing.T) {
	h := newTestRoot(t)
	if err := config.SaveSettings(h.deps.NDOHome, config.Settings{CompletionPromptAnswered: true}); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	maybeOfferCompletionSetup(h.cmd, h.deps, &stderr)

	if stderr.Len() != 0 {
		t.Fatalf("expected no output when already answered, got %q", stderr.String())
	}
}

func TestMaybeOfferCompletionSetupSkipsWhenStdinNotATerminal(t *testing.T) {
	// newTestRoot wires deps.Stdin to a bytes.Reader, not *os.File, so this
	// exercises the same non-interactive path a real script/CI run takes.
	h := newTestRoot(t)

	var stderr bytes.Buffer
	maybeOfferCompletionSetup(h.cmd, h.deps, &stderr)

	if stderr.Len() != 0 {
		t.Fatalf("expected no output for non-terminal stdin, got %q", stderr.String())
	}
	settings, err := config.LoadSettings(h.deps.NDOHome)
	if err != nil {
		t.Fatal(err)
	}
	if settings.CompletionPromptAnswered {
		t.Fatal("CompletionPromptAnswered should stay false when the prompt was never actually shown")
	}
}
