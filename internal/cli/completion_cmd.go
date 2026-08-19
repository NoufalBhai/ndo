package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/config"
)

// newCompletionCmd builds ndo's own "completion" command tree in place of
// cobra's auto-generated one (disabled via root.CompletionOptions.
// DisableDefaultCmd), so "install"/"uninstall" can sit alongside the
// standard per-shell script generators.
func newCompletionCmd(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
		Long: `Generate the autocompletion script for ndo for the specified shell.

Run one of the per-shell subcommands to print the script (see the README
for how to source it), or use "install"/"uninstall" to have ndo detect
your shell and wire it into your startup file itself.`,
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:                   "bash",
			Short:                 "Generate the autocompletion script for bash",
			Args:                  cobra.NoArgs,
			DisableFlagsInUseLine: true,
			RunE: func(c *cobra.Command, args []string) error {
				return c.Root().GenBashCompletionV2(c.OutOrStdout(), true)
			},
		},
		&cobra.Command{
			Use:                   "zsh",
			Short:                 "Generate the autocompletion script for zsh",
			Args:                  cobra.NoArgs,
			DisableFlagsInUseLine: true,
			RunE: func(c *cobra.Command, args []string) error {
				return c.Root().GenZshCompletion(c.OutOrStdout())
			},
		},
		&cobra.Command{
			Use:                   "fish",
			Short:                 "Generate the autocompletion script for fish",
			Args:                  cobra.NoArgs,
			DisableFlagsInUseLine: true,
			RunE: func(c *cobra.Command, args []string) error {
				return c.Root().GenFishCompletion(c.OutOrStdout(), true)
			},
		},
		&cobra.Command{
			Use:                   "powershell",
			Short:                 "Generate the autocompletion script for PowerShell",
			Args:                  cobra.NoArgs,
			DisableFlagsInUseLine: true,
			RunE: func(c *cobra.Command, args []string) error {
				return c.Root().GenPowerShellCompletionWithDesc(c.OutOrStdout())
			},
		},
		&cobra.Command{
			Use:   "install",
			Short: "Detect your shell and wire ndo completion into its startup file",
			Args:  cobra.NoArgs,
			RunE: func(c *cobra.Command, args []string) error {
				return runCompletionInstall(deps, c.Root(), c.OutOrStdout())
			},
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Remove the completion block ndo previously installed",
			Args:  cobra.NoArgs,
			RunE: func(c *cobra.Command, args []string) error {
				return runCompletionUninstall(deps, c.OutOrStdout())
			},
		},
	)

	return cmd
}

func runCompletionInstall(deps Deps, root *cobra.Command, out io.Writer) error {
	path, shell, err := installCompletion(root)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Shell completion enabled for %s (added to %s). Open a new terminal for it to take effect.\n", shell, path)

	settings, err := config.LoadSettings(deps.NDOHome)
	if err != nil {
		return err
	}
	settings.CompletionPromptAnswered = true
	return config.SaveSettings(deps.NDOHome, settings)
}

func runCompletionUninstall(deps Deps, out io.Writer) error {
	path, shell, removed, err := uninstallCompletion()
	if err != nil {
		return err
	}
	if !removed {
		fmt.Fprintln(out, "No ndo completion block found — nothing to remove.")
		return nil
	}
	fmt.Fprintf(out, "Removed %s completion from %s.\n", shell, path)

	settings, err := config.LoadSettings(deps.NDOHome)
	if err != nil {
		return err
	}
	settings.CompletionPromptAnswered = false
	return config.SaveSettings(deps.NDOHome, settings)
}
