# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [1.0.0] - 2026-08-22

First stable release. The recipe/vars TOML schema and CLI surface
(`add`, `list`, `edit`, `remove`, `init`, `var`, `update`, `completion`,
and bare `ndo <name> [args...]`) are the compatibility baseline going
forward — see [Versioning](README.md#versioning).

### Added

- `depends` on a recipe now runs — other recipes to run first, in order,
  resolved transitively (a dependency's own `depends` also runs) and
  deduped so a shared dependency runs once per invocation even when
  reached through multiple paths. Dependency cycles and dependencies with
  required params (which could never receive arguments) are hard errors,
  caught before anything runs. Fail-fast: the first non-zero exit in the
  chain stops everything after it. `--dry-run` and `--verbose` cover the
  whole chain, not just the recipe you invoked.
- `ndo add --depends <name>` (repeatable) — sets a recipe's `depends`
  without hand-editing the TOML file.

## [0.4.2] - 2026-08-21

### Fixed

- `ndo completion install`/`uninstall` (and the one-time completion
  prompt) always asked `pwsh` for `$PROFILE` before falling back to
  `powershell`, regardless of which PowerShell edition was actually
  running the invoking session. On a machine with both Windows PowerShell
  5.1 and PowerShell 7+ installed, this could resolve the wrong edition's
  `$PROFILE` and write the completion block to a profile the live session
  never sources — so completion silently appeared to do nothing. Now
  picks which binary to query first from the invoking session's own
  `PSModulePath` (Windows PowerShell's default always includes a
  `WindowsPowerShell` path segment; PowerShell 7+'s doesn't), so it asks
  the edition that actually matches the session running the command.

## [0.4.1] - 2026-08-21

### Fixed

- `DetectChannel`'s Scoop/Homebrew/`GOBIN`/`GOPATH` path matching used
  `filepath.ToSlash`, which only normalizes separators for the *host* OS
  running the process — a no-op for backslash-separated Windows-style
  paths when running on Linux/macOS. Since `DetectChannel` takes an
  explicit `goos` precisely so its tests can exercise Windows-shaped paths
  from any CI runner, this broke `TestDetectChannel/scoop` on Linux CI
  (real-world behavior was unaffected, since production always calls it
  with `goos` matching the actual host). Fixed with an OS-independent
  `toSlash` helper.

## [0.4.0] - 2026-08-21

### Added

- `ndo update` — checks the latest GitHub release against the running
  version. For Homebrew/Scoop/`.deb`/`.rpm`/`go install` installs
  (detected from the running binary's path), prints the right upgrade
  command instead of touching the binary, so that package manager's
  records don't go stale. For a plain binary install (`install.sh` or a
  manual download), downloads the new release, verifies its checksum
  against `checksums.txt`, and atomically replaces the running binary.
  `--check` only reports what's available without installing it.
- `install/install.ps1` and `install/install.cmd` — one-line install for
  Windows (PowerShell and cmd.exe), mirroring `install/install.sh`.

### Changed

- Install scripts moved into `install/` (`install.sh`, `install.ps1`,
  `install.cmd`) instead of living at the repo root.

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

[Unreleased]: https://github.com/green-threads/ndo/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/green-threads/ndo/compare/v0.4.2...v1.0.0
[0.4.2]: https://github.com/green-threads/ndo/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/green-threads/ndo/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/green-threads/ndo/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/green-threads/ndo/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/green-threads/ndo/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/green-threads/ndo/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/green-threads/ndo/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/green-threads/ndo/releases/tag/v0.1.0
