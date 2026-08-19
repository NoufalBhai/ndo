# Security Policy

## Supported versions

`ndo` is pre-1.0. Only the latest released version is supported with
security fixes.

## Reporting a vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, use [GitHub's private vulnerability reporting](https://github.com/green-threads/ndo/security/advisories/new)
for this repository. You should get an acknowledgement within a few days.

If the report is confirmed, we'll work on a fix, credit you (unless you'd
prefer otherwise), and publish a GitHub Security Advisory alongside the
patched release.

## Scope

`ndo` executes shell commands you configure yourself via recipes — that's
its job, not a vulnerability by itself. In-scope reports include things
like: shell-quoting/escaping bugs that let an untrusted argument break out
of `{{param}}` interpolation in an unexpected way, path-resolution bugs in
`NDO_HOME`/local-file search that could load recipes from an unintended
location, or supply-chain issues in the release/install tooling
(`install.sh`, `.goreleaser.yaml`, the Homebrew/Scoop taps).
