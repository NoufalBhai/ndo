// Command ndo is a CLI-first task runner with centrally+locally layered
// recipes. See DESIGN.md for the full rationale.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/green-threads/ndo/internal/cli"
	"github.com/green-threads/ndo/internal/config"
)

// version is overridden at build time via:
//
//	go build -ldflags "-X main.version=v1.2.3"
var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	ndoHome, err := config.ResolveNDOHome()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ndo:", err)
		return 1
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ndo:", err)
		return 1
	}

	deps := cli.Deps{
		NDOHome: ndoHome,
		Cwd:     cwd,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}

	root := cli.NewRootCmd(version, deps)
	if err := root.Execute(); err != nil {
		var exitErr *cli.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.Code
		}
		fmt.Fprintln(os.Stderr, "ndo:", err)
		return 1
	}
	return 0
}
