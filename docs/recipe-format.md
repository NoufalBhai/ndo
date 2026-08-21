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
| `depends` | array of string | no | Recipe names to run first, in order. See [Dependencies](#dependencies) below. |

A recipe's name can't match one of ndo's own built-in commands (`add`,
`list`, `edit`, `remove`, `init`, `var`, `completion`, `update`) — `ndo
add` rejects it, since `ndo <name>` always dispatches to the built-in
command of that name first, so the recipe could never actually be run.

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

## Dependencies

`depends` names other recipes to run first. Set it via `ndo add --depends
<name>` (repeatable, in order):

```bash
ndo add build "go build ./..." --local
ndo add lint "go vet ./..." --local
ndo add deploy "./scripts/deploy.sh {{env}}" --param env --depends build --depends lint --local
```

— or by hand-editing the recipe file directly (`ndo edit`):

```toml
[recipes.deploy]
command = "./scripts/deploy.sh {{env}}"
params = ["env"]
depends = ["build", "lint"]

[recipes.build]
command = "go build ./..."

[recipes.lint]
command = "go vet ./..."
```

```
ndo deploy prod
# runs, in order: go build ./...  ->  go vet ./...  ->  ./scripts/deploy.sh prod
```

Rules:

- **Order and transitivity.** Dependencies run before the recipe's own
  command, in the order listed. If a dependency itself has `depends`,
  that resolves recursively (depth-first) — the full graph, not just one
  level.
- **Runs once.** Each recipe runs at most once per invocation, even if
  more than one thing in the chain depends on it (e.g. both `build` and
  `test` depending on `lint`). The recipe you actually invoked always
  runs, even if something upstream already ran it as a dependency.
- **Cycles are a hard error.** `a` depending on `b` depending on `a` is
  rejected before anything runs, not left to recurse.
- **Dependencies never receive arguments.** Only the recipe you invoke
  directly gets the CLI args you passed. A recipe reachable purely as a
  dependency can't declare required `params` — that's a hard
  configuration error, not a runtime surprise. It's still fine for a
  recipe with `params` to have its own `depends`, as long as none of
  *its* dependencies require params.
- **Fail-fast.** If any recipe in the chain exits non-zero, everything
  after it — including the recipe you invoked — is skipped, and that
  exit code becomes `ndo`'s own.
- **`--dry-run` and `--verbose` cover the whole chain**, not just the
  final recipe: `--dry-run` prints every resolved command in run order
  instead of executing any of them, and `--verbose` prints a `# name
  (from source)` line for each recipe as it runs.

Dependencies always run sequentially — there's no `--parallel` yet.

## Vars: named lookup tables for param shortcuts

`vars` groups let a positional arg expand into a longer value — e.g. a
short folder alias expanding to a full path — without any special syntax
in the recipe itself:

```toml
[recipes.o]
command = "code {{folder}}"
params = ["folder"]

[vars.folder]
work = "C:\\Users\\dev\\projects"
ndo  = "C:\\Users\\dev\\projects\\ndo"
```

```
ndo o work
# param "folder" checks the vars group named "folder" for key "work"
# -> runs: code C:\Users\dev\projects

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
ndo var add folder work "C:\Users\dev\projects" [--local]
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
