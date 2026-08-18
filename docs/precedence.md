# Recipe precedence: central vs. local

`ndo` reads recipes from two places:

- **Central**: `~/.ndo/central.toml` (or `$NDO_HOME/central.toml` if
  `NDO_HOME` is set) — recipes you want available everywhere.
- **Local**: `.ndo.toml` in the current directory, or the nearest one found
  by searching upward through parent directories (the same way `git` finds
  `.git`) — recipes specific to the project you're in.

## When a recipe name exists in both

**The local recipe silently wins.** No warning is printed. This is
deliberate, not an oversight.

```toml
# ~/.ndo/central.toml
[recipes.deploy]
command = "./scripts/deploy.sh production"
```

```toml
# ./.ndo.toml
[recipes.deploy]
command = "./scripts/deploy.sh staging"
```

Running `ndo deploy` here runs the **staging** deploy — the local file's
version — with no message telling you the central one was shadowed.

## Why no warning?

This matches a pattern you already rely on elsewhere: a local `.gitconfig`,
`.npmrc`, or shell RC file overriding a global one doesn't print a warning
every time you use it — it's just how layered config works, and you learn
it once. Printing "note: local `deploy` overrides central `deploy`" on
every single invocation of a recipe you override on purpose would be pure
noise for the common case.

## The whole recipe is replaced, not merged field-by-field

If a local recipe shares a name with a central one, the local entry
replaces it **entirely** — command, params, description, everything. There
is no way to, say, keep the central `command` but override just the
`description`. If you need a different behavior, give the recipe its own
full definition locally.

## If you're not sure which one ran

Pass `--verbose` to see the source file each resolved recipe came from:

```
$ ndo --verbose deploy
# deploy (from /path/to/project/.ndo.toml)
```

Or use `--dry-run` to see the fully interpolated command without running
it:

```
$ ndo --dry-run deploy
./scripts/deploy.sh staging
```

## Multiple local files up the tree

Only the **nearest** `.ndo.toml` is used. If both
`~/projects/.ndo.toml` and `~/projects/app/.ndo.toml` exist, running `ndo`
from inside `~/projects/app/` only reads the nearer one — they are not
merged together.
