package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/NoufalBhai/ndo/internal/config"
)

// newVarCmd builds `ndo var`, a named lookup-table group used to expand
// short values (e.g. a folder alias) to full ones during recipe param
// binding. A recipe param named "folder" checks the vars group named
// "folder" for a matching key before falling back to the literal arg —
// see runRecipe in run.go.
func newVarCmd(deps Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "var",
		Short: "Manage named lookup tables used to expand recipe param values",
	}
	cmd.AddCommand(newVarAddCmd(deps), newVarRemoveCmd(deps), newVarListCmd(deps))
	return cmd
}

func newVarAddCmd(deps Deps) *cobra.Command {
	var local bool
	cmd := &cobra.Command{
		Use:   "add <group> <key> <value>",
		Short: "Add or overwrite an entry in a vars group",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVarAdd(deps, cmd.OutOrStdout(), args[0], args[1], args[2], local)
		},
	}
	cmd.Flags().BoolVar(&local, "local", false, "write to the nearest local .ndo.toml instead of the central file")
	return cmd
}

func runVarAdd(deps Deps, stdout io.Writer, group, key, value string, local bool) error {
	targetPath, err := targetFilePath(deps, local)
	if err != nil {
		return err
	}

	rf, err := config.LoadRecipeFile(targetPath)
	if err != nil {
		return err
	}
	if rf.Vars[group] == nil {
		rf.Vars[group] = map[string]string{}
	}
	rf.Vars[group][key] = value

	if err := config.SaveRecipeFile(targetPath, rf); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "set %s.%s = %q in %s\n", group, key, value, targetPath)
	return nil
}

func newVarRemoveCmd(deps Deps) *cobra.Command {
	var local bool
	cmd := &cobra.Command{
		Use:   "remove <group> [key]",
		Short: "Remove one entry from a vars group, or the whole group if key is omitted",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := ""
			if len(args) == 2 {
				key = args[1]
			}
			return runVarRemove(deps, cmd.OutOrStdout(), args[0], key, local)
		},
	}
	cmd.Flags().BoolVar(&local, "local", false, "remove from the nearest local .ndo.toml instead of the central file")
	return cmd
}

// runVarRemove removes a single key from group, or the entire group when
// key is "".
func runVarRemove(deps Deps, stdout io.Writer, group, key string, local bool) error {
	var targetPath string
	if local {
		path, ok := config.SearchUpward(deps.Cwd, config.LocalFileName())
		if !ok {
			return fmt.Errorf("no local recipe file found (searched upward from %s)", deps.Cwd)
		}
		targetPath = path
	} else {
		targetPath = centralFilePath(deps.NDOHome)
	}

	rf, err := config.LoadRecipeFile(targetPath)
	if err != nil {
		return err
	}

	if key == "" {
		if _, exists := rf.Vars[group]; !exists {
			return fmt.Errorf("vars group %q not found in %s", group, targetPath)
		}
		delete(rf.Vars, group)
		if err := config.SaveRecipeFile(targetPath, rf); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "removed group %q from %s\n", group, targetPath)
		return nil
	}

	if _, exists := rf.Vars[group][key]; !exists {
		return fmt.Errorf("var %s.%s not found in %s", group, key, targetPath)
	}

	delete(rf.Vars[group], key)
	if len(rf.Vars[group]) == 0 {
		delete(rf.Vars, group)
	}
	if err := config.SaveRecipeFile(targetPath, rf); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "removed %s.%s from %s\n", group, key, targetPath)
	return nil
}

func newVarListCmd(deps Deps) *cobra.Command {
	var local, central bool
	cmd := &cobra.Command{
		Use:   "list [group]",
		Short: "List resolved vars entries, optionally filtered to one group",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			group := ""
			if len(args) == 1 {
				group = args[0]
			}
			return runVarList(deps, cmd.OutOrStdout(), group, local, central)
		},
	}
	cmd.Flags().BoolVar(&local, "local", false, "show only the local file's vars, unmerged")
	cmd.Flags().BoolVar(&central, "central", false, "show only the central file's vars, unmerged")
	return cmd
}

func runVarList(deps Deps, stdout io.Writer, group string, local, central bool) error {
	res, err := loadResolved(deps)
	if err != nil {
		return err
	}

	rf := res.Merged
	switch {
	case local && central:
		return fmt.Errorf("--local and --central are mutually exclusive")
	case local:
		rf = res.Local
	case central:
		rf = res.Central
	}

	groups := make([]string, 0, len(rf.Vars))
	for g := range rf.Vars {
		if group != "" && g != group {
			continue
		}
		groups = append(groups, g)
	}
	sort.Strings(groups)

	for _, g := range groups {
		keys := make([]string, 0, len(rf.Vars[g]))
		for k := range rf.Vars[g] {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Fprintf(stdout, "%s:\n", g)
		for _, k := range keys {
			fmt.Fprintf(stdout, "    %s: %s\n", k, rf.Vars[g][k])
		}
	}
	return nil
}
