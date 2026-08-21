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

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		env       map[string]string
		wantShell string
		wantOK    bool
	}{
		{
			name:      "windows git bash",
			goos:      "windows",
			env:       map[string]string{"MSYSTEM": "MINGW64", "SHELL": "/usr/bin/bash"},
			wantShell: "bash",
			wantOK:    true,
		},
		{
			// Regression case for the bug this test suite exists to catch:
			// PSModulePath is inherited by child processes, so Git Bash
			// launched from a PowerShell terminal still has it set. MSYSTEM
			// must win, or completion gets wired into $PROFILE instead of
			// .bashrc for someone sitting in bash right now.
			name:      "windows git bash launched from a powershell terminal",
			goos:      "windows",
			env:       map[string]string{"MSYSTEM": "MINGW64", "PSModulePath": `C:\Program Files\WindowsPowerShell\Modules`},
			wantShell: "bash",
			wantOK:    true,
		},
		{
			name:      "windows powershell",
			goos:      "windows",
			env:       map[string]string{"PSModulePath": `C:\Program Files\WindowsPowerShell\Modules`},
			wantShell: "powershell",
			wantOK:    true,
		},
		{
			name:   "windows cmd.exe (undetectable)",
			goos:   "windows",
			env:    map[string]string{},
			wantOK: false,
		},
		{
			name:      "unix zsh",
			goos:      "linux",
			env:       map[string]string{"SHELL": "/bin/zsh"},
			wantShell: "zsh",
			wantOK:    true,
		},
		{
			name:      "unix fish",
			goos:      "darwin",
			env:       map[string]string{"SHELL": "/opt/homebrew/bin/fish"},
			wantShell: "fish",
			wantOK:    true,
		},
		{
			name:      "unix bash default",
			goos:      "linux",
			env:       map[string]string{"SHELL": "/bin/bash"},
			wantShell: "bash",
			wantOK:    true,
		},
		{
			name:      "unix unset SHELL defaults to bash",
			goos:      "linux",
			env:       map[string]string{},
			wantShell: "bash",
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell, ok := detectShell(tt.goos, tt.env)
			if ok != tt.wantOK {
				t.Fatalf("detectShell() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && shell != tt.wantShell {
				t.Fatalf("detectShell() shell = %q, want %q", shell, tt.wantShell)
			}
		})
	}
}

func TestRemoveMarkedBlockRemovesOnlyTheMarkedBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rc")
	if err := os.WriteFile(path, []byte("existing content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendMarkedBlock(path, "source <(ndo completion bash)"); err != nil {
		t.Fatal(err)
	}

	removed, err := removeMarkedBlock(path)
	if err != nil {
		t.Fatalf("removeMarkedBlock() error: %v", err)
	}
	if !removed {
		t.Fatal("removeMarkedBlock() removed = false, want true")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "existing content\n" {
		t.Fatalf("file content = %q, want exactly the pre-existing content restored", got)
	}
}

func TestRemoveMarkedBlockNoOpWhenAbsent(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "nope")
	removed, err := removeMarkedBlock(path)
	if err != nil {
		t.Fatalf("removeMarkedBlock() error on missing file: %v", err)
	}
	if removed {
		t.Fatal("removeMarkedBlock() removed = true for a missing file, want false")
	}

	path2 := filepath.Join(dir, "rc")
	if err := os.WriteFile(path2, []byte("no marker here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	removed, err = removeMarkedBlock(path2)
	if err != nil {
		t.Fatalf("removeMarkedBlock() error: %v", err)
	}
	if removed {
		t.Fatal("removeMarkedBlock() removed = true for a file with no marker, want false")
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

func TestPreferredPowerShellBinary(t *testing.T) {
	tests := []struct {
		name         string
		psModulePath string
		want         string
	}{
		{
			name:         "windows powershell 5.1 default PSModulePath",
			psModulePath: `C:\Users\dev\Documents\WindowsPowerShell\Modules;C:\Program Files\WindowsPowerShell\Modules;C:\WINDOWS\system32\WindowsPowerShell\v1.0\Modules`,
			want:         "powershell",
		},
		{
			name:         "pwsh (powershell 7+) default PSModulePath",
			psModulePath: `C:\Users\dev\Documents\PowerShell\Modules;C:\Program Files\PowerShell\Modules;c:\program files\powershell\7\Modules`,
			want:         "pwsh",
		},
		{
			// Regression case for the bug this test exists to catch: always
			// trying pwsh first — regardless of which edition the live
			// session invoking `ndo completion install` actually is —
			// resolved the wrong $PROFILE on a machine with both editions
			// installed, writing the completion block to a profile the
			// live session never sources.
			name:         "case-insensitive match",
			psModulePath: `c:\users\dev\documents\windowspowershell\modules`,
			want:         "powershell",
		},
		{
			name:         "empty PSModulePath falls back to pwsh",
			psModulePath: "",
			want:         "pwsh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preferredPowerShellBinary(tt.psModulePath); got != tt.want {
				t.Fatalf("preferredPowerShellBinary(%q) = %q, want %q", tt.psModulePath, got, tt.want)
			}
		})
	}
}
