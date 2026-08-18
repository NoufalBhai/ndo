# Recipe file format

`central.toml` and `.ndo.toml` share the same schema. `config.toml` is a
separate file for app settings — it never contains recipes.

## `central.toml` / `.ndo.toml`

```toml
[recipes.open]
command = "code {{file}}"
params = ["file"]

[recipes.deploy]
command = "./scripts/deploy.sh {{env}}"
params = ["env"]
description = "Deploy to the given environment"

[recipes.test]
command = "go test ./..."
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `command` | string | yes | Shell command. `{{param}}` tokens are substituted with bound argument values. |
| `params` | array of string | no (default `[]`) | Declared parameter names, in the order positional arguments bind to them. |
| `description` | string | no | Shown in `ndo list`. |
| `depends` | array of string | no | **Reserved for a future release.** Parsed so files round-trip cleanly, but `ndo` does not act on it yet. |

## Positional argument binding

```
ndo open ./file.txt
# open has params = ["file"]
# -> runs: code ./file.txt
```

Args bind to `params` in declared order. If fewer args are given than
`params` declares, `ndo` fails with a hard error — it never silently
renders `{{param}}` as empty (a recipe like `rm {{path}}` running as bare
`rm` would be a nasty footgun).

Extra args beyond the declared params are appended raw (quoted) to the end
of the command line:

```
ndo test -- -run TestFoo
# test has params = []
# -> runs: go test ./... -run TestFoo
```

## Interpolation and quoting

By default, `{{param}}` values are shell-quoted automatically — a value
containing spaces or shell metacharacters is inserted as a single safe
token, not word-split or executed. If you deliberately need an unquoted,
raw substitution (e.g. because the value itself is a snippet of shell
syntax), use the `|raw` escape hatch:

```toml
[recipes.grep-todo]
command = "grep -rn {{pattern|raw}} ."
params = ["pattern"]
```

Referencing a `{{token}}` in `command` that isn't in `params` is an error —
recipes fail loudly on typos rather than silently rendering an empty
string.

## Vars: named lookup tables for param shortcuts

`vars` groups let a positional arg expand into a longer value — e.g. a
short folder alias expanding to a full path — without any special syntax
in the recipe itself:

```toml
[recipes.o]
command = "code {{folder}}"
params = ["folder"]

[vars.folder]
work = "D:\\learn\\work"
ndo  = "D:\\learn\\work\\ndo"
```

```
ndo o work
# param "folder" checks the vars group named "folder" for key "work"
# -> runs: code D:\learn\work

ndo o D:\some\other\path
# "D:\some\other\path" isn't a key in vars.folder, so it's used as-is
# -> runs: code D:\some\other\path
```

The lookup is keyed by **param name**, not recipe name — any recipe with a
param called `folder` benefits from the same `vars.folder` table. An arg
that doesn't match a key in the matching group is left untouched, so this
is purely additive: recipes without a matching vars group behave exactly
as if vars didn't exist.

Manage entries with:

```bash
ndo var add folder work "D:\learn\work" [--local]
ndo var remove folder work [--local]   # remove one entry
ndo var remove folder [--local]        # remove the whole group
ndo var list [group] [--local] [--central]
```

Central and local vars merge **at the key level** (unlike recipes, which
replace wholesale on collision) — central can hold your common shortcuts,
and a local `.ndo.toml` can add or override individual keys without
losing the rest of the central group.

## `config.toml` (settings)

```toml
[settings]
shell = "sh -c"      # default: "sh -c" on Unix, "cmd /C" on Windows
editor = "code -w"   # used by `ndo edit`; falls back to $EDITOR if unset
color = true
```

This file holds app **behavior**, not recipes — kept separate so
recipe-parsing code never has to deal with non-recipe keys.

## Environment variable overrides

| Variable | Effect |
|---|---|
| `NDO_HOME` | Overrides `~/.ndo/` as the directory for `central.toml` and `config.toml`. No XDG fallback — if unset, it's always `~/.ndo/`. |
| `NDO_LOCAL_FILE` | Overrides the local recipe filename searched for upward from the current directory (default `.ndo.toml`). |
