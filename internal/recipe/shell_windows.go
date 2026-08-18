//go:build windows

package recipe

import "strings"

// DefaultShell returns the argv prefix used to invoke a recipe's command
// when no shell is configured in settings.toml.
func DefaultShell() []string {
	return []string{"cmd", "/C"}
}

// QuoteArg double-quotes a value for interpolation into a `cmd /C` command
// line. cmd.exe has no fully safe quoting mechanism for every
// metacharacter (&, |, %, ^ remain best-effort); this covers the common
// case of values containing spaces.
func QuoteArg(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, " \t\"&|<>^%") {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
