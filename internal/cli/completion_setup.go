package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/green-threads/ndo/internal/config"
)

const (
	completionMarkerBegin = "# >>> ndo completion >>>"
	completionMarkerEnd   = "# <<< ndo completion <<<"
)

// maybeOfferCompletionSetup asks, once ever per NDO_HOME, whether to
// install shell tab-completion, and records the answer in config.toml so
// it's never asked again. It's a strict no-op unless stdin is a real
// terminal, so it can never block or surprise a script/CI/test run.
func maybeOfferCompletionSetup(root *cobra.Command, deps Deps, stderr io.Writer) {
	settings, err := config.LoadSettings(deps.NDOHome)
	if err != nil || settings.CompletionPromptAnswered {
		return
	}

	stdinFile, ok := deps.Stdin.(*os.File)
	if !ok || !isTerminalFile(stdinFile) {
		return
	}

	fmt.Fprint(stderr, "Enable ndo shell tab-completion? Adds a couple of lines to your shell's startup file so recipe names and vars complete with <TAB>. [Y/n] ")
	line, readErr := bufio.NewReader(stdinFile).ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))

	settings.CompletionPromptAnswered = true
	switch {
	case readErr != nil && answer == "":
		// Nothing could be read at all (e.g. stdin closed mid-prompt) —
		// treat as declined. Never let a failed/absent read fall through
		// to the same branch as "pressed Enter to accept the default".
		fmt.Fprintln(stderr, "\nSkipping. Run `ndo completion install` any time if you change your mind.")
	case answer == "" || answer == "y" || answer == "yes":
		if path, shell, err := installCompletion(root); err != nil {
			fmt.Fprintf(stderr, "Couldn't set up completion automatically: %v\nYou can do it manually — see `ndo completion --help`.\n", err)
		} else {
			fmt.Fprintf(stderr, "Shell completion enabled for %s (added to %s). Open a new terminal for it to take effect.\n", shell, path)
		}
	default:
		fmt.Fprintln(stderr, "Skipping. Run `ndo completion install` any time if you change your mind.")
	}

	if err := config.SaveSettings(deps.NDOHome, settings); err != nil {
		fmt.Fprintf(stderr, "warning: couldn't save completion preference: %v\n", err)
	}
}

// isTerminalFile reports whether f is a real interactive terminal, using
// the actual isatty syscall — unlike a file-mode check, this correctly
// returns false for /dev/null (or NUL on Windows), which is itself a
// character device but never an interactive terminal.
func isTerminalFile(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// detectShell picks which shell integration to use, given the OS and a
// snapshot of the environment. Kept pure (no direct env/OS reads) so the
// detection logic — including the priority order between signals — is
// unit tested without touching real files or environment variables.
func detectShell(goos string, env map[string]string) (shell string, ok bool) {
	if goos == "windows" {
		// Git Bash sets MSYSTEM. Check it (and SHELL) *before*
		// PSModulePath: PSModulePath is inherited by child processes, so
		// Git Bash launched from a PowerShell terminal would otherwise be
		// misdetected as PowerShell itself, writing to the wrong file.
		if env["MSYSTEM"] != "" || strings.Contains(strings.ToLower(env["SHELL"]), "bash") {
			return "bash", true
		}
		if env["PSModulePath"] != "" {
			return "powershell", true
		}
		return "", false
	}

	switch filepath.Base(env["SHELL"]) {
	case "zsh":
		return "zsh", true
	case "fish":
		return "fish", true
	default:
		return "bash", true
	}
}

func currentShellEnv() map[string]string {
	return map[string]string{
		"MSYSTEM":      os.Getenv("MSYSTEM"),
		"SHELL":        os.Getenv("SHELL"),
		"PSModulePath": os.Getenv("PSModulePath"),
	}
}

// installCompletion detects the current shell and wires ndo's completion
// script into it, returning the file that was touched and the shell name.
func installCompletion(root *cobra.Command) (installedTo, shellName string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("resolving home directory: %w", err)
	}

	env := currentShellEnv()
	shell, ok := detectShell(runtime.GOOS, env)
	if !ok {
		return "", "", fmt.Errorf("couldn't detect a supported shell (PowerShell or Git Bash) — cmd.exe isn't supported")
	}

	switch shell {
	case "powershell":
		return installCompletionPowerShell(env["PSModulePath"])
	case "fish":
		return installCompletionFish(root, home)
	case "zsh":
		return installCompletionBashLike(home, "zsh", ".zshrc")
	default:
		return installCompletionBashLike(home, "bash", ".bashrc")
	}
}

// uninstallCompletion detects the current shell the same way
// installCompletion does, and removes whatever it would have installed.
func uninstallCompletion() (path, shellName string, removed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", false, fmt.Errorf("resolving home directory: %w", err)
	}

	env := currentShellEnv()
	shell, ok := detectShell(runtime.GOOS, env)
	if !ok {
		return "", "", false, fmt.Errorf("couldn't detect a supported shell (PowerShell or Git Bash) — cmd.exe isn't supported")
	}

	if shell == "fish" {
		fishPath := filepath.Join(home, ".config", "fish", "completions", "ndo.fish")
		if _, statErr := os.Stat(fishPath); os.IsNotExist(statErr) {
			return fishPath, "fish", false, nil
		}
		if err := os.Remove(fishPath); err != nil {
			return "", "", false, fmt.Errorf("removing %s: %w", fishPath, err)
		}
		return fishPath, "fish", true, nil
	}

	var rcPath string
	switch shell {
	case "powershell":
		rcPath, err = resolvePowerShellProfile(env["PSModulePath"])
		if err != nil {
			return "", "", false, err
		}
	case "zsh":
		rcPath = filepath.Join(home, ".zshrc")
	default:
		rcPath = filepath.Join(home, ".bashrc")
	}

	removed, err = removeMarkedBlock(rcPath)
	return rcPath, shell, removed, err
}

func installCompletionBashLike(home, shellName, rcFileName string) (string, string, error) {
	rcPath := filepath.Join(home, rcFileName)
	line := fmt.Sprintf("source <(ndo completion %s)", shellName)
	if err := appendMarkedBlock(rcPath, line); err != nil {
		return "", "", err
	}
	return rcPath, shellName, nil
}

func installCompletionFish(root *cobra.Command, home string) (string, string, error) {
	dir := filepath.Join(home, ".config", "fish", "completions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("creating %s: %w", dir, err)
	}
	path := filepath.Join(dir, "ndo.fish")
	f, err := os.Create(path)
	if err != nil {
		return "", "", fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()
	if err := root.GenFishCompletion(f, true); err != nil {
		return "", "", fmt.Errorf("generating fish completion: %w", err)
	}
	return path, "fish", nil
}

func installCompletionPowerShell(psModulePath string) (string, string, error) {
	profilePath, err := resolvePowerShellProfile(psModulePath)
	if err != nil {
		return "", "", err
	}
	line := "if (Get-Command ndo -ErrorAction SilentlyContinue) { ndo completion powershell | Out-String | Invoke-Expression }"
	if err := appendMarkedBlock(profilePath, line); err != nil {
		return "", "", err
	}
	return profilePath, "powershell", nil
}

// preferredPowerShellBinary picks which PowerShell binary to query first
// when resolving $PROFILE, based on the invoking session's own
// PSModulePath. Querying the wrong edition (e.g. asking pwsh for $PROFILE
// when the live session running `ndo completion install` is actually
// Windows PowerShell 5.1) writes the completion block to a profile the
// live session never sources — on a machine with both editions installed,
// always trying pwsh first got this wrong whenever the invoking session
// was Windows PowerShell. Windows PowerShell's default PSModulePath
// always includes a "WindowsPowerShell" path segment; PowerShell 7+'s
// (pwsh) default doesn't. Kept pure (just a string check) so the
// preference logic is unit tested without shelling out to either binary.
func preferredPowerShellBinary(psModulePath string) string {
	if strings.Contains(strings.ToLower(psModulePath), "windowspowershell") {
		return "powershell"
	}
	return "pwsh"
}

// resolvePowerShellProfile asks PowerShell itself for $PROFILE rather than
// guessing between the Windows PowerShell 5.1 and PowerShell 7+ default
// paths, since only PowerShell knows which one the running session uses —
// querying preferredPowerShellBinary(psModulePath) first, so it asks the
// edition that actually matches the invoking session before falling back
// to the other one.
func resolvePowerShellProfile(psModulePath string) (string, error) {
	preferred := preferredPowerShellBinary(psModulePath)
	fallback := "powershell"
	if preferred == "powershell" {
		fallback = "pwsh"
	}
	for _, bin := range []string{preferred, fallback} {
		out, err := exec.Command(bin, "-NoProfile", "-Command", "$PROFILE").Output()
		if err == nil {
			if path := strings.TrimSpace(string(out)); path != "" {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("couldn't resolve $PROFILE (neither pwsh nor powershell found on PATH)")
}

// appendMarkedBlock appends line to path wrapped in an identifying marker,
// creating the file (and its parent directory) if needed. It's idempotent:
// if the marker is already present, it does nothing.
func appendMarkedBlock(path, line string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	if strings.Contains(string(existing), completionMarkerBegin) {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	block := fmt.Sprintf("\n%s\n%s\n%s\n", completionMarkerBegin, line, completionMarkerEnd)
	if _, err := f.WriteString(block); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// removeMarkedBlock strips the marked block (as written by
// appendMarkedBlock, including the blank line before it) from path, if
// present. Reports removed=false, no error, for a missing file or a file
// with no marker — both are "nothing to do", not failures.
func removeMarkedBlock(path string) (removed bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)
	start := strings.Index(content, completionMarkerBegin)
	if start == -1 {
		return false, nil
	}
	end := strings.Index(content, completionMarkerEnd)
	if end == -1 {
		return false, nil
	}
	end += len(completionMarkerEnd)
	if end < len(content) && content[end] == '\n' {
		end++
	}
	if start > 0 && content[start-1] == '\n' {
		start--
	}

	if err := os.WriteFile(path, []byte(content[:start]+content[end:]), 0o644); err != nil {
		return false, fmt.Errorf("writing %s: %w", path, err)
	}
	return true, nil
}
