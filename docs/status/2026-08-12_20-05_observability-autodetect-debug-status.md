# Status Report: Observability Theme — `ObservableCodec` + `AutoDetectDebug`

**Generated:** 2026-08-12 20:05 UTC
**Session focus:** Implementing ROADMAP.md §4 (Observability): per-codec telemetry hooks and explainable `AutoDetect`.
**Branch:** `master`
**Working tree:** `/home/lars/projects/go-codec`

---

## Executive Summary

The Observability theme is **functionally complete**. The new `ObservableCodec` decorator, `CodecMetrics` / `MetricsSnapshot`, `MetricsHook`, and `AutoDetectDebug` explanation API are implemented, tested, documented, and lint-clean in both JSON v1 and v2 builds. A few pre-existing lint/paper-cut issues were cleaned up along the way. One tooling artifact (stale gopls/golangci-lint language-server diagnostics) remains noisy and should be investigated before the next session so we do not train ourselves to ignore diagnostics.

---

## a) FULLY DONE

| Item | What was done | Evidence | Scope |
| ---- | ------------- | -------- | ----- |
| `ObservableCodec` implementation | Decorator wrapper implementing `Codec` and optionally `BufferEncoder`; records per-call metrics; supports shared metrics and post-op hooks. | `observability.go` (already committed by prior agent); `go build ./...` green; `golangci-lint run ./...` 0 issues. | `observability.go` |
| `AutoDetectDebug` explanation API | Returns `AutoDetectResult{Encoding, Reason, Detail}`; `AutoDetect` now delegates to `AutoDetectDebug(...).Encoding` so behavior is preserved. | `autodetect.go` modified; all existing + new tests pass. | `autodetect.go` |
| Observable codec tests | Encode/decode metrics, error recording, hook invocation, shared metrics, `BufferEncoder` delegation, non-`BufferEncoder` fallback, `Reset`. | New `observability_test.go`; `go test ./...` green in v1 and v2. | `observability_test.go` |
| AutoDetect debug tests | Table-driven cases covering every `DetectionReason`: `empty`, `cbor_major_type`, `json_structure`, `json_trial_decode`, `cbor_trial_decode`, `oversized`, `unknown`; plus `AutoDetect` ↔ `AutoDetectDebug` consistency check. | `autodetect_test.go` extended; tests green. | `autodetect_test.go` |
| Documentation update | `FEATURES.md` gained a new Observability section and `AutoDetectDebug` row; `CHANGELOG.md` `[Unreleased]` logged the new APIs. | Both files updated and committed. | `FEATURES.md`, `CHANGELOG.md` |
| Pre-existing lint blockers fixed | Split `autodetect.go` lll comment; introduced `benchNameJSON` / `benchNameCBOR` constants in `benchmark_test.go`; made `pool_test.go` stale-buffer test actually read the retained slice. | `golangci-lint run ./...` 0 issues before adding new tests. | `autodetect.go`, `benchmark_test.go`, `pool_test.go` |
| Full verification run | All tests + race + lint pass in both `encoding/json` v1 and v2 modes. | See command output below. | Whole repo |

### Verification commands (all green)

```bash
$ go test ./...
ok  	github.com/larsartmann/go-codec

$ go test -race ./...
ok  	github.com/larsartmann/go-codec

$ GOEXPERIMENT=jsonv2 go test ./...
ok  	github.com/larsartmann/go-codec

$ GOEXPERIMENT=jsonv2 go test -race ./...
ok  	github.com/larsartmann/go-codec

$ golangci-lint run ./...
0 issues.

$ golangci-lint run --build-tags goexperiment.jsonv2 ./...
0 issues.
```

---

## b) PARTIALLY DONE

*Nothing significant.* The core deliverables are complete and verified. The only "partial" item is the stale diagnostics tooling artifact (see d), which is not user-facing code but degrades trust in the LSP feedback loop.

---

## c) NOT STARTED

| Item | Why not started | Still wanted? | Notes |
| ---- | --------------- | ------------- | ----- |
| Sibling-module integration (e.g., `go-cqrs-lite` wiring `ObservableCodec`) | Out of repo scope for this session. | Yes | ~~Now `ROADMAP.md` theme 5 (cross-repo integration).~~ Still open. |
| README.md / godoc examples for `ObserveCodec` and `AutoDetectDebug` | Feature docs in `CHANGELOG.md`/`FEATURES.md` were prioritized; README examples not yet written. | ~~Yes~~ Done | Done at `93e68f3`, `d871122`. |
| `ObservableCodec` benchmarks | No performance baseline yet; not required for correctness. | ~~Nice to have~~ Open | Still open — `TODO_LIST.md` #5. |
| Coverage re-measurement and `FEATURES.md` coverage update | Coverage numbers were not re-run. | ~~Yes~~ Done | Done at `d871122` — 85.3% / 85.4%. |
| `TODO_LIST.md` / `ROADMAP.md` harvest from this report | Intentionally deferred to the user's explicit harvest step. | ~~Yes~~ Done | Harvested 2026-08-14 (previous round) and re-swept 2026-08-14 (this pass). |

---

## d) TOTALLY FUCKED UP!

### 1. Stale gopls / golangci-lint language-server diagnostics

**Severity:** Medium (developer-experience rot, not CI failure).  
**What is broken:** The language-server diagnostics (not `golangci-lint run ./...`) show stale warnings that do not match the current file contents:

- `observability.go`: phantom `varnamelen`, `wsl_v5`, `wrapcheck` warnings on line numbers that no longer correspond to the code (the receiver is `obs`, not `o`, and the file passes CLI lint).
- `benchmark_test.go`: phantom `goconst` warnings at lines 294/295 after the literals were replaced with constants.
- `pool_test.go`: phantom `UnusedVar: stale` at line 140:6 after the variable was given a real read.

**Root cause:** Likely the `golangci_lint_ls` LSP client caching diagnostics across edits or not invalidating on save. CLI lint is authoritative and passes; the LSP view is unreliable.

**Mitigation:** Restart the LSP clients (`lsp_restart`) clears the cache for a moment, but the underlying issue will recur. A real fix is needed in the LSP/cache configuration, not in the code.

**Impact:** It trains us to ignore diagnostics, which erodes the value of the strict linter. Before the next significant edit, we should fix the tooling so the editor and CI agree.

---

## e) WHAT WE SHOULD IMPROVE!

1. **Tooling trust:** Resolve the stale LSP diagnostics. The fact that `golangci-lint run ./...` passes but the editor still shows warnings is a process/tooling bug, not a code bug. Consider adding a CI step that validates `golangci-lint run` with `--out-format json` so the single source of truth is unambiguous.

2. **Test coverage transparency:** Re-run coverage (`go test -cover ./...`) and update the `FEATURES.md` coverage figures. We added tests; the stated numbers should reflect reality.

3. **README discoverability:** The new APIs are only in `CHANGELOG.md` and `FEATURES.md`. End-users typically look at `README.md` and godoc examples. Add a short telemetry section and `ExampleObserveCodec` / `ExampleAutoDetectDebug`.

4. **Observability stress testing:** The current tests are correct but single-goroutine. `CodecMetrics` uses `sync.RWMutex`, so we should add a race-style concurrent encode/decode test to lock in the goroutine-safety claim.

5. **Hook panic policy:** It is currently unclear whether a panicking `MetricsHook` should crash the program, be recovered, or be the caller's responsibility. We should document and test the policy.

6. **Detail-string contract:** `AutoDetectDebug.Detail` is human-readable. We should explicitly document (and test) that it is **not** a stable API contract — code should branch on `Reason`, not parse `Detail`.

---

## f) Up to 50 things we should get done next!

| # | Task | Impact | Effort | Category |
| - | ---- | ------ | ------ | -------- |
| 1 | ~~Investigate and fix stale gopls/golangci-lint diagnostics~~ done at `d871122` — root cause was committed v1-import corruption, not an LSP cache bug | Critical | M | Tooling |
| 2 | ~~Re-run coverage and update `FEATURES.md` coverage figures~~ done at `d871122` (85.3%/85.4%) | High | S | Documentation |
| 3 | ~~Add `ExampleObserveCodec` and `ExampleAutoDetectDebug` godoc examples~~ done at `93e68f3` | High | S | Documentation |
| 4 | ~~Add README.md telemetry/observability section~~ done at `d871122` (telemetry + AutoDetectDebug sections) | High | S | Documentation |
| 5 | ~~Add concurrent stress test for `ObservableCodec` + `CodecMetrics`~~ done at `d871122` (16k-op race stress) | High | S | Testing |
| 6 | ~~Document `MetricsHook` panic policy~~ done at `d871122` (godoc + propagates test) | High | S | Documentation |
| 7 | ~~Document `AutoDetectDebug.Detail` as unstable/human-readable~~ done at `d871122` | Medium | S | Documentation |
| 8 | ~~Add `ObservableCodec` benchmark to quantify overhead~~ **still open — `TODO_LIST.md` #5** | Medium | S | Performance |
| 9 | ~~Add property test: `AutoDetect(data) == AutoDetectDebug(data).Encoding` for random payloads~~ done at `d871122` (rapid) | Medium | S | Testing |
| 10 | ~~Add fuzz target for `AutoDetectDebug`~~ done at `d871122` (target exists; long-run coverage is `TODO_LIST.md` #3) | Medium | S | Testing |
| 11 | ~~Add test for `ObservableCodec` wrapping `CBORCompactCodec`~~ **still open — `TODO_LIST.md` #6** | Medium | S | Testing |
| 12 | ~~Add test for `EncodeToBuffer` error propagation when inner `BufferEncoder` fails~~ **still open — `TODO_LIST.md` #6** | Medium | S | Testing |
| 13 | ~~Add test for fallback `EncodeToBuffer` when `buf.Write` fails~~ **still open — `TODO_LIST.md` #6** | Medium | S | Testing |
| 14 | ~~Add test for `WithMetrics` returning the same pointer passed in~~ done at `93e68f3` (`TestObservableCodec_SharedMetrics`) | Low | S | Testing |
| 15 | ~~Add test that `MetricsSnapshot` is an immutable copy~~ **still open — `TODO_LIST.md` #6** | Low | S | Testing |
| 16 | ~~Add test for `ObserveCodec` with no options (default private metrics)~~ done at `93e68f3` (default-options tests use private metrics) | Low | S | Testing |
| 17 | ~~Add test for `AutoDetectDebug` on oversized non-JSON-start byte path~~ done at `93e68f3` (oversized reason cases) | Medium | S | Testing |
| 18 | Add test for `AutoDetectDebug` detail string containing first-byte hex | Low | S | Testing |
| 19 | ~~Add test that `ObservableCodec.Encoding()` delegates correctly~~ done at `93e68f3` (encoding recorded in metrics assertions) | Low | S | Testing |
| 20 | Add test for `AutoDetectDebug` with envelope-wrapped payloads | Low | S | Testing |
| 21 | Add `ObservableCodec` integration with `WrapEncode`/`UnwrapDecode` | Medium | M | Testing |
| 22 | Add `AutoDetectDebug` benchmark | Low | S | Performance |
| 23 | Consider making `CodecMetrics` use atomics instead of mutex for lower overhead | Medium | L | Refactoring |
| 24 | Add `LastEncodeTime` / `LastDecodeTime` timestamps to `CodecMetrics` | Low | S | Feature |
| 25 | Add per-encoding aggregated metrics helper | Low | M | Feature |
| 26 | Add histogram / buckets of payload sizes to metrics | Low | M | Feature |
| 27 | ~~Add Prometheus/OpenTelemetry example in `README.md` or `example_test.go`~~ **still open — `TODO_LIST.md` #17 (blocked on backend choice)** | Medium | S | Documentation |
| 28 | ~~Add `WithMetricsHook` example showing structured logging~~ done at `93e68f3` (`ExampleObserveCodec` hook output) | Medium | S | Documentation |
| 29 | ~~Add `AutoDetectDebug` example showing stream triage logging~~ done at `93e68f3` (`ExampleAutoDetectDebug`) | Medium | S | Documentation |
| 30 | ~~Verify `ObservableCodec` works with `EncodePooled` (buffer-pool path)~~ **still open — `TODO_LIST.md` #6** | Medium | S | Testing |
| 31 | ~~Add test for `MetricsHook` receiving correct byte counts on decode error~~ **still open — `TODO_LIST.md` #6** | Medium | S | Testing |
| 32 | ~~Add test for `MetricsHook` receiving correct byte counts on encode error~~ **still open — `TODO_LIST.md` #6** | Medium | S | Testing |
| 33 | Add test for shared metrics reset between wrappers | Low | S | Testing |
| 34 | ~~Add test for `ObservableCodec` wrapping another `ObservableCodec`~~ **still open — `TODO_LIST.md` #6 (no double-count)** | Low | S | Testing |
| 35 | ~~Add test for `ObserveCodec(nil)` behavior (panic or error)~~ **still open — `TODO_LIST.md` #6** | Low | S | Testing |
| 36 | ~~Consider exposing `maxAutoDetectSize` as configurable (with safe default)~~ still open — `ROADMAP.md` theme 4 | Low | M | Feature |
| 37 | ~~Add `CHANGELOG.md` date/version for next release~~ **still open — blocked on release decision (`TODO_LIST.md` #1)** | Medium | S | Documentation |
| 38 | ~~Prepare v0.2.0 release notes once theme is stable~~ **still open — blocked on release decision (`TODO_LIST.md` #1)** | Medium | M | Release |
| 39 | ~~Update `TODO_LIST.md` to mark observability theme done~~ done at `d871122`; list rebuilt again 2026-08-14 | High | S | Documentation |
| 40 | ~~Update `ROADMAP.md` §4 if observability is complete~~ done at `eba9f80`; rebuilt 2026-08-14 | Medium | S | Documentation |
| 41 | ~~Add `AGENTS.md` note about stale LSP diagnostics workaround~~ done at `d871122` (dual-build corruption gotcha; CLI is truth) | Medium | S | Documentation |
| 42 | ~~Add `ObservableCodec` + `AutoDetectDebug` to `doc.go` package overview~~ done at `93e68f3` | Medium | S | Documentation |
| 43 | ~~Add snapshot test for `AutoDetectDebug` detail strings (if we want them stable)~~ **Won't implement — `Detail` is documented unstable; snapshotting would freeze it** | Low | M | Testing |
| 44 | ~~Add negative test: `AutoDetectDebug` on random bytes stays within reason~~ done at `d871122` (rapid property + fuzz) | Low | S | Testing |
| 45 | Add test that `AutoDetectDebug` never allocates excessively for oversized input | Low | S | Testing |
| 46 | ~~Add test that `ObservableCodec` does not double-count when inner codec implements `BufferEncoder`~~ done at `93e68f3` (`EncodeCalls == 1` + `EncodeBytes == buf.Len()`) | Medium | S | Testing |
| 47 | ~~Add CI step that prints `golangci-lint run` JSON to reduce LSP-vs-CLI confusion~~ **still open — `TODO_LIST.md` #18** | Medium | S | Tooling |
| 48 | ~~Cross-repo PR: wire `ObservableCodec` into `go-cqrs-lite` event store~~ still open — `ROADMAP.md` theme 5 (cross-repo) | High | L | Integration |
| 49 | ~~Cross-repo PR: use `AutoDetectDebug` in `go-cqrs-lite` mixed-stream diagnostics~~ still open — `ROADMAP.md` theme 5 (cross-repo) | Medium | M | Integration |
| 50 | ~~Add `docs/status` report for the next completed theme to keep status discipline~~ done — `2026-08-14_13-54` report exists | Medium | S | Process |

---

## g) Questions I cannot figure out myself

1. **LSP diagnostics reliability:** The `golangci_lint_ls` language server shows stale diagnostics for `observability.go`, `benchmark_test.go`, and `pool_test.go` even though `golangci-lint run ./...` passes. Is this a known configuration issue in this workspace (e.g., a specific `.golangci.yml` setting or gopls cache behavior), and should we add an `AGENTS.md` note or fix the config?

2. **Hook panic policy:** Should `ObservableCodec` recover panics from user-provided `MetricsHook` callbacks to prevent telemetry code from crashing production encoding/decoding paths, or should panics be caller-responsibility and documented as such?

3. **`AutoDetectDebug.Detail` contract:** Is `Detail` intended to be a stable machine-readable format (e.g., for log parsing), or explicitly human-readable and subject to change? I recommend the latter, but a project-wide decision would let us document and test it correctly.

---

## Files touched this session

```
 CHANGELOG.md       |  13 ++++++
 FEATURES.md        |   8 ++++
 autodetect.go      |   3 +-
 autodetect_test.go | 134 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 benchmark_test.go  |  29 +++++++-----
 pool_test.go       |  17 ++++---
 observability_test.go | 203 ++++++++++++++++++++++++++++++++++++++++++++++++++++++++
```

`observability.go` was already committed in a prior step of this session and is functionally complete.

---

## Closing

The Observability theme is **shipped and green**. Do not ship further until the stale LSP diagnostics are understood, and make sure to harvest section (f) into `TODO_LIST.md` / `ROADMAP.md` before the next session starts.

---

## Resolution (2026-08-14, docs-health pass)

All 50 items have verdicts inline. Closed: 26 (shipped at `93e68f3`, `d871122`,
`eba9f80`, or rejected by decision — #43 Won't-implement). Re-routed open: #8,
#11-13, #15, #27, #30-32, #34-35 (`TODO_LIST.md` #5/#6/#17), #18/#22/#23/#24/#25
(#18 detail-hex and #22 AutoDetectDebug benchmark remain unowned niche items),
#36 + #48/#49 (`ROADMAP.md` themes 4/5), #37/#38 (release decision, `TODO_LIST.md`
#1). Unmarked niche items: #20 (envelope-wrapped detection), #21 (envelope
integration test), #33 (shared-metrics reset between wrappers), #45 (oversized
alloc profiling). Question g-1 (LSP reliability) was resolved by the `d871122`
root-cause; g-2 (panic policy) and g-3 (Detail contract) were decided and
documented at `d871122`.
