package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/green-threads/ndo/internal/update"
)

func newUpdateCmd(deps Deps) *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for a newer ndo release and update if possible",
		Long: `Check for a newer ndo release.

If ndo was installed via a package manager (Homebrew, Scoop, apt/dnf) or
"go install", this prints the right command to update it there instead of
touching the binary directly — updating any other way would leave that
package manager's own records out of sync. Only a plain binary install
(e.g. via install.sh, or a manually downloaded release) is updated
in place.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runUpdate(c.Root().Version, c.OutOrStdout(), checkOnly)
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "only check for a newer version, don't install it")
	return cmd
}

func runUpdate(currentVersion string, out io.Writer, checkOnly bool) error {
	if currentVersion == "" || currentVersion == "dev" {
		fmt.Fprintln(out, "Running a dev build (no embedded version) — skipping the update check.")
		return nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	latestTag, err := update.LatestRelease(client)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestTag, "v")
	switch update.CompareVersions(current, latest) {
	case 0:
		fmt.Fprintf(out, "Already on the latest version (%s).\n", currentVersion)
		return nil
	case 1:
		fmt.Fprintf(out, "Running %s, which is ahead of the latest published release (%s) — nothing to do.\n", currentVersion, latestTag)
		return nil
	}
	fmt.Fprintf(out, "A newer version is available: %s -> %s\n", currentVersion, latestTag)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving the running executable's path: %w", err)
	}
	ch := update.DetectChannel(execPath, runtime.GOOS, map[string]string{
		"GOBIN":  os.Getenv("GOBIN"),
		"GOPATH": os.Getenv("GOPATH"),
	})

	instruction, selfReplace := update.InstructionFor(ch)
	if !selfReplace {
		fmt.Fprintf(out, "Installed via a package manager — update with: %s\n", instruction)
		return nil
	}
	if checkOnly {
		fmt.Fprintln(out, "Run `ndo update` (without --check) to install it.")
		return nil
	}

	fmt.Fprintln(out, "Downloading and installing...")
	if err := update.SelfReplace(client, execPath, latestTag, runtime.GOOS, runtime.GOARCH); err != nil {
		return fmt.Errorf("updating: %w", err)
	}
	fmt.Fprintf(out, "Updated to %s.\n", latestTag)
	return nil
}
