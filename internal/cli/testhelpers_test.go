package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

type rootHarness struct {
	cmd    *cobra.Command
	deps   Deps
	out    *bytes.Buffer
	errOut *bytes.Buffer
}

func (h *rootHarness) Execute() error { return h.cmd.Execute() }

// newTestRoot builds a root command wired to temp NDO_HOME/cwd dirs, with
// stdout/stderr captured in buffers instead of touching the real terminal
// or the machine's real ~/.ndo.
func newTestRoot(t *testing.T, args ...string) *rootHarness {
	t.Helper()
	ndoHome := t.TempDir()
	cwd := t.TempDir()

	deps := Deps{
		NDOHome: ndoHome,
		Cwd:     cwd,
		Stdin:   bytes.NewReader(nil),
	}

	root := NewRootCmd("test", deps)
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)

	return &rootHarness{cmd: root, deps: deps, out: &out, errOut: &errOut}
}
