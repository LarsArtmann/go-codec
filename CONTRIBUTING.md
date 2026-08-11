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
```

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
