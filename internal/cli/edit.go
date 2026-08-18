package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	fields := strings.Fields(editor)
	fields = append(fields, path)

	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
