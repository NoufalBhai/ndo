# ndo

[![CI](https://github.com/green-threads/ndo/actions/workflows/ci.yml/badge.svg)](https://github.com/green-threads/ndo/actions/workflows/ci.yml)
[![Go Reference](https://img.shields.io/badge/go-1.24-blue)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-informational)](LICENSE)

**ndo** is a CLI-first command runner: define short, named shell commands
("recipes") and run them by name, with arguments, from anywhere. It ships
as a single static binary — no runtime, no interpreter, no config beyond
plain TOML files.

## Table of contents

- [Why ndo](#why-ndo)
- [Install](#install)
- [Quickstart](#quickstart)
- [Concepts](#concepts)
- [Command reference](#command-reference)
  - [`ndo <name> [args...]`](#ndo-name-args)
  - [`ndo add`](#ndo-add)
  - [`ndo list`](#ndo-list)
  - [`ndo edit`](#ndo-edit)
  - [`ndo remove`](#ndo-remove)
  - [`ndo init`](#ndo-init)
  - [`ndo var`](#ndo-var)
  - [Global flags](#global-flags)
- [Configuration](#configuration)
- [Docs](#docs)
- [Development](#development)
- [Testing](#testing)
- [Versioning](#versioning)
- [Design](#design)
- [License](#license)

## Why ndo

- **Add recipes from the CLI.** `ndo add <name> "<command>" --param x` —
  no need to hand-edit a config file just to add a one-liner.
- **Central + local recipes, layered.** A global recipe set
  (`~/.ndo/central.toml`) holds shortcuts you want on every machine and
  every project; a per-project `.ndo.toml` holds project-specific ones.
  Same name in both → the local one silently wins, no runtime noise. See
  [`docs/precedence.md`](docs/precedence.md) for the full rationale.
- **Natural positional arguments.** `ndo open ./file.txt` binds
  `./file.txt` straight to the recipe's declared parameter — no flag
  ceremony.
- **Named lookup tables for arguments (`ndo var`).** Expand a short alias
  into a full value automatically — e.g. `ndo o work` → opens
  `D:\learn\work` — without any special syntax in the recipe itself.
- **Single static binary, cross-platform.** Linux, macOS, and Windows,
  amd64 and arm64. No dependency on any other task runner.

## Install

**macOS / Linux — Homebrew:**

```bash
brew install green-threads/ndo/ndo
```

**Windows — [Scoop](https://scoop.sh):**

```powershell
scoop bucket add ndo https://github.com/green-threads/scoop-ndo
scoop install ndo
```

**Linux — .deb / .rpm:**

Download the `.deb` or `.rpm` for your architecture from the
[Releases page](../../releases) and install with your package manager:

```bash
sudo apt install ./ndo_<version>_<arch>.deb     # Debian/Ubuntu
sudo rpm -i ndo_<version>_<arch>.rpm            # Fedora/RHEL
```

**macOS / Linux — one-line install script:**

```bash
curl -fsSL https://raw.githubusercontent.com/green-threads/ndo/main/install.sh | sh
```

Downloads the right binary for your OS/arch and installs it to
`/usr/local/bin` (or `~/.local/bin` if that isn't writable).

**Any platform — raw binary:**

Download a prebuilt archive for your platform from the
[Releases page](../../releases), or build from source (requires Go 1.24+):

```bash
go build -o ndo ./cmd/ndo
```

## Quickstart

```bash
# create a project-local recipe file (.ndo.toml) in the current directory
ndo init

# add a recipe — writes to .ndo.toml because --local is passed
ndo add open "code {{file}}" --param file --local

# run it — the positional arg binds to {{file}}
ndo open ./main.go

# see everything that resolves in this directory (central + local, merged)
ndo list
```

Recipes added *without* `--local` go into the central file
(`~/.ndo/central.toml`) and are available in every project on the machine.

## Concepts

| Term | Meaning |
|---|---|
| **Recipe** | A named shell command with optional declared parameters. |
| **Central file** | `~/.ndo/central.toml` (or `$NDO_HOME/central.toml`) — recipes available everywhere. |
| **Local file** | `.ndo.toml`, found by searching upward from the current directory (git-style). Project-specific recipes. |
| **Vars** | Named lookup tables (`[vars.<group>]`) that expand a positional argument into a longer value, keyed by param name. |
| **Settings** | `~/.ndo/config.toml` — app behavior (shell, editor, color), never recipes. |

## Command reference

### `ndo <name> [args...]`

The default invocation. Resolves `name` against the merged central+local
recipe set, binds `args` positionally to the recipe's declared `params`
(checking [`ndo var`](#ndo-var) lookup tables along the way), and executes
the result.

```bash
ndo open ./file.txt
```

Args beyond the declared params are appended raw (quoted) to the end of
the command — useful for passthrough:

```bash
ndo test -- -run TestFoo
```

### `ndo add`

```
ndo add <name> "<command>" [--param <name>]... [--local] [--desc "<text>"]
```

Adds a new recipe. Fails with an error (not a silent overwrite) if the
name already exists in the **target** file — use `ndo edit` to change an
existing recipe by hand.

```bash
ndo add deploy "./scripts/deploy.sh {{env}}" --param env --local --desc "Deploy to the given environment"
```

| Flag | Effect |
|---|---|
| `--param <name>` | Declares a parameter, in positional order. Repeatable. |
| `--local` | Write to the nearest local `.ndo.toml` (creating one in the current directory if none is found upward) instead of the central file. |
| `--desc "<text>"` | Description shown in `ndo list`. |

### `ndo list`

```
ndo list [--local] [--central]
```

Prints resolved recipes. With no flags, shows the merged central+local
view. `--local`/`--central` restrict to just that file's contents,
unmerged.

### `ndo edit`

```
ndo edit [--local] [--central]
```

Opens the relevant recipe file in `$EDITOR` (or `config.toml`'s
`settings.editor`, which takes precedence). Defaults to the central file;
`--local` opens the nearest local file instead.

### `ndo remove`

```
ndo remove <name> [--local]
```

Deletes a recipe from the target file. Errors if not found in that
specific file — it doesn't cascade-search the other one.

### `ndo init`

```
ndo init
```

Creates an empty `.ndo.toml` in the current directory. Errors if a local
recipe file already exists anywhere in the tree above the current
directory, so it never silently shadows one.

### `ndo var`

Named lookup tables that expand a recipe's positional argument into a
longer value, without any new recipe syntax — the lookup is keyed by
**param name**. See [`docs/recipe-format.md`](docs/recipe-format.md#vars-named-lookup-tables-for-param-shortcuts)
for the full mechanics and TOML schema.

```bash
ndo add o "code {{folder}}" --param folder --local
ndo var add folder work "D:\learn\work" --local
ndo var add folder ndo "D:\learn\work\ndo" --local

ndo o work            # -> code D:\learn\work
ndo o D:\other\path   # no match in vars.folder -> code D:\other\path (used literally)
```

| Subcommand | Effect |
|---|---|
| `ndo var add <group> <key> <value> [--local]` | Add or overwrite one entry. |
| `ndo var remove <group> <key> [--local]` | Remove one entry. |
| `ndo var remove <group> [--local]` | Remove the **entire group** at once. |
| `ndo var list [group] [--local] [--central]` | List resolved entries, optionally filtered to one group. |

`ndo var list` output:

```
folder:
    ndo: D:\learn\work\ndo
    work: D:\learn\work
```

Central and local vars merge **at the key level** — unlike recipes, which
replace wholesale on a name collision. A local `.ndo.toml` can add or
override individual keys without losing the rest of the central group.

### Global flags

| Flag | Effect |
|---|---|
| `--verbose` | Print extra diagnostics, including which file a resolved recipe came from. |
| `--dry-run` | Print the resolved, interpolated command without executing it. |
| `--version` | Print the ndo version. |
| `--help` | Show help for any command. |

## Configuration

### `config.toml` (app settings — never recipes)

```toml
[settings]
shell = "sh -c"      # default: "sh -c" on Unix, "cmd /C" on Windows
editor = "code -w"   # used by `ndo edit`; falls back to $EDITOR if unset
color = true
```

### Environment variables

| Variable | Effect |
|---|---|
| `NDO_HOME` | Overrides `~/.ndo/` as the directory for `central.toml` and `config.toml`. No XDG fallback — if unset, it's always `~/.ndo/`. |
| `NDO_LOCAL_FILE` | Overrides the local recipe filename searched for upward from the current directory (default `.ndo.toml`). |

## Docs

- [`docs/precedence.md`](docs/precedence.md) — how central and local
  recipes combine, and why the override is silent by design.
- [`docs/recipe-format.md`](docs/recipe-format.md) — the full TOML schema
  for recipe, vars, and settings files, plus interpolation/quoting rules.
- [`CHANGELOG.md`](CHANGELOG.md) — notable changes, in
  [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.
- [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) — community standards for
  this repository's issues and pull requests.

## Development

```bash
go build -o ndo ./cmd/ndo   # build
go test ./...                # test everything
go vet ./...                  # lint-ish
gofmt -l .                    # check formatting (should output nothing)
```

`internal/` is genuinely internal — only `cmd/ndo/main.go` depends on
`internal/cli`. No dependency on any other task runner binary anywhere in
this codebase.

## Testing

Table-driven unit tests cover the merge logic, param binding/quoting, and
upward local-file search. `internal/cli` tests exercise every command
end-to-end against `t.TempDir()` with injected dependencies (no real
filesystem or `~/.ndo/` touched), including a golden-file test for
`ndo list` output. No network calls anywhere in the suite.

CI runs `gofmt`/`go vet`/`go test -race -cover` plus a cross-compile
matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
windows/amd64) on every push and pull request. Tagged pushes (`v*.*.*`)
trigger a release workflow that builds, archives, checksums, and publishes
binaries for all five targets.

## Versioning

ndo follows [Semantic Versioning](https://semver.org/): given a version
`MAJOR.MINOR.PATCH`, `MAJOR` changes on breaking changes (recipe/vars
schema, CLI flag/command removal or renaming), `MINOR` on
backwards-compatible feature additions, `PATCH` on backwards-compatible
fixes. Release tags follow `vMAJOR.MINOR.PATCH` (e.g. `v1.2.3`), matching
what the [release workflow](.github/workflows/release.yml) builds from.
The project hasn't cut `v1.0.0` yet — see
[`CHANGELOG.md`](CHANGELOG.md) for what's landed so far.

## Design

Architecture, resolution algorithm, and the rationale behind every
non-obvious decision (why TOML, why `~/.ndo/` instead of XDG, why local
silently overrides central, why vars merge at the key level while recipes
don't) are kept in a design document maintained alongside the code but not
published in this repository.

## License

[MIT](LICENSE)
