# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [v0.2.0] — 2026-08-16

### Security

- `normalizeForJSON` (`json_compat_v1.go`) now enforces a `maxNormalizeDepth`
  (100) recursion cap, returning an error instead of recursing indefinitely.
  Closes a stack-exhaustion DoS vector from adversarial deeply-nested CBOR.
  Affects v1 JSON mode only; v2 handles `map[interface{}]interface{}` natively.
- `AutoDetect` (`autodetect.go`) now skips trial-decode for payloads over
  `maxAutoDetectSize` (1 MiB), returning `EncodingRaw` for oversized ambiguous
  input. First-byte heuristic remains O(1) for any size.
- Go toolchain raised 1.26.5 → 1.26.6 (`go.mod`, `.go-version`,
  `.golangci.yml`): closes the `GO-2026-5972` stdlib vulnerability
  (`encoding/asn1` recursion, reachable via the fxamacker/cbor `EncMode` path
  from `CBORCodec.EncodeToBuffer`) that had the CI `govulncheck` job failing
  every push. Consumers now need a ≥1.26.6 toolchain (auto-downloaded under
  default `GOTOOLCHAIN=auto`). Hermetic Nix checks stay on 1.26.5 until nixpkgs
  packages 1.26.6 — tracked in `TODO_LIST.md`.

### Added

- `DecodeEnvelopeOrLegacy[T]` (`envelope.go`): one-call decode for data that
  may be envelope-wrapped or raw (pre-envelope). Envelope data decodes via its
  stamped codec; raw data via the configured codec with exactly one JSON↔CBOR
  cross-retry, so legacy rows read correctly regardless of which standard codec
  wrote them or which is configured now. Custom codecs get the configured-codec
  attempt only. Errors return unwrapped for caller classification —
  `envelope_legacy_test.go`.
- `scripts/check-go-version.sh` (+ CI step): single-source tripwire enforcing
  that `go.mod`, `.go-version`, and `.golangci.yml` `run.go` agree on the Go
  version — the same drift-lock pattern as the FEATURES tripwire.
- CI now validates `lint-report.json` with `jq` before uploading, closing the
  artifact-validity gap (recovered 20-07 follow-up item #34).
- `docs/benchmark-baseline.md`: 10-run benchstat reference baseline (v1 mode,
  go1.26.5, 67 sub-benchmarks). FEATURES performance figures upgraded from
  indicative `~` values to baseline-cited means.
- `TODO_LIST.md`: recovered lost mermaid-diagram CI render-check item (20-07
  follow-up item #36, never shipped nor routed); closed the 2026-08-12 `gosec`
  item as already covered — `.golangci.yml` enables `gosec` (G304/G115
  excluded) and the lint matrix runs it on every push.
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
- `ExampleMetricsHook`: dependency-free push-style metrics example using a
  simple counter map inside a `MetricsHook` — `example_test.go`.
- Architecture diagram (mermaid) added to README, showing the codec contract and
  sibling module boundaries (event, signing, encryption, storage/pebble, kv).
- README sections for Streaming JSON (NDJSON), Pooled Encoding (`EncodePooled`),
  and Size Comparison (`Size` / `SizeResult`).
- Streaming benchmarks: `BenchmarkStreamingJSON_Encode/Decode`,
  `BenchmarkStreamingCBOR_Encode/Decode`, and a v2-only
  `BenchmarkStreamingJSONV2_DecoderComparison` comparing `jsontext.Decoder` (the
  path used by `NewJSONDecoder`) against `json.UnmarshalRead` per line —
  `streaming_benchmark_test.go`, `json_streaming_v2_bench_test.go`.
- CI fuzz job: weekly cron (`0 2 * * 0`) plus `workflow_dispatch` that runs all
  fuzz targets for 30s in both JSON v1 and v2 modes and uploads the generated
  corpus as an artifact — `.github/workflows/ci.yml`.
- Seed corpus for `FuzzAutoDetectDebug_Consistency` committed under
  `testdata/fuzz/FuzzAutoDetectDebug_Consistency/`.
- `testdata/fuzz/README.md`: documents the Go fuzz corpus file format (two-line
  text file with `go test fuzz v1` header and `[]byte(...)` literal) and the
  CI corpus policy (artifact upload, no auto-commit).
- `TestCBORCodec_AndCBORCompactCodec_ProduceDifferentBytes`: proves the two
  CBOR codecs are not wire-compatible by encoding a map with mixed-length integer
  keys and asserting different bytes.
- `TestCBORMode_SingletonsReturnIdenticalValues`: proves `CBOREncMode` and
  `CBORDecMode` are process-wide singletons and return identical values on
  repeated calls.
- `DeterministicCodec` (`codec.go`): marker interface (`Codec` plus unexported
  `signingSafe()`) identifying codecs whose `Encode` output is byte-deterministic
  and therefore safe for cryptographic signing. Satisfied by `CBORCodec` and
  `CBORCompactCodec` in every build and by `JSONCodec` in the opt-in v2 build
  only — signing with non-deterministic v1 JSON becomes a compile-time error.
  Implements the approved proposal in
  `docs/planning/2026-08-14_encryption-signing-cose-architecture-review.md` §4.
- DeterministicCodec satisfaction-matrix tests: `TestDeterministicCodec_CBORAndCompactAlwaysSatisfy`
  and `TestDeterministicCodec_RawNeverSatisfies` (every build) plus build-tagged
  `TestDeterministicCodec_JSONCodecV1DoesNotSatisfy` /
  `TestDeterministicCodec_JSONCodecV2Satisfies` — runtime-assert exactly which
  codecs satisfy the marker interface per build, so the compile-time signing
  contract cannot silently drift — `deterministic_codec_test.go`,
  `deterministic_codec_v1_test.go`, `deterministic_codec_v2_test.go`.
- `FuzzStreamingJSON_NDJSONRoundtrip` (`streaming_fuzz_test.go`): fuzzes the
  NDJSON streaming contract (value framing survives encode/decode round-trip,
  including byte-at-a-time readers). Seeded with the `1e700` regression corpus
  entry (`testdata/fuzz/FuzzStreamingJSON_NDJSONRoundtrip/`): decoding into
  `any` overflows float64 on out-of-range numbers, so the fuzz target decodes
  into `RawJSONValue` to keep values byte-exact.
- `FuzzObservableCodec_HookSafety` (`observability_fuzz_test.go`): arbitrary
  payloads through the `ObservableCodec` decorator with `MetricsHook` attached —
  metric counts and error bookkeeping stay consistent and never panic.
- `scripts/check-features-planned.sh`: FEATURES.md drift tripwire — fails when
  a symbol documented as `PLANNED` actually resolves in the package, so the
  feature inventory cannot lag behind shipped code. Wired into CI (v1 test
  job); self-tested to pass clean and fail on injected drift.
- `doc.go`: codec-choice guidance now covers `DeterministicCodec` (which codecs
  are signing-safe in which build) and an explicit warning that `CBORCodec` and
  `CBORCompactCodec` bytes are not interchangeable.
- CI: coverage summary step per JSON mode, and the two new fuzz targets
  (`FuzzStreamingJSON_NDJSONRoundtrip`, `FuzzObservableCodec_HookSafety`) added
  to both the v1 and v2 fuzz matrices — `.github/workflows/ci.yml`.
- Benchmarks: `BenchmarkAutoDetect` / `BenchmarkAutoDetectDebug`
  (`autodetect_benchmark_test.go`), `BenchmarkWrapEncode` /
  `BenchmarkUnwrapDecode` (`envelope_benchmark_test.go`), `BenchmarkSize`
  (`size_benchmark_test.go`), and `BenchmarkCBORCompact_vs_Canon_Decode`
  (`benchmark_test.go`, compact vs canonical CBOR decode cost).
- Edge/error-path tests: `TestDiagnose_InvalidCBOR` (`codec_test.go`), 8-case
  `TestNormalizeCOSEAlgorithm` table (`cose_test.go`),
  `TestStreaming_JSONEncoderWriterError` /
  `TestStreaming_JSONDecoderTruncatedInput` via `failingWriter`
  (`streaming_test.go`), and `TestObservableCodec_HookEncodingTagWithForEncoding`
  — the hook reports the wrapped codec's encoding tag across `ForEncoding`
  resolution (`observability_test.go`).
- Godoc examples: `ExampleAutoDetect`, `ExampleTranscodeToJSON`,
  `ExampleDeterministicCodec`, `ExampleCBORCodec_time`, `ExampleMetricsSnapshot`
  — `example_test.go`.
- Docs: `CONTRIBUTING.md` fuzzing section (local fuzz commands + corpus
  policy), `README.md` ASCII summary of the architecture diagram,
  `testdata/fuzz/README.md` `$GOCACHE/fuzz` crash-corpus location guide,
  `.go-version` pinned to `1.26.5`, two new High-Value References rows in
  `AGENTS.md`, and inline resolution annotations on the historical reports
  `docs/status/2026-08-14_17-29_*` and `2026-08-14_18-09_*`.

### Fixed

- `TestObservableCodec_MetricsSnapshotImmutability` now asserts the mutated
  snapshot fields are visible on the returned copy (the writes were previously
  never read — `unusedwrite` smell at `observability_test.go:560`).
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

- `SizeResult` (`size.go`) now carries explicit JSON tags (`json` / `cbor`):
  JSON-serialized keys change from `JSON` / `CBOR` to lowercase. The type is a
  sizing diagnostic; no stored-format impact.
- `makezero` config reverted to `always: true` with targeted `//nolint` on the
  one legitimate false positive (`raw.go:46` copy pattern).
- `goconst` and `tagliatelle` re-enabled for `_test.go` files: extracted shared
  test fixture constants (`testdata_test.go`, `testdata_ext_test.go`).
- CI lint job now runs as a v1/v2 matrix; `gitleaks` secret-scan job added.
- README Go version corrected from `"Go 1.23+"` to `"Go 1.26.5+"`.
- `doc.go`: added explicit one-way contract note for `TranscodeToJSON`.
- `CONTRIBUTING.md`: added snapshot-update dual-mode flow and test conventions.
- CI workflow hardened: GitHub Actions pinned to commit SHAs, `golangci-lint`
  v2.12.2 installed explicitly, and a `govulncheck` vulnerability-scan job
  added alongside the gitleaks secret scan.
- CI lint job now produces a JSON report artifact (`lint-report-json-v1` /
  `lint-report-json-v2`) via `--output.json.path`, uploaded even when lint fails
  so LSP-vs-CLI diagnostic differences can be resolved from the authoritative
  CLI output.
- README telemetry example now links to the dependency-free `ExampleMetricsHook`.
- `flake.nix`: hermetic `checks.build` / `checks.test` via `buildGoModule`
  (dependencies fetched through the Nix sandbox), plus a `packages.default`
  so plain `nix build` works. Fixes the `/homeless-shelter` sandbox failure
  that made `nix flake check` red.
- `docs/planning/2026-08-14_encryption-signing-cose-architecture-review.md`:
  architecture decisions on COSE extraction (no), JWS/JWE (no), signing /
  encryption module structure (keep separate), and the approved
  `DeterministicCodec` marker-interface proposal.
- Docs: docs-health audit pass — remaining multi-line `~~` annotation spans in
  the older status reports (`2026-08-11_23-38`, `2026-08-12_03-24`,
  `2026-08-12_09-25`, `2026-08-12_12-42`) migrated to the renderer-safe
  per-line form; README CI badge added; FEATURES performance table extended
  with `BenchmarkRealisticPayload_*` and `BenchmarkObserveCodec` rows;
  `docs/DOMAIN_LANGUAGE.md` compat-helper count corrected (five helpers, was
  four names); forward-looking items harvested from the 20:47/20:58 reports
  into `TODO_LIST.md` and `ROADMAP.md`.

### Performance

- `UnwrapDecode` (`envelope.go`) now sniffs the first byte: any value ≥
  `cborMinMajorType` (0x80 — CBOR arrays, maps, tags, simple values) can never
  begin valid JSON, so the doomed envelope parse is skipped and the data falls
  straight through to the fallback codec. The backward-compat read path for
  blind stores holding pre-envelope CBOR drops from ~181ns / 184B / 6 allocs to
  ~1.6ns / 0B / 0 allocs per read (-99% time, -100% allocs, n=10 benchstat);
  the wrapped-envelope path is statistically unchanged (p=0.912). Behavior is
  byte-identical by construction — a successful envelope parse requires a JSON
  object, which always starts with `{` (0x7B). New benchmark:
  `BenchmarkUnwrapDecode_FallbackRawCBOR`; sniff pinned by
  `TestUnwrapDecode_FirstByteSniff` (all 128 high bytes), plus
  `_EmptyData` and `_RawCBORScalarsBelowSniffThreshold` edge cases.

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
