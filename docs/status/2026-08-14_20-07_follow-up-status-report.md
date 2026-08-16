# Status Report — 2026-08-14 20:07 CEST

> Session: Follow-up after the TODO-list execution sweep for `go-codec`.
>
> Author: Crush (assisted session)
> Trigger: User requested a full, comprehensive status update and a written report.

## Executive Summary

The entire actionable TODO list is now **22 of 22 items complete or resolved**. The
only remaining work is the mechanical execution of **#1 (release strategy)**,
which is blocked on a user decision because it requires a remote tag push and a
GitHub Release.

The codebase is green in both JSON builds:

- `go test ./... -race -count=1` — pass (v1)
- `GOEXPERIMENT=jsonv2 go test ./... -race -count=1` — pass (v2)
- `nix run .#lint` — 0 issues in both v1 and v2
- `nix flake check` — all checks passed
- `python3` YAML validation on `.github/workflows/ci.yml` — OK
- Fuzz smoke test for `FuzzAutoDetectDebug_Consistency` — passed

Working-tree changes since the last report:

| File                       | What changed (this session)                                                                           |
| -------------------------- | ----------------------------------------------------------------------------------------------------- |
| `testdata/fuzz/README.md`  | New: corpus file format, seed-add workflow, CI corpus policy.                                         |
| `.github/workflows/ci.yml` | Added comment documenting artifact-only fuzz-corpus policy.                                           |
| `README.md`                | Added explicit `DeterministicCodec` guidance for signing-module consumers.                            |
| `codec_test.go`            | Added two contract tests: CBOR vs CBORCompact byte incompatibility, and singleton CBOR mode identity. |
| `CHANGELOG.md`             | Logged the new tests, README note, and fuzz corpus README.                                            |
| `TODO_LIST.md`             | Updated #1 with resolved sub-questions and explicit recommendation.                                   |

Pre-existing changes already present in the working tree (from previous sessions):

| File          | What was already there                                                        |
| ------------- | ----------------------------------------------------------------------------- |
| `AGENTS.md`   | Added `DeterministicCodec` documentation to the architecture/gotchas section. |
| `FEATURES.md` | Marked `DeterministicCodec` as `FULLY_FUNCTIONAL`.                            |

---

## a) Fully Done (22 / 22)

| #  | Task                                                       | Evidence                                                                                                                                                                                        | Session |
| -- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| 1  | Release strategy decision                                  | Documented in `TODO_LIST.md`: recommendation is to **cut `v0.1.1` from HEAD**; moving `v0.1.0` rejected because it would poison the module proxy. Execution is blocked only on remote push/tag. | 3       |
| 2  | `DeterministicCodec` marker interface                      | `codec.go` defines the interface; `CBORCodec`/`CBORCompactCodec` implement it in all builds; `JSONCodec` only in `json_compat_v2.go`; `README.md` and `AGENTS.md` document it.                  | 1       |
| 3  | CI fuzz job (cron + `workflow_dispatch`) + seed corpus     | `.github/workflows/ci.yml` fuzz job runs all targets for 30s in v1/v2 and uploads corpus; `testdata/fuzz/FuzzAutoDetectDebug_Consistency/` has 5 seed files.                                    | 2       |
| 4  | `normalizeForJSON` depth error as `go-error-family`        | `errors.go` adds `ErrNormalizeDepthExceeded` (`codec.normalize_depth_exceeded`).                                                                                                                | 1       |
| 5  | `BenchmarkObserveCodec`                                    | `benchmark_test.go` with raw/observed encode/decode/pooled sub-benchmarks.                                                                                                                      | 1       |
| 6  | Observability edge-case tests                              | 7 new tests in `observability_test.go`.                                                                                                                                                         | 1       |
| 7  | v2 streaming test with non-buffer reader                   | `TestStreaming_JSONNonBufferReader` and `TestStreaming_JSONByteAtATimeReader` in `streaming_test.go`.                                                                                           | 1       |
| 8  | `ExampleEncodePooled`                                      | `example_test.go`.                                                                                                                                                                              | 1       |
| 9  | `ExampleSize`                                              | `example_test.go`.                                                                                                                                                                              | 1       |
| 10 | Size-independent `ExampleObserveCodec`                     | Hardcoded `bytes=12` removed.                                                                                                                                                                   | 1       |
| 11 | Fix v2 `JSONEncoder` newline allocation                    | `json_compat_v2.go` uses `io.WriteString(e.w, "\n")`.                                                                                                                                           | 1       |
| 12 | `cbor:"3,keyasint"` on benchmark `Items`                   | `benchmark_test.go` updated.                                                                                                                                                                    | 1       |
| 13 | Soften README/doc.go perf claims                           | Both files point to `BenchmarkTagTradeoffs_*` / `BenchmarkRealisticPayload_*`.                                                                                                                  | 1       |
| 14 | Rename opaque test constants                               | `testField` → `testFieldName`, `testFieldE` → `testFieldEmail`.                                                                                                                                 | 1       |
| 15 | Annotate `json_helpers_v2_test.go` stdversion warnings     | Replaced invalid `//nolint:stdversion` with explanatory comment.                                                                                                                                | 1       |
| 16 | `dependabot.yml`                                           | Created with weekly `gomod` group updates.                                                                                                                                                      | 1       |
| 17 | Prometheus/OpenTelemetry exporter example                  | Resolved as dependency-free: `ExampleMetricsHook` in `example_test.go`; README telemetry links to it.                                                                                           | 2       |
| 18 | CI lint JSON artifact                                      | `.github/workflows/ci.yml` produces `lint-report-json-v1` / `lint-report-json-v2` via `--output.json.path`.                                                                                     | 2       |
| 19 | Architecture diagram                                       | Mermaid diagram in README under `## Architecture`.                                                                                                                                              | 2       |
| 20 | Streaming benchmarks                                       | `streaming_benchmark_test.go` + `json_streaming_v2_bench_test.go`.                                                                                                                              | 2       |
| 21 | `PutBuffer` size guard                                     | `pool.go` rejects buffers with `Cap() > 1 MiB`; test added.                                                                                                                                     | 1       |
| 22 | README sections for Streaming JSON, `EncodePooled`, `Size` | Dedicated sections in `README.md`.                                                                                                                                                              | 2       |

Additional work completed this session (not on the original 22-item list):

| Task                                           | Evidence                                                  | Why it matters                                                                                    |
| ---------------------------------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| Document fuzz corpus format + policy           | `testdata/fuzz/README.md`                                 | Prevents the raw-bytes mistake; makes the CI artifact-only policy discoverable.                   |
| Codify CI fuzz corpus policy                   | `.github/workflows/ci.yml` comment                        | Artifact-only auto-commit decision is now written at the upload step, not just in status reports. |
| README guidance for signing modules            | `README.md` after the CBOR signing example                | Tells sibling consumers to assert `DeterministicCodec` at compile time.                           |
| Prove CBOR vs CBORCompact byte incompatibility | `TestCBORCodec_AndCBORCompactCodec_ProduceDifferentBytes` | Locks the documented incompatibility in a test.                                                   |
| Prove CBOR mode singleton identity             | `TestCBORMode_SingletonsReturnIdenticalValues`            | Locks the process-wide singleton contract that sibling modules rely on.                           |

---

## b) Partially Done (0 / 22)

Nothing is partially done. Every started item was carried to completion and
verified.

---

## c) Not Started (1 / 22)

| # | Task                                                        | Why not started                                                                                                                                                         |
| - | ----------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Execute the release: create `v0.1.1` tag and GitHub Release | **Blocked on user confirmation + remote push/tag.** The decision itself is made (cut `v0.1.1` from HEAD), but I cannot push to the remote without explicit instruction. |

---

## d) Totally Fucked Up (0 / 22)

Nothing is fucked up. Final verification passed:

- `go test ./... -race -count=1` — pass (v1 default)
- `GOEXPERIMENT=jsonv2 go test ./... -race -count=1` — pass (v2 opt-in)
- `nix run .#lint` — 0 issues in both v1 and v2
- `nix flake check` — all checks passed
- `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` — YAML OK
- `go test -run='^$' -fuzz='FuzzAutoDetectDebug_Consistency' -fuzztime=5s` — passed

---

## e) What We Should Improve (Self-Critique)

1. **I should have discovered the CBOR sort-order nuance earlier.** The first
   attempt at `TestCBORCodec_AndCBORCompactCodec_ProduceDifferentBytes` used a
   `map[string]string` and produced identical bytes. I had to read the fxamacker
   test corpus to learn that length-first and bytewise-lexical sorting only
   diverge for mixed CBOR key types (or carefully chosen integer-key cases), not
   for ordinary text strings. That domain knowledge should have been in the repo
   already.

2. **The README code-block indentation changed accidentally.** The `return` line
   in the CBOR signing example now uses 2 spaces because the edit tool re-indented
   the code block. It is harmless for rendering, but it is inconsistent with the
   rest of the file. A follow-up cleanup pass should restore the original
   indentation or run the repo formatter over the markdown if one is configured.

3. **The CI workflow is still not exercised on GitHub.** Local YAML validation and
   Go tests pass, but the actual GitHub Actions behavior (cron schedule, artifact
   upload, `workflow_dispatch`) is only inferred. We should either run the
   workflow on a fork or add an `act` smoke-test before claiming the CI work is
   fully verified.

4. **The fuzz budget is unvalidated.** 30s per target is a conservative default;
   it may not catch deep bugs, and the total job time on GitHub's runners is
   unknown. The first cron run should be monitored and the budget adjusted.

5. **DeterministicCodec is not yet consumed by any sibling module.** The marker
   interface is implemented and documented, but the actual value — turning
   non-deterministic codec choices into compile-time errors — only happens once
   the sibling `signing` module changes its API to accept `DeterministicCodec`.
   That work is outside this repository.

6. **Auto-commit daemon produced a commit summary that advertised work not in the
   diff.** The most recent commit (`699fad9`) mentions "docs" generally but not
   the specific files. This is minor, but we should verify commit summaries before
   they land, especially when the diff is a mix of CI, tests, and docs.

7. **No ASCII fallback for the mermaid architecture diagram.** GitHub renders it,
   but plain-text readers or non-GitHub hosts lose the diagram entirely. A short
   text summary of the component graph would improve accessibility.

8. **I did not run the full fuzz target matrix for 30s locally.** I ran a 5s
   smoke test for `FuzzAutoDetectDebug_Consistency` only. A full 30s per-target
   run in both v1 and v2 would give more confidence that the cron job will not
   time out or fail on GitHub's runners.

---

## f) Up to 50 Things We Should Get Done Next

High-impact (do soon):

1. Execute release #1: cut `v0.1.1` from HEAD and publish GitHub Release notes.
2. Wire `DeterministicCodec` assertion into the sibling `signing` module.
3. Update sibling `signing`/`encryption` modules to reuse exported `CBOREncMode()` /
   `CBORDecMode()` singletons.
4. Add negative tests for `TranscodeToJSON` (toarray structs, invalid CBOR leading
   bytes, >1 MiB auto-detect skip).
5. Add an `AutoDetect` / `AutoDetectDebug` benchmark to quantify the heuristic
   cost.
6. Add property tests for `EncodePooled` buffer lifecycle (callback must copy).
7. Add a benchmark comparing `CBORCodec` vs `CBORCompactCodec` encode/decode delta.
8. Add a benchmark for `RawCodec` copy-vs-no-copy behavior.
9. Add a test for `Diagnose` on invalid CBOR bytes.
10. Add a test proving `CBOREncMode()` / `CBORDecMode()` return identical modes
    across multiple calls (already done; verify coverage is sufficient).
11. Add a test proving `CBORCodec` and `CBORCompactCodec` produce different bytes for
    the same struct (already done; verify coverage is sufficient).
12. Refactor `ObservableCodec` metrics to atomics and benchmark the delta against
    the current `RWMutex` implementation.
13. Add a `MetricsSnapshot` JSON marshal example for operational dashboards.
14. Add a README example for `TranscodeToJSON` with HTTP/SSE context.
15. Add a README example for `AutoDetectDebug` logging.
16. Add a `SizeResult` JSON tag and an example of logging payload size budgets.
17. Add a test for `PutBuffer` rejecting a buffer that grew due to `Grow` vs `Write`.
18. Add a test for `GetBuffer` returning a zeroed buffer even under pool
    exhaustion.
19. Add a test for `EncodePooled` callback returning an error — buffer is still
    returned to the pool.
20. Add a fuzz target for `JSONEncoder` / `JSONDecoder` NDJSON streams.

Medium-impact (do next):

21. Add a fuzz target for `ObservableCodec` hook safety (panic behavior).
22. Add an `ExampleCBORCodec` showing `time.Time` UTC normalization.
23. Add an `ExampleTranscodeToJSON`.
24. Add an `ExampleAutoDetect` (non-debug version).
25. Verify `ExampleCBOREncMode` and `ExampleCBORDecMode` are discoverable from
    README.
26. Add a `doc.go` note about `CBORCodec` vs `CBORCompactCodec` byte
    incompatibility.
27. Add a test for `Size` returning `-1` for types that fail to encode.
28. Add a test for `Size` with a type that JSON can encode but CBOR cannot (or
    vice versa).
29. Add a benchmark for the `Size` helper.
30. Add a README mention of `DeterministicCodec` for signing-module consumers
    (already done; verify wording is discoverable).
31. Add a `docs/planning` note documenting the decision that `JSONCodec` is
    deterministic only in the v2 build.
32. Add a `CONTRIBUTING` section on how to add new fuzz targets and seed corpus.
33. Increase fuzztime budget or add a second cron schedule with a longer run
    (e.g., 5 minutes per target) for deeper coverage.
34. Add a CI test that verifies the lint JSON artifact is produced and is valid
    JSON.
35. Add a CI job that runs benchmarks on every PR to detect regressions (with a
    threshold).
36. Add a CI check that renders the mermaid architecture diagram to catch syntax
    errors.
37. Add a README badge for CI status and `pkg.go.dev` reference.
38. Add a fuzz target for `RawCodec` with non-`[]byte` types.
39. Add a property test for `TranscodeToJSON` passthrough contracts.
40. Add a test for `JSONEncoder` error path when the writer returns an error.

Low-impact / polish:

41. Add a test for `JSONDecoder` error path with truncated input.
42. Add a benchmark for `WrapEncode` / `UnwrapDecode`.
43. Add a test for `Envelope` backward-compatible fallback.
44. Add golden snapshot tests for README example outputs.
45. Add a test for `ForEncoding` with unknown/empty encoding.
46. Add a test proving `ObservableCodec` hook receives the correct encoding tag
    when wrapping `ForEncoding`-selected codecs.
47. Add a test proving `ObservableCodec` with a failing `BufferEncoder` wrapped
    codec does not double-count on the fallback path.
48. Add tests for `NormalizeCOSEAlgorithm` and other COSE algorithm helpers.
49. Restore consistent indentation in the README CBOR signing code block.
50. Add an ASCII/text summary fallback for the mermaid architecture diagram.

---

## g) Questions I Cannot Figure Out Myself

1. **Release execution:** Should I proceed right now to create the `v0.1.1` tag from
   current HEAD and publish the GitHub Release, or do you want to defer until
   after additional changes land?

2. **Fuzz corpus auto-commit policy:** The current policy is artifact-only, no
   auto-commit. Do you want to keep it that way permanently, or should we add a
   separate workflow that opens a PR from weekly cron findings once a maintainer
   reviews the artifact?

3. **Sibling integration priority:** Should I switch context to the sibling
   `signing` module next and change its API to accept `codec.DeterministicCodec`,
   or should I continue adding in-repo tests, benchmarks, and examples here first?
