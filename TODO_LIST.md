# TODO List

> Short-term, actionable, bounded work items. For long-term vision and unrefined
> ideas, see `ROADMAP.md`. Completed items are deleted from this list and logged
> in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                      |
| ---------------- | ------------------------------------------------------------ |
| 🔴 `TODO`        | Not started. Needs doing.                                    |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                    |
| 🔵 `BLOCKED`     | Cannot proceed; external dependency or decision needed.      |

## Active items

| #  | Task                                                                                             | Status        | Impact | Effort | Notes                                                                                                                     |
| -- | ------------------------------------------------------------------------------------------------ | ------------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------- |
| 1  | Decide release strategy and create the GitHub Release (`gh release create`)                      | 🔵 `BLOCKED`  | High   | 5min   | Awaiting user decision: cut `v0.1.1` from HEAD (recommended — `v0.1.0` tag at `3f8ac9d` predates the v1-build fix `d871122`) vs move the tag. Moving a published tag poisons the module proxy. |
| 2  | Implement `DeterministicCodec` marker interface (`signingSafe()` unexported method)              | 🔴 `TODO`     | High   | 30min  | Approved proposal, ~15 lines, zero behavior change: `CBORCodec`/`CBORCompactCodec` always implement it, `JSONCodec` only in the v2 build (`json_compat_v2.go`), `RawCodec` never. Turns silent signature corruption into a compile error. Evidence: `docs/planning/2026-08-14_encryption-signing-cose-architecture-review.md` §4; no code exists (grep `DeterministicCodec` → 0 hits). |
| 3  | Add CI fuzz job (cron, short fuzztime) for all fuzz targets + commit seed corpus for `FuzzAutoDetectDebug_Consistency` under `testdata/fuzz/` | 🔴 `TODO` | Medium | 1-2h   | Corpus dir only holds `FuzzCBORCodec_CanonicalFidelity` (`testdata/fuzz/`); local runs so far were 10s smoke tests. Also closes the old "run fuzz targets 60s in both modes" verification gap. Evidence: `docs/status/2026-08-14_13-54_*` §f-4/f-5. |
| 4  | Convert `normalizeForJSON` depth error to `go-error-family` with stable code (e.g. `codec.normalize_depth_exceeded`) | 🔴 `TODO` | Medium | 30min  | Only non-categorized error in the package; bare `fmt.Errorf` at `json_compat_v1.go:86` breaks the stable-code pattern in `errors.go`. |
| 5  | `BenchmarkObserveCodec` — quantify `ObservableCodec` decorator overhead vs raw codec             | 🔴 `TODO`     | Medium | 30min  | No observability benchmark exists (grep `BenchmarkObserveCodec` → 0). Informs the atomics-vs-RWMutex refactor decision. |
| 6  | Observability edge-case tests: wrap `CBORCompactCodec`, `EncodeToBuffer` inner-error propagation, `buf.Write` failure fallback, `MetricsSnapshot` immutability, nested `ObservableCodec` (no double-count), `ObserveCodec(nil)` behavior, composition with `EncodePooled`, hook byte counts on error paths | 🔴 `TODO` | Medium | 2h     | `observability_test.go` covers the happy paths (9 tests) but none of these edges (grep test names). |
| 7  | v2 streaming test with a non-buffer reader (`strings.Reader` / byte-at-a-time)                    | 🔴 `TODO`     | Medium | 30min  | `streaming_test.go` only exercises `bytes.Buffer`, which masks over-read bugs — the exact class that bit the v2 decoder once. |
| 8  | `ExampleEncodePooled` godoc example                                                              | 🔴 `TODO`     | Low    | 15min  | `EncodePooled` shipped without an example; `example_test.go` has none (grep → 0). |
| 9  | `ExampleSize` godoc example showing `SizeResult`                                                  | 🔴 `TODO`     | Low    | 15min  | `Size` doc comment updated but no `ExampleSize` (grep → 0). Was plan item F65, never done. |
| 10 | Make `ExampleObserveCodec` output size-independent (print derived values, not `bytes=12`)        | 🔴 `TODO`     | Low    | 15min  | Hardcoded CBOR size at `example_test.go:283` breaks on any cbor version bump. |
| 11 | Fix v2 `JSONEncoder` per-call `[]byte{'\n'}` allocation (`io.WriteString` or package-level var)  | 🔴 `TODO`     | Low    | 15min  | `json_compat_v2.go:56` allocates per Encode call. |
| 12 | Add `cbor:"3,keyasint"` to `realisticOrderKeyInt.Items` in benchmarks                            | 🔴 `TODO`     | Low    | 5min   | `benchmark_test.go:467` — slice encodes with string key, skewing the keyasint benchmark. |
| 13 | Measure or soften README/doc.go headline perf claims ("19-43% smaller / 25-72% faster") with a reproducible benchmark citation | 🔴 `TODO` | Low | 30min | Claims predate the benchmark suite (`doc.go:15`, `README.md` §When to Use). Tag-tradeoffs benchmarks exist now; cite them or hedge. |
| 14 | Rename opaque test constants (`testField`, `testFieldE`, `testMapKey` → self-documenting)        | 🔴 `TODO`     | Low    | 15min  | `testdata_test.go:13-16`; `testName`/`testEmail` already exist alongside — finish the job. |
| 15 | Annotate/nolint `json_helpers_v2_test.go` gopls `stdversion` warnings                            | 🔴 `TODO`     | Low    | 15min  | Inherent to dual-build on Go 1.26; currently unannotated. |
| 16 | Add `dependabot.yml` or `renovate` config for dependency updates                                 | 🔴 `TODO`     | Low    | 15min  | `.github/` has no dependency automation (ls confirmed). |
| 17 | Prometheus/OpenTelemetry exporter example (README or `example_test.go`)                          | 🔵 `BLOCKED`  | Low    | 30min  | Needs user decision: dependency-free pseudo-metrics example (keeps the lib dep-light) vs real `prometheus/client_golang` dev-dependency. |
| 18 | CI step: `golangci-lint run --out-format json` artifact to disambiguate LSP-vs-CLI lint truth     | 🔴 `TODO`     | Low    | 15min  | Recurring gopls diagnostic staleness; CLI is authoritative but undocumented in CI. |
| 19 | Add an architecture diagram (codec → store/event/signing/encryption) to README                    | 🔴 `TODO`     | Low    | 20min  | README is text-only; a small diagram or mermaid block clarifies the sibling boundaries. |
| 20 | Streaming benchmarks: `BenchmarkStreamingJSON_Encode/Decode` (+ CBOR streaming; v2 `jsontext.Decoder` vs `json.UnmarshalRead`) | 🔴 `TODO` | Low | 45min  | Quantifies NDJSON overhead and the v2 decoder buffer cost; no streaming benchmark exists today. |
| 21 | `PutBuffer` size guard — reject buffers above a threshold (e.g. 1 MiB) in the pool                 | 🔴 `TODO`     | Low    | 15min  | `pool.go:32-38` pools unconditionally; one huge payload pins a large buffer in the `sync.Pool` forever. |
| 22 | README: add Streaming JSON (NDJSON), `EncodePooled`, and `Size`/`SizeResult` sections               | 🔴 `TODO`     | Low    | 30min  | README documents CBOR streaming and `BufferEncoder` but not the JSON streaming, pooled, or size helpers. |
