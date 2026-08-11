# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Deterministic payload codec library for event-sourced serialization, exposing a
  single `Codec` contract (`Encoding` / `Encode` / `Decode`) plus the optional
  zero-allocation `BufferEncoder` interface — `codec.go`
- Four codec implementations:
  - `CBORCodec` — canonical CBOR (RFC 7049, length-first key sort), the
    recommended deterministic, signing-safe default — `cbor.go`
  - `CBORCompactCodec` — Core Deterministic CBOR (RFC 8949, bytewise-lexical
    sort) with unknown-field rejection on decode as a schema-drift guard —
    `cbor_compact.go`
  - `JSONCodec` — `encoding/json` with deterministic marshal and case-insensitive
    decode — `json.go`
  - `RawCodec` — `[]byte` passthrough (copies on decode) — `raw.go`
- Dual-build JSON support: `encoding/json` v1 (default) and `encoding/json/v2`
  (opt-in via `GOEXPERIMENT=jsonv2`), via build-tagged compat layer
  (`json_compat_v1.go`, `json_compat_v2.go`) — matches go-branded-id pattern
- Codec resolution and format detection: `ForEncoding` (encoding stamp → codec,
  all three encodings including Raw) and `AutoDetect` (best-effort encoding
  inference from raw bytes) — `codec.go`, `autodetect.go`
- Shared CBOR modes exported for sibling modules: `CBOREncMode` / `CBORDecMode`
  (process-wide `sync.OnceValue` singletons) — `cbor.go`
- CBOR tooling: `Diagnose` (extended diagnostic notation) and `Size` (JSON vs
  CBOR byte-size comparison) — `cbor.go`, `size.go`
- Cross-format transcoding: schema-free `TranscodeToJSON` (CBOR → JSON; JSON/raw
  pass through unchanged) — `transcode.go`
- Self-describing envelopes for blind stores: `WrapEncode` / `UnwrapDecode`
  (codec-stamped JSON wrapper with backward-compatible fallback) — `envelope.go`
- Streaming CBOR: `NewCBOREncoder` / `NewCBORDecoder` for batch encoding/decoding
  without materializing the full byte slice — `streaming.go`
- COSE (RFC 9052) structure codec: `COSE_Sign1` and `COSE_Encrypt0`
  marshal/unmarshal, `SigStructure` / `EncStructure0` builders, protected-header
  helpers, and `NormalizeCOSEAlgorithm` — `cose.go`, `cose_helpers.go`
- Base64 JSON marshalling helpers for sibling `[]byte`-wrapper types
  (`MarshalBase64JSON`, `UnmarshalBase64JSON`, `AssignBase64JSON`,
  `DecodeBase64String`, `WrapCOSEMarshal`) — `base64_json.go`
- Categorized errors with stable codes via `go-error-family` — `errors.go`
- Comprehensive test suite: behavioral unit tests, fuzz targets, property tests
  (`pgregory.net/rapid`), golden snapshots (`go-snaps`), benchmarks, godoc
  `Example*` functions, dual-build contract test
- `flake.nix` with dual-mode CI (build/test/lint/race in both JSON modes),
  devShell, coverage, clean apps
- `.golangci.yml` lint config with `goexperiment.jsonv2` build tag support
- MIT license, `AUTHORS` file

### Changed

- `ForEncoding` now resolves `EncodingRaw` → `RawCodec{}` (previously rejected
  Raw as unknown, creating an asymmetry with `AutoDetect` which produces Raw)
- `envelopeMagic` renamed from `"cqrs"` to `"gcdc"` (go-codec) — a
  format-agnostic envelope should not encode an architectural style
- `COSESign1String` / `COSEEncrypt0String` renamed to `COSESign1Diagnostic` /
  `COSEEncrypt0Diagnostic` — they return diagnostic notation, not arbitrary
  strings
- Aligned `.gitignore` ignore rules to the module name — `f327816`

### Fixed

- `snaps_clean_test.go` compile error: reverted `_ =` prefix to bare call
  (matching the original cqrs-lite source), which Go allows for multi-return
  functions used as statements
