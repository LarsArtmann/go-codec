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

- `ObservableCodec` / `ObserveCodec` / `CodecMetrics` / `MetricsSnapshot` /
  `MetricsHook` (`observability.go`): opt-in, decorator-based telemetry for any
  `Codec`. Records per-operation encode/decode call counts, byte totals, error
  counts, and last errors; implements `BufferEncoder` when the wrapped codec does.
  `MetricsHook` enables push-style telemetry (Prometheus, OpenTelemetry) without
  polling. Goroutine-safe via `sync.RWMutex` — `observability_test.go`.
- `AutoDetectDebug` / `AutoDetectResult` / `DetectionReason` (`autodetect.go`):
  explainable version of `AutoDetect` that returns not only the inferred encoding
  but also a stable `DetectionReason` (`empty`, `cbor_major_type`,
  `json_structure`, `json_trial_decode`, `cbor_trial_decode`, `oversized`,
  `unknown`) and a human-readable `Detail` string for triage and logging. The
  original `AutoDetect` now delegates to `AutoDetectDebug(...).Encoding`, preserving
  behavior — `autodetect_test.go`.
- `EncodePooled` (`pool.go`): callback-based pool-backed encode helper that
  manages `GetBuffer`/`EncodeToBuffer`/`PutBuffer` lifecycle automatically.
  Eliminates per-call `[]byte` allocation in hot paths where the caller
  processes encoded bytes immediately (e.g., writing to a store).
- `NewJSONEncoder` / `NewJSONDecoder` (`streaming.go`): streaming JSON
  encoder/decoder for parity with `NewCBOREncoder` / `NewCBORDecoder`. Uses
  NDJSON (newline-delimited JSON): each `Encode` writes one JSON value followed
  by a newline; `Decode` reads one value at a time. Dual-build (v1 wraps
  `json.Encoder`/`json.Decoder`; v2 wraps `json.MarshalWrite`/`json.UnmarshalRead`).
- `BenchmarkTagTradeoffs_Encode` / `BenchmarkTagTradeoffs_Decode`: comprehensive
  benchmarks comparing map (default), `toarray`, and `keyasint` CBOR encoding
  across small (3-field), medium (7-field), and large (12-field) payload shapes.
  Includes size comparison logging.
- `BenchmarkCBORReflectionCache`: benchmark measuring cold (first encode) vs
  warm (cached) CBOR performance. Documents that fxamacker/cbor's internal
  `sync.Map` cache eliminates the need for code generation.
- `BenchmarkEncodePooled`: benchmark comparing pool-backed encode vs plain
  `Encode` across JSON, CBOR, and CBOR compact codecs.
- `pool_test.go`: tests for `GetBuffer`, `PutBuffer`, and `EncodePooled`
  (reset, nil-safety, capacity retention, CBOR/JSON round-trip, callback error,
  buffer-stale-after-return).
- Streaming JSON tests: `TestStreaming_JSON`, `TestStreaming_JSONEncoderMultiple`,
  `TestStreaming_JSONNewlineDelimited` — `streaming_test.go`.
- `ExampleNewJSONEncoder`: godoc example for NDJSON streaming — `example_test.go`.
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

- `TestObservableCodec_ConcurrentStress`: 16,000 concurrent encode/decode
  operations across a shared `CodecMetrics` and `MetricsHook` under `-race`,
  locking in the goroutine-safety claim with exact call/byte/hook-count
  assertions — `observability_test.go`.
- `TestObservableCodec_HookPanicPropagates`: locks the documented `MetricsHook`
  panic policy (propagates to caller; metrics recorded before the hook, so
  counters stay consistent) — `observability_test.go`.
- `TestProperty_AutoDetectDelegatesToDebug` (rapid) and
  `FuzzAutoDetectDebug_Consistency` (native fuzz): arbitrary payloads keep
  `AutoDetect` and `AutoDetectDebug(...).Encoding` in lockstep, always return a
  known encoding and reason, and never panic — `autodetect_test.go`.
- `ExampleObserveCodec` / `ExampleAutoDetectDebug`: godoc examples for the
  observability APIs — `example_test.go`.

### Fixed

- **Default (v1 JSON) build was broken at HEAD**: `json_compat_v1.go` and
  `json_helpers_v1_test.go` had been corrupted to import `encoding/json/v2`
  while retaining `!goexperiment.jsonv2` build tags, so the default toolchain
  could not compile the package at all and the JSON contract test failed.
  Restored the `encoding/json` imports (caught by `json_contract_test.go` —
  the guard did its job).
- Documented `MetricsHook` panic policy (propagates, not recovered) and
  `AutoDetectResult.Detail` as unstable human-readable prose (`Reason` is the
  stable contract) in godoc and README.

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
