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
```

Without Nix, use the plain Go toolchain (Go 1.26+):

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

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
