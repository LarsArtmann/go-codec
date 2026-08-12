# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Security

- `normalizeForJSON` (`json_compat_v1.go`) now enforces a `maxNormalizeDepth`
  (100) recursion cap, returning an error instead of recursing indefinitely.
  Closes a stack-exhaustion DoS vector from adversarial deeply-nested CBOR.
  Affects v1 JSON mode only; v2 handles `map[interface{}]interface{}` natively.
- `AutoDetect` (`autodetect.go`) now skips trial-decode for payloads over
  `maxAutoDetectSize` (1 MiB), returning `EncodingRaw` for oversized ambiguous
  input. First-byte heuristic remains O(1) for any size.

### Added

- `SizeResult` struct (`size.go`): `Size` now returns `SizeResult{JSON, CBOR}`
  instead of positional `(int, int)`. More readable at call sites.
- `GetBuffer` / `PutBuffer` (`pool.go`): `sync.Pool[*bytes.Buffer]` helper for
  `BufferEncoder` hot paths — reduces allocation pressure.
- `TestNormalizeForJSON`: table-driven test covering nil, scalar, map, nested,
  int-keyed maps, empty containers, and depth-cap boundary conditions.
- `TestNormalizeForJSON_DepthCap` / `TestNormalizeForJSON_AtMaxDepth`: verify
  the recursion cap rejects >100 depth and accepts exactly 100.
- `FuzzNormalizeForJSON`: fuzz target exercising the normalizer with random
  CBOR-decoded structures.
- `BenchmarkNormalizeForJSON` / `BenchmarkJSONCodec_MarshalUnmarshal`:
  benchmarks comparing v1 vs v2 JSON performance. v2 shows 35-40% lower
  latency and ~50% fewer allocations.
- `export_test.go`: standard Go pattern exposing `CanonicalEncMode`,
  `CanonicalDecMode`, `JSONUnmarshal`, `EnvelopeMagic`, `Envelope`,
  `RawJSONValue` to the external test package.
- `CODEOWNERS`, `SECURITY.md`, `.github/ISSUE_TEMPLATE/bug_report.yml`,
  `.github/PULL_REQUEST_TEMPLATE.md`: repo hygiene files.
- `testpackage` linter re-enabled: 10 test files migrated from `package codec`
  to `package codec_test` (white-box access via `export_test.go`).
- `paralleltest` linter re-enabled: all 19 previously-unparallelized test
  functions now call `t.Parallel()`.

### Changed

- `makezero` config reverted to `always: true` with targeted `//nolint` on the
  one legitimate false positive (`raw.go:46` copy pattern).
- `goconst` and `tagliatelle` re-enabled for `_test.go` files: extracted shared
  test fixture constants (`testdata_test.go`, `testdata_ext_test.go`).
- CI lint job now runs as a v1/v2 matrix; `gitleaks` secret-scan job added.
- README Go version corrected from `"Go 1.23+"` to `"Go 1.26.5+"`.
- `doc.go`: added explicit one-way contract note for `TranscodeToJSON`.
- `CONTRIBUTING.md`: added snapshot-update dual-mode flow and test conventions.

## [0.1.0] - 2026-08-12

First tagged release. Deterministic payload codec library for event-sourced
serialization, tag at `3f8ac9d`.

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

[Unreleased]: https://github.com/larsartmann/go-codec/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/larsartmann/go-codec/releases/tag/v0.1.0
