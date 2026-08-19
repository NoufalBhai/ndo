# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.3.1] - 2026-08-20

### Added

- `ndo completion install` / `ndo completion uninstall` — force a
  (re-)install or cleanly remove the completion block without going
  through the one-time prompt.

### Fixed

- Windows shell detection checked `PSModulePath` before `MSYSTEM`, so Git
  Bash launched from a PowerShell terminal (which inherits
  `PSModulePath`) could be misdetected as PowerShell and get the
  completion block written to `$PROFILE` instead of `.bashrc`. `MSYSTEM`
  (and `SHELL` containing `bash`) now take priority.
- The completion-setup prompt could fire from inside `ndo completion
  bash`/`zsh`/etc. themselves — since those run as a real subprocess every
  time a shell sources them, this could re-trigger the prompt on shell
  startup for anyone who'd set up completion manually rather than through
  the prompt. Fixed by skipping the prompt for the whole `completion`
  command subtree, not just an exact `cmd.Name() == "completion"` match.

## [0.3.0] - 2026-08-20

### Added

- One-time shell completion setup: the first time `ndo` runs in a real
  interactive terminal, it asks once whether to enable tab-completion and,
  if yes, wires the completion script into your shell's startup file
  itself (bash/zsh: appends a `source <(ndo completion ...)` line; fish:
  writes to `~/.config/fish/completions/`; PowerShell: appends to
  `$PROFILE`, resolved by asking `pwsh`/`powershell` directly rather than
  guessing between the 5.1 and 7+ default paths). The answer is recorded
  in `config.toml` (`completion_prompt_answered`) so it's never asked
  again. Strictly a no-op for non-interactive stdin (scripts, CI,
  pipes) — gated on an actual isatty check via `golang.org/x/term`, not a
  file-mode heuristic.

## [0.2.0] - 2026-08-20

### Added

- Shell tab-completion (`ndo completion bash|zsh|fish|powershell`, via
  cobra's built-in completion command). `ndo <TAB>` completes recipe
  names; `ndo <recipe> <TAB>` completes that recipe's next declared
  param, offering the matching `vars` group's keys (if any) alongside
  normal file completion.
- `CONTRIBUTING.md`, `SECURITY.md`, and GitHub issue/PR templates.

### Changed

- CI now runs `go test` on Linux, macOS, and Windows (previously
  Linux-only for the test job; other OSes were cross-compiled but never
  actually test-executed).

### Fixed

- A `.gitattributes` was missing, so a golden test fixture got checked
  out with CRLF line endings on Windows CI and failed a byte-for-byte
  comparison against program output (which always emits `\n`). Fixed by
  marking golden fixtures `-text` and normalizing everything else to
  `eol=lf`.

## [0.1.1] - 2026-08-18

### Added

- Installable packages via goreleaser: a Homebrew tap
  (`green-threads/homebrew-ndo`), a Scoop bucket (`green-threads/scoop-ndo`),
  `.deb`/`.rpm` packages, and a one-line install script (`install.sh`).
- `go install github.com/green-threads/ndo/cmd/ndo@latest` as an
  additional install path.

### Changed

- Release pipeline switched from a hand-rolled GitHub Actions matrix to
  `.goreleaser.yaml` + `goreleaser/goreleaser-action`, which now also
  drives the packages above. CI's cross-compile job was replaced by
  `goreleaser check` + `goreleaser build --snapshot` for equivalent
  per-PR coverage.

## [0.1.0] - 2026-08-18

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
- `CODE_OF_CONDUCT.md` (Contributor Covenant v2.1, reports routed through
  GitHub issues) and a README section documenting the SemVer versioning
  policy.

### Changed

- Vars merge **at the key level** between central and local files —
  distinct from recipes, which replace wholesale on a name collision.
  Central holds shared defaults; a local `.ndo.toml` can add or override
  individual keys without losing the rest of the central group.
- CI/release Actions dependencies kept current via Dependabot
  (`actions/checkout@v7`, `actions/setup-go@v7`,
  `actions/download-artifact@v8`, `actions/upload-artifact@v7`,
  `softprops/action-gh-release@v3`).

[Unreleased]: https://github.com/green-threads/ndo/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/green-threads/ndo/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/green-threads/ndo/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/green-threads/ndo/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/green-threads/ndo/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/green-threads/ndo/releases/tag/v0.1.0
