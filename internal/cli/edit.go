package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/config"
)

func newEditCmd(deps Deps) *cobra.Command {
	var local, central bool

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Open a recipe file in $EDITOR",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(deps, local, central)
		},
	}

	cmd.Flags().BoolVar(&local, "local", false, "open the nearest local .ndo.toml instead of the central file")
	cmd.Flags().BoolVar(&central, "central", false, "open the central file explicitly (default target)")

	return cmd
}

func runEdit(deps Deps, local, central bool) error {
	if local && central {
		return fmt.Errorf("--local and --central are mutually exclusive")
	}

	path, err := targetFilePath(deps, local)
	if err != nil {
		return err
	}

	editor, err := resolveEditor(deps.NDOHome)
	if err != nil {
		return err
	}

	fields, err := splitCommand(editor)
	if err != nil {
		return fmt.Errorf("invalid editor command %q: %w", editor, err)
	}
	fields = append(fields, path)

	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// splitCommand tokenizes an editor command string on whitespace, with
// double quotes allowed around a single field (e.g. `"C:\Program
// Files\Editor\editor.exe" --wait`) so an editor path containing spaces
// doesn't get split apart. Returns an error for an empty/whitespace-only
// command or an unterminated quote — never a zero-length result that
// would silently make the target file itself fields[0].
func splitCommand(s string) ([]string, error) {
	var fields []string
	var cur strings.Builder
	inQuotes := false
	hasCur := false

	for _, r := range s {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			hasCur = true
		case unicode.IsSpace(r) && !inQuotes:
			if hasCur {
				fields = append(fields, cur.String())
				cur.Reset()
				hasCur = false
			}
		default:
			cur.WriteRune(r)
			hasCur = true
		}
	}
	if inQuotes {
		return nil, fmt.Errorf("unterminated %q", `"`)
	}
	if hasCur {
		fields = append(fields, cur.String())
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return fields, nil
}

func resolveEditor(ndoHome string) (string, error) {
	settings, err := config.LoadSettings(ndoHome)
	if err != nil {
		return "", err
	}
	if settings.Editor != "" {
		return settings.Editor, nil
	}
	if env := os.Getenv("EDITOR"); env != "" {
		return env, nil
	}
	return "", fmt.Errorf("no editor configured: set $EDITOR or config.toml's [settings] editor")
}
