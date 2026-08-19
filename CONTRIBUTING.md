# Contributing to ndo

Thanks for considering a contribution. This is a small project with a
deliberately narrow scope — please read this before opening a PR so your
change has the best chance of landing cleanly.

## Before you start

- For anything beyond a small fix (a new flag, a new command, a change to
  the recipe/vars schema), open an issue first to discuss the approach.
  This project has a written design doc with rationale behind most
  non-obvious decisions (config format, directory layout, override
  behavior, etc.) — some things that look like oversights are deliberate,
  and it's better to find that out before you write code.
- Check [`CHANGELOG.md`](CHANGELOG.md) and open issues/PRs to avoid
  duplicating work already in flight.
- `ndo` is a standalone tool, not a wrapper or fork of any other task
  runner. Don't add a dependency on one.

## Development setup

Requires Go 1.24+.

```bash
git clone https://github.com/green-threads/ndo.git
cd ndo
go build -o ndo ./cmd/ndo
go test ./...
```

## Making a change

- **Format**: `gofmt -l .` must print nothing before you commit.
- **Vet**: `go vet ./...` must be clean.
- **Tests**: `go test ./...` (CI also runs `-race -cover`). Add
  table-driven tests for new logic, especially around recipe/vars merging,
  parameter binding, and file resolution — these are the areas most prone
  to edge cases.
- Keep `internal/` internal — only `cmd/ndo/main.go` should import from
  `internal/cli`.
- No new external dependencies without discussing them in an issue first.
  No YAML, ever (see `docs/recipe-format.md` for why TOML was chosen).

```bash
go build -o ndo ./cmd/ndo   # build
go test ./...                # test
go vet ./...                  # lint-ish
gofmt -l .                    # formatting check (should output nothing)
```

## Commit messages

Plain, descriptive, present-tense summary line (e.g. `Add ndo var remove
for whole-group deletion`). No fixed prefix convention is enforced, but
keep one logical change per commit.

## Pull requests

- Keep PRs focused — one feature or fix per PR.
- Update `CHANGELOG.md` under `[Unreleased]` for any user-facing change.
- Update `README.md` / `docs/` if you're adding or changing a command or
  the recipe/vars schema.
- CI (`gofmt`, `go vet`, `go test -race -cover`, `goreleaser check`) must
  pass before review.

## Reporting bugs / requesting features

Use [GitHub issues](../../issues). Include your OS, `ndo --version`, and
(for bugs) the minimal recipe/command that reproduces the problem.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).

## License

By contributing, you agree your contributions are licensed under this
project's [MIT license](LICENSE).
