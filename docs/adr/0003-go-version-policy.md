# ADR-0003: Go Version Contract — One Full-Patch String, Three Declarations

**Status:** Accepted
**Date:** 2026-10-16
**Supersedes:** none
**Related:** ADR-0001 (error taxonomy), `scripts/check-go-version.sh` (the
enforcing tripwire), `scripts/write-go-version.sh` (the generator), the
AGENTS.md "nixpkgs lags Go patch releases" gotcha

## Context

The Go toolchain version appears in three declarations that must agree:

| File                     | Role                                                                    |
| ------------------------ | ----------------------------------------------------------------------- |
| `go.mod` `go`            | **Authoritative.** Consumer-visible minimum floor; CI setup-go reads it |
| `.go-version`            | Tooling pin (editors, direnv, version managers)                         |
| `.golangci.yml` `run.go` | Linter's target Go version (golangci-lint analyzes against it)          |

Before this policy the contract lived only in an AGENTS.md gotcha, and drift
happened in practice: `.golangci.yml` once carried a quoted value
(`go: "1.27.1"`) that made an early tripwire comparison vacuously pass, and a
two-part `go 1.27` in go.mod would silently compare equal to nothing. Each
drift class was only patched into the tripwire script after it fired.

The floor is also load-bearing beyond style: golangci-lint's analyzers are
coupled to the stdlib they analyze (the 2026-10-15 `buildir` panic was
golangci v2.12.2 analyzing the go1.27.1 stdlib), the dual JSON build only
works on ≥1.27 (`encoding/json/v2`), and hermetic Nix builds pin the exact
toolchain via `goPkg`. A drifted declaration makes one of these layers lie.

## Decision

1. **One full-patch string.** All three declarations carry the identical
   full-patch version (e.g. `1.27.2`). Two-part versions (`go 1.27`) are
   forbidden in every declaration: the floor is a patch-level commitment,
   because stdlib security fixes ship as patch releases (the 1.27.1 → 1.27.2
   bump for GO-2026-6617/6613/6611/6603 is the standing example).

2. **`go.mod` is the single source of truth.** Tooling and docs derive from
   it; nothing else is edited first. A bump is one atomic change touching all
   three declarations — never a partial commit.

3. **Enforcement is mechanical, not customary.**
   `scripts/check-go-version.sh` fails the build (CI v1 leg, pre-commit hook,
   `nix run .#tripwires`) when the three disagree, when any value is quoted,
   or when any value is not three-part. The script's strictness (quote
   stripping before compare, two-part rejection with an explanatory error)
   encodes the two historical drift incidents; do not simplify it.

4. **Bump procedure.** Bump all three together (or run
   `scripts/write-go-version.sh <version>`), require a CHANGELOG entry — the
   floor is consumer-visible and hermetic builders re-vendor against it — and
   confirm nixpkgs ships the matching `go_1_N` before tagging. Since flake.nix
   pins `goPkg = pkgs.go_1_27` and overrides `buildGoModule`/`checks.format`
   with it, only nixpkgs' _default_ go can lag, which affects auxiliary
   tooling, not the package/build/format checks.

5. **Generation over hand-editing (direction).** The generator
   (`scripts/write-go-version.sh`) writes all three declarations from one
   argument; the tripwire remains the verification of record. Hand edits are
   acceptable but the generator is the blessed path (see AGENTS.md).

## Consequences

- Drift between the three files is a build failure, not a review catch.
- Patch bumps are routine and cheap: three-line diff, one changelog bullet.
- Consumers always see an honest, maximal floor: `go get` on an older
  toolchain fetches the pinned one via `GOTOOLCHAIN=auto` instead of silently
  mis-analyzing.
- The linter pin (`golangci-lint@v2.14.0` in ci.yml) is a separate contract
  with its own floor guard (`scripts/check-golangci-pin.sh`); this ADR covers
  only the Go version triangle.
