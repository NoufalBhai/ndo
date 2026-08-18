# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project intends to follow [Semantic Versioning](https://semver.org/)
once the first tagged release ships.

## [Unreleased]

### Added

- Core v1 scaffold: `internal/recipe` (positional argument binding,
  automatic shell-quoting with a `{{param|raw}}` escape hatch, per-OS
  default shell selection via Go build tags), `internal/config`
  (central + local TOML recipe loading with local-wins merge, upward
  `.ndo.toml` search, `NDO_HOME` resolution), and `internal/cli`
  (`add`, `list`, `edit`, `remove`, `init`, plus the bare
  `ndo <name> [args...]` invocation).
- `ndo var` — named lookup tables that expand a recipe's positional
  argument into a longer value (e.g. a short folder alias expanding to a
  full path), without any new recipe syntax. Subcommands:
  - `ndo var add <group> <key> <value> [--local]`
  - `ndo var remove <group> <key> [--local]` — remove one entry
  - `ndo var remove <group> [--local]` — remove an entire group at once
  - `ndo var list [group] [--local] [--central]`
- GitHub Actions CI (`gofmt`/`go vet`/`go test -race -cover`, plus a
  5-target cross-compile matrix: linux/amd64, linux/arm64, darwin/amd64,
  darwin/arm64, windows/amd64) and a tag-triggered release workflow that
  builds, archives, checksums, and publishes binaries for all five
  targets.
- `docs/precedence.md` and `docs/recipe-format.md`.

### Changed

- Vars merge **at the key level** between central and local files —
  distinct from recipes, which replace wholesale on a name collision.
  Central holds shared defaults; a local `.ndo.toml` can add or override
  individual keys without losing the rest of the central group.
- CI/release Actions dependencies kept current via Dependabot
  (`actions/checkout@v7`, `actions/setup-go@v7`,
  `actions/download-artifact@v8`, `actions/upload-artifact@v7`,
  `softprops/action-gh-release@v3`).
