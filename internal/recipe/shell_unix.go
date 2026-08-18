//go:build !windows

package recipe

import "strings"

// DefaultShell returns the argv prefix used to invoke a recipe's command
// when no shell is configured in settings.toml.
func DefaultShell() []string {
	return []string{"sh", "-c"}
}

// QuoteArg POSIX-single-quotes a value for safe interpolation into a
// `sh -c` command line. Single quotes suppress all shell metacharacter
// interpretation except the single quote itself, which is closed,
// escaped, and reopened.
func QuoteArg(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
