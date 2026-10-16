# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

This project uses [Nix flakes](https://nixos.wiki/wiki/Flakes) for build
automation. If you have Nix installed:

```bash
nix develop          # enter dev shell with Go, golangci-lint, gopls
nix run .#test       # run tests in BOTH JSON modes (v1 and v2)
nix run .#test-race  # race-test both modes
nix run .#lint       # lint both modes
nix run .#build      # build both modes
nix run .#tripwires  # all four repo-consistency check scripts
```

Without Nix, use the plain Go toolchain (Go 1.27.2+; see the section below for
the repository checks without Nix):

```bash
# Build (v1 JSON is the default)
go build ./...
GOEXPERIMENT=jsonv2 go build ./...

# Test
go test ./... -race
GOEXPERIMENT=jsonv2 go test ./... -race

# Lint
golangci-lint run ./...
golangci-lint run --build-tags goexperiment.jsonv2 ./...
```

## Repository Checks Without Nix

The four consistency tripwires are plain bash — no Nix required. Run them
before submitting (CI runs the same scripts on every push, v1 leg):

```bash
bash scripts/check-go-version.sh          # go.mod / .go-version / .golangci.yml agree on one full-patch version
bash scripts/check-features-planned.sh    # FEATURES.md PLANNED symbols must not resolve via go doc
bash scripts/check-error-codes.sh         # errors.go <-> docs/error-codes.md registry sync
bash scripts/check-erraudit-version.sh    # erraudit pin agrees between ci.yml and flake.nix
```

`check-features-planned.sh` needs the Go toolchain on PATH; the other three
are pure text checks. Pin the JSON mode for any `go`-driven work — an ambient
`GOEXPERIMENT=jsonv2` silently flips plain invocations to v2:

```bash
env -u GOEXPERIMENT go test ./...    # v1 (the true default)
GOEXPERIMENT=jsonv2 go test ./...    # v2 (opt-in)
```

## Pre-commit Tripwire Hooks

The same four checks can run automatically before every commit:

```bash
git config core.hooksPath .githooks
```

`.githooks/pre-commit` fast-fails on the first drift with a fix hint. Opt in
once per clone. Bypass with `git commit --no-verify` when you must (CI still
enforces everything on push). If commits stop running hooks entirely, check
that `.githooks/` exists in your checkout — `core.hooksPath` pointing at a
missing directory fails silently.

## Dual JSON Build

This library supports both `encoding/json` (v1, default) and `encoding/json/v2`
(opt-in via `GOEXPERIMENT=jsonv2`). **Never import `encoding/json` or
`encoding/json/v2` directly** — always use the compat helpers in
`json_compat_v*.go`. The contract test (`json_contract_test.go`) enforces this.

Always run tests in **both modes** before submitting changes.

## Golden Snapshots

This project uses `go-snaps` for golden snapshots. If you intentionally change
output format, update the snapshots:

```bash
UPDATE_SNAPSHOTS=true go test ./...
UPDATE_SNAPSHOTS=true GOEXPERIMENT=jsonv2 go test ./...
```

Stale snapshots are automatically cleaned by `snaps.Clean(m)` in `TestMain`.

## Test Conventions

- **External tests preferred**: most test files use `package codec_test` (black-box).
  A few files (`normalize_test.go`, `export_test.go`, `testdata_test.go`) use
  `package codec` (white-box) for testing unexported helpers — these carry
  `//nolint:testpackage`.
- **Parallel by default**: all top-level test functions call `t.Parallel()`.
  Subtests that are safe for parallel execution should too.
- **Shared fixtures**: test constants live in `testdata_ext_test.go` (external)
  and `testdata_test.go` (internal). Reuse existing constants instead of
  duplicating string literals.
- **Both JSON modes**: always test in both v1 and v2 modes. CI runs both
  automatically.

## Fuzzing

Fuzz targets live in dedicated `*_fuzz_test.go` files and run as normal
tests (seed corpus only) on every `go test`. To fuzz a target for real:

```bash
go test -run '^$' -fuzz='FuzzCBORCodec_Roundtrip' -fuzztime=30s
GOEXPERIMENT=jsonv2 go test -run '^$' -fuzz='FuzzCBORCodec_Roundtrip' -fuzztime=30s
```

- Seed corpus files live in `testdata/fuzz/<FuzzTargetName>/` — see
  `testdata/fuzz/README.md` for the two-line file format and the
  `$GOCACHE/fuzz` cache location.
- A weekly CI job (`fuzz` in `.github/workflows/ci.yml`) runs every target
  in both JSON modes and uploads the generated corpus as an artifact
  (artifact-only; new seeds are PR'd manually after review).
- New targets must be added to BOTH target lists in the CI `fuzz` job
  (the v1 list includes v2-incompatible targets such as `FuzzNormalizeForJSON`).
- If a fuzz run finds a crasher, Go writes the input into `testdata/fuzz/`.
  Fix the bug, then keep that file as a committed regression seed with a
  descriptive name (e.g. `seed-number-float64-overflow`).

## Benchmarks

`Benchmark*` targets run as normal tests (compile-checked on every `go test`).
For baseline-grade timing runs on this multi-tenant machine, use the gate —
never a bare `go test -bench` under load (contention inflates ns/op ~3x with
identical allocs):

```bash
scripts/bench-gate.sh OUT_DIR                 # quiet-machine gate + both JSON-mode suites
python3 scripts/bench-compare.py OLD NEW      # sanity-gate two raw outputs (ratio bounds 0.33..3)
```

The reference tables live in [`docs/benchmark-baseline.md`](docs/benchmark-baseline.md);
diff against them with benchstat before accepting performance-sensitive changes.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
