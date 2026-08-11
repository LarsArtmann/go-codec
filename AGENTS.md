# AGENTS.md — go-codec

> Context for AI agents working in this repository. Enduring facts only —
> things hard to discover from reading the code. No changelogs, no task lists.

## What This Is

`go-codec` (`github.com/larsartmann/go-codec`, package `codec`) is the payload
serialization layer for an event-sourcing/CQRS stack. It defines the `Codec`
contract used by event stores, snapshots, event construction, signing, and
encryption across sibling modules. It is published as an independent Go module
designed for consumption by `go-cqrs-lite` and other projects.

It does NOT produce a binary. It is a library only: four codecs (CBOR canonical,
CBOR compact, JSON, Raw) plus shared COSE structure helpers, transcoding,
envelope wrapping, autodetection, streaming, and size/diagnostic utilities.

## Directory Structure

Flat single-package layout — every `.go` file is package `codec` at the repo
root. There are no sub-packages.

```
go-codec/
├── *.go               # all implementation, one package (codec)
├── *_test.go          # tests, fuzz, property, snapshot, benchmarks, examples
├── json_compat_v*.go  # dual-build JSON compat layer (v1 default, v2 opt-in)
├── testdata/          # fuzz corpus + go-snaps golden snapshots
├── flake.nix          # build, test, lint, devShell automation
├── go.mod / go.sum    # module metadata
└── README.md          # user-facing docs
```

## Commands

Plain Go toolchain (Go 1.26+). The project supports both `encoding/json` v1
(default) and v2 (opt-in via `GOEXPERIMENT=jsonv2`):

```bash
go build ./...                          # build (v1 JSON, default)
GOEXPERIMENT=jsonv2 go build ./...      # build (v2 JSON)
go test ./... -race                     # test with race detector (v1)
GOEXPERIMENT=jsonv2 go test ./... -race # test with race detector (v2)
golangci-lint run ./...                 # lint (v1)
golangci-lint run --build-tags goexperiment.jsonv2 ./... # lint (v2)
```

With flake.nix (preferred in LarsArtmann projects):

```bash
nix build                              # build
nix run .#test                         # test both JSON modes
nix run .#test-race                    # race-test both modes
nix run .#lint                         # lint both modes
```

## Architecture

- **`Codec` interface** (`codec.go`) — `Encoding() / Encode / Decode`. The
  optional `BufferEncoder` interface adds zero-alloc `EncodeToBuffer`. `Encoding`
  is a string tag (`"json"`, `"cbor"`, `"raw"`) stamped on payloads so blind
  stores stay self-describing.
- **Dual JSON build.** `json_compat_v1.go` (`//go:build !goexperiment.jsonv2`)
  and `json_compat_v2.go` (`//go:build goexperiment.jsonv2`) provide the same
  five helpers (`jsonMarshal`, `jsonMarshalDet`, `jsonUnmarshal`, `jsonMarshalBuf`,
  `rawJSONValue`) backed by `encoding/json` v1 and v2 respectively. v1 is the
  default; v2 is opt-in via `GOEXPERIMENT=jsonv2`. All source files call helpers,
  never `json.*` directly.
- **CBOR modes are process-wide singletons.** `canonicalEncMode`/`canonicalDecMode`
  (and the compact variants) are computed once via `sync.OnceValue`. Sibling
  modules MUST reuse the exported `CBOREncMode()`/`CBORDecMode()` instead of
  rebuilding identical modes — byte-identical output depends on the exact options.
- **Two CBOR codecs are NOT interoperable.** `CBORCodec` = CanonicalEncOptions
  (RFC 7049, length-first key sort); `CBORCompactCodec` = CoreDetEncOptions
  (RFC 8949, bytewise-lexical sort) + unknown-field rejection on decode. Different
  sort → different bytes → they cannot read each other's data.
- **COSE layer** (`cose.go`/`cose_helpers.go`) — RFC 9052 structure
  marshal/unmarshal for `COSE_Sign1` and `COSE_Encrypt0`, plus `SigStructure` /
  `EncStructure0` builders. This package does NOT perform crypto; it shapes bytes
  for the `signing` and `encryption` modules to sign/encrypt.
- **Errors** use `github.com/larsartmann/go-error-family` with stable codes
  (`codec.raw_encode_type`, `codec.invalid_cose_sign1`, …). See `errors.go`.

## Conventions

- **Indentation:** tabs (see `.editorconfig`). Markdown keeps trailing
  whitespace (`trim_trailing_whitespace = false` for `*.md`).
- **Error wrapping:** codec methods are thin wrappers over `cbor`/`json` and use
  `//nolint:wrapcheck` rather than re-wrapping. Public helpers that orchestrate
  multiple steps (`TranscodeToJSON`, `WrapEncode`, COSE marshal) DO wrap with
  `fmt.Errorf("codec: ...: %w", err)`.
- **Testing stack:** stdlib `testing`; `onsi/gomega` for assertions;
  `pgregory.net/rapid` for property tests; `gkampitakis/go-snaps` for golden
  snapshots (output under `testdata/golden/`); native fuzz targets; godoc
  `Example*` functions (also run as tests). `TestMain` calls `snaps.Clean`.
- **Naming:** codecs are value types (`CBORCodec{}`), constructed zero-valued.

## Gotchas

- **Dual-build: never import `json.*` directly.** All JSON calls go through the
  `json_compat_v*.go` helpers. goimports may corrupt v1 files to import v2 — the
  contract test (`json_contract_test.go`) catches this. Always run both modes.
- **`CBORCodec` ≠ `CBORCompactCodec` bytes.** Never assume data written by one
  round-trips through the other (different key sort + compact rejects unknown
  fields). Document per-store which codec owns the data.
- **`time.Time` must be `.UTC()` before encoding.** CBOR uses `TimeUnixDynamic`
  (float64 epoch); decoded times reconstruct in `time.Local`, not the original
  zone. Wall-clock times must not use `time.Time` (store components + IANA zone).
- **`toarray` / `keyasint` cbor tags are wire-format commitments.** Field order
  and integer keys become part of the bytes; reordering fields or renumbering
  keys breaks existing stored data and invalidates signatures.
- **`AutoDetect` is a heuristic, not a security boundary.** It inspects the
  leading byte and falls back to trial decode. Never use it to skip encoding
  validation — always pair with the matched codec's `Decode`.
- **`TranscodeToJSON` is schema-free.** CBOR maps → JSON objects with
  non-deterministic key order (Go map iteration). `toarray` structs stay arrays
  (field names are lost). Not suitable for byte-deterministic uses.
- **`RawCodec` copies on Decode.** The target `*[]byte` receives a fresh copy so
  callers can't mutate the input buffer. Encode accepts only `[]byte` or
  `rawJSONValue`.

## Dependencies

- `github.com/fxamacker/cbor/v2` — CBOR codec; canonical & core-deterministic modes.
- `github.com/larsartmann/go-error-family` — categorized errors with stable codes.
- `encoding/json` (v1, default) or `encoding/json/v2` (opt-in via `GOEXPERIMENT=jsonv2`).
- Test-only: `onsi/gomega`, `pgregory.net/rapid`, `gkampitakis/go-snaps`.

## High-Value References

| File             | Why it matters                                            |
| ---------------- | --------------------------------------------------------- |
| `doc.go`         | Package-level overview, codec-choice guidance, tag usage  |
| `README.md`      | User-facing usage + when-to-use matrix; sibling cross-refs |
| `FEATURES.md`    | Honest feature inventory with status                      |
| `errors.go`      | Stable error sentinels and codes                          |
