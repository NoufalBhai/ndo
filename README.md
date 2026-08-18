# ndo

A CLI-first task runner. Like [`just`](https://github.com/casey/just), but:

- **Add recipes from the CLI** — `ndo add <name> "<command>"`, no editor
  required.
- **Central + local recipes** — a global recipe set (`~/.ndo/central.toml`)
  layered under a per-project one (`.ndo.toml`); the local one silently
  wins on a name collision. See [`docs/precedence.md`](docs/precedence.md).
- **Natural positional arguments** — `ndo open ./file.txt` binds
  `./file.txt` to the recipe's first declared parameter automatically.

Single static binary, cross-platform (Linux, macOS, Windows).

## Install

Download a release binary from the
[Releases page](../../releases) for your platform, or build from source:

```bash
go build -o ndo ./cmd/ndo
```

## Quickstart

```bash
# create a project-local recipe file
ndo init

# add a recipe (writes to .ndo.toml since --local is passed)
ndo add open "code {{file}}" --param file --local

# run it — the positional arg binds to {file}
ndo open ./main.go

# see everything that resolves in this directory
ndo list
```

Recipes added without `--local` go into the central file
(`~/.ndo/central.toml`), available in every project on the machine.

## Docs

- [`docs/precedence.md`](docs/precedence.md) — how central and local
  recipes combine.
- [`docs/recipe-format.md`](docs/recipe-format.md) — the TOML schema for
  recipe and settings files.

## Development

```bash
go build -o ndo ./cmd/ndo   # build
go test ./...                # test everything
go vet ./...                  # lint-ish
gofmt -l .                    # check formatting (should output nothing)
```
