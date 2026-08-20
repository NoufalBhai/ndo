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
  - [`ndo update`](#ndo-update)
  - [Shell completion](#shell-completion)
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
  `C:\Users\dev\projects` — without any special syntax in the recipe
  itself.
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
curl -fsSL https://raw.githubusercontent.com/green-threads/ndo/main/install/install.sh | sh
```

Downloads the right binary for your OS/arch and installs it to
`/usr/local/bin` (or `~/.local/bin` if that isn't writable).

**Windows — one-line install script:**

```powershell
irm https://raw.githubusercontent.com/green-threads/ndo/main/install/install.ps1 | iex
```

Or, from `cmd.exe`, download
[`install.cmd`](install/install.cmd) and double-click it (or run it
directly) — it's a thin wrapper around the same PowerShell script for
people who'd rather not type a PowerShell one-liner. Both install to
`%LOCALAPPDATA%\Programs\ndo` by default (override with
`$env:NDO_INSTALL_DIR` / `NDO_INSTALL_DIR`).

**Go toolchain:**

```bash
go install github.com/green-threads/ndo/cmd/ndo@latest
```

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
| **Dependencies** | `depends = [...]` on a recipe — other recipes that run first, in order, deduped across the whole chain. See [`docs/recipe-format.md`](docs/recipe-format.md#dependencies). |
| **Settings** | `~/.ndo/config.toml` — app behavior (shell, editor, color), never recipes. |

## Command reference

### `ndo <name> [args...]`

The default invocation. Resolves `name` against the merged central+local
recipe set, binds `args` positionally to the recipe's declared `params`
(checking [`ndo var`](#ndo-var) lookup tables along the way), and executes
the result. If the recipe has `depends`, those run first, in order, each
recipe running at most once even across a shared dependency — see
[`docs/recipe-format.md`](docs/recipe-format.md#dependencies) for the full
rules (cycles, fail-fast, `--dry-run`/`--verbose` behavior).

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
ndo var add folder work "C:\Users\dev\projects" --local
ndo var add folder ndo "C:\Users\dev\projects\ndo" --local

ndo o work            # -> code C:\Users\dev\projects
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
    ndo: C:\Users\dev\projects\ndo
    work: C:\Users\dev\projects
```

Central and local vars merge **at the key level** — unlike recipes, which
replace wholesale on a name collision. A local `.ndo.toml` can add or
override individual keys without losing the rest of the central group.

### `ndo update`

```
ndo update [--check]
```

Checks the latest GitHub release against the running version. What
happens next depends on how `ndo` was installed, detected from the
running binary's path:

| Detected as | Behavior |
|---|---|
| Homebrew, Scoop, a `.deb`/`.rpm` install, or `go install` | Prints the right command for that channel (`brew upgrade ndo`, etc.) — never touches the binary, so that package manager's own records stay accurate. |
| Anything else (raw binary via `install.sh` or a manual download) | Downloads the new release, verifies its checksum against `checksums.txt`, and atomically replaces the running binary in place. |

`--check` only reports whether an update is available (and the right
command to get it) without installing anything, regardless of which
channel is detected.

### Shell completion

`ndo <TAB>` completes recipe names; `ndo <recipe> <TAB>` completes that
recipe's next declared param — offering the matching `vars` group's keys
(if any) alongside normal file completion, so you never have to memorize
what you named an entry.

The first time you run `ndo` in a real terminal, it asks once whether to
set this up for you — if you say yes, it wires the completion script into
your shell's startup file itself, so it just works in every new terminal
from then on. It never asks again (the answer is remembered in
`config.toml`), and it's a strict no-op in scripts/CI/pipes. To set it up
by hand instead:

```bash
# bash (current shell)
source <(ndo completion bash)

# zsh (current shell)
source <(ndo completion zsh)

# fish
ndo completion fish | source
```

```powershell
# PowerShell — add to your profile ($PROFILE) to persist across sessions
ndo completion powershell | Out-String | Invoke-Expression
```

Or skip the one-time prompt and trigger the same auto-detect-and-install
logic directly:

```bash
ndo completion install     # detect your shell, wire it in
ndo completion uninstall   # remove what install added
```

Run `ndo completion --help` for instructions on installing the script
permanently for your shell.

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
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — dev setup, coding conventions, and
  the PR process.

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
