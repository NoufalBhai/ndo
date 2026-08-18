package recipe

import (
	"io"
	"os/exec"
)

// Execute runs command via the given shell argv prefix (e.g. ["sh", "-c"]
// or ["cmd", "/C"]), wiring stdin/stdout/stderr directly, and returns the
// child process's exit code.
func Execute(shell []string, command string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	argv := append(append([]string{}, shell...), command)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), nil
	}
	return -1, err
}
