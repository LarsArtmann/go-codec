# Status Report: Performance & Allocation Discipline

> 2026-08-12 12:42 — Session executing ROADMAP section 2 (Performance & allocation discipline)

---

## Executive Summary

Implemented all four raw ideas from ROADMAP section 2: buffer-pool-backed
encode helper, CBOR reflection caching investigation, toarray/keyasint
benchmarks, and streaming JSON encoder/decoder. All tests pass in both JSON
modes (v1 + v2) with `-race`. One critical v2 streaming bug was caught and
fixed during self-review.

---

## a) FULLY DONE

### 1. Buffer-pool-backed encode helper (`EncodePooled`)

**Status: FULLY DONE — production-ready, tested, benchmarked**

- `EncodePooled(enc BufferEncoder, v any, fn func([]byte) error) error` in
  `pool.go:40` — callback-based API managing `GetBuffer` → `EncodeToBuffer` →
  `fn` → `PutBuffer` lifecycle automatically
- 7 tests in `pool_test.go`: reset verification, nil-safety, round-trip,
  CBOR+JSON round-trip, callback error propagation, buffer-stale-after-return
- `BenchmarkEncodePooled` in `benchmark_test.go` comparing pool vs plain Encode
  across JSON, CBOR, CBOR compact

### 2. Streaming JSON encoder/decoder

**Status: FULLY DONE — production-ready, tested in both modes, example added**

- `NewJSONEncoder(w io.Writer) *JSONEncoder` and
  `NewJSONDecoder(r io.Reader) *JSONDecoder` in `streaming.go`
- `JSONEncoder` / `JSONDecoder` types defined in dual-build files:
  - `json_compat_v1.go`: wraps `json.Encoder` / `json.Decoder` (v1 stdlib
    streaming, handles NDJSON natively)
  - `json_compat_v2.go`: uses `json.MarshalWrite` per value + manual `\n` for
    encoding, `jsontext.NewDecoder` + `json.UnmarshalDecode` for decoding
- NDJSON convention: each `Encode` writes one JSON value + `\n`
- 3 streaming tests: single round-trip, multiple encodes, NDJSON batch
- Godoc `ExampleNewJSONEncoder` in `example_test.go`

### 3. toarray / keyasint benchmarks

**Status: FULLY DONE — comprehensive, documented, sizes logged**

- `BenchmarkTagTradeoffs_Encode` and `BenchmarkTagTradeoffs_Decode` in
  `benchmark_test.go`
- 9 sub-benchmarks each: map vs toarray vs keyasint × small (3 fields) /
  medium (7 fields) / large (12 fields)
- `b.Log` output with byte sizes and percentage savings for each shape
- Key findings:
  - toarray: smallest payloads (23-41% size reduction vs map)
  - keyasint: close behind (18-37% reduction)
  - toarray: fewest allocs on decode (field names eliminated)
  - keyasint: slightly slower on decode (integer key lookup overhead)

### 4. CBOR reflection caching investigation

**Status: FULLY DONE — root cause verified, documented, benchmarked**

- Investigated `fxamacker/cbor/v2@v2.9.2/cache.go` — confirmed process-wide
  `sync.Map` caches keyed by `reflect.Type` (lines 24-27)
- `BenchmarkCBORReflectionCache` in `benchmark_test.go` — cold vs warm
- Measured: cold ~117µs / 104 allocs → warm ~340ns / 2 allocs (344x faster)
- Conclusion documented in `doc.go`: "Code generation is NOT needed — the
  cache handles it"
- Remaining idea noted in ROADMAP: pre-warming cache at startup for consistent
  first-request latency

### 5. Documentation updates

**Status: FULLY DONE — all project docs updated**

- `doc.go`: Added Streaming, EncodePooled, and Performance sections with
  measured data and guidance
- `FEATURES.md`: Added EncodePooled, JSON streaming, performance benchmarks
  section
- `CHANGELOG.md`: Added all new features under `[Unreleased]`
- `ROADMAP.md`: Section 2 updated — completed items marked, remaining ideas
  listed
- `AGENTS.md`: Added streaming, buffer pool, and performance architecture
  notes

---

## b) PARTIALLY DONE

### v2 JSON streaming encoder — asymmetric implementation

The v2 `JSONEncoder` uses `json.MarshalWrite` per value (stateless write) +
manual `\n` write. The v1 `JSONEncoder` wraps `json.Encoder` (stateful, handles
newlines internally). Both produce identical NDJSON output, but the
implementations are structurally different. This is unavoidable given the v2
API design (`jsontext.Encoder` corrupts NDJSON by inserting separator tokens
between top-level values — verified and documented in the code comments). Not
a defect, but worth noting as an asymmetry that could confuse maintainers.

### Decode-side buffer pool

Only the encode path has pool-backed helpers (`EncodePooled`, `GetBuffer` /
`PutBuffer`). The decode path takes `[]byte` directly, so the caller owns the
read buffer. A decode-side pool would need a different shape (pooled
`bytes.Reader` or similar). Listed as a remaining idea in ROADMAP.

---

## c) NOT STARTED

Nothing from the original four roadmap items remains unstarted. All four were
implemented.

---

## d) TOTALLY FUCKED UP

### CRITICAL BUG: v2 JSON streaming was broken (caught and fixed during self-review)

**What happened:** The initial v2 `JSONEncoder` used `json.MarshalWrite` per
value (correct), but the v2 `JSONDecoder` used `json.UnmarshalRead` per value
(incorrect for streaming). `json.UnmarshalRead` creates a new internal buffer
on each call, causing it to over-read from the `io.Reader` — bytes consumed
into the internal buffer but not part of the current JSON value are lost
between calls. This made sequential `Decode` calls fail silently (the second
call returns EOF because the reader position has jumped past the second value).

A second attempt used `jsontext.NewEncoder` + `json.MarshalEncode` for the
encoder. This was also broken: `jsontext.Encoder` inserts separator tokens
(`{`) between top-level values, producing `{"ID":1}\n{{"ID":2}\n\n` instead of
`{"ID":1}\n{"ID":2}\n`. Verified via a standalone test program.

**Final fix:**

- Encoder: `json.MarshalWrite(w, v)` + `w.Write([]byte{'\n'})` (stateless per
  value, no encoder state corruption)
- Decoder: `jsontext.NewDecoder(r)` + `json.UnmarshalDecode(dec, v)` (stateful
  streaming decoder, maintains read position across calls)

**Why this matters:** The initial implementation was "tested" only with v1
(`go test ./... -race` passed). The v2 mode (`GOEXPERIMENT=jsonv2 go test
./... -race`) was NOT run until the self-review. This is a process failure:
the AGENTS.md explicitly says "Always run both modes," and I didn't follow my
own rule until prompted to self-review.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Run v2 tests immediately after v1 tests.** The v2 streaming bug existed
   for the entire session and was only caught during self-review. Both modes
   should be tested after every change that touches the dual-build files.
   Institutionalized — CI runs the dual-mode test/lint matrix (`ef1f4f4`).

2. **The `nix run .#test` command runs both modes but was never used.** The
   flake.nix has `nix run .#test` which tests both JSON modes. Using it would
   have caught the v2 bug immediately. Plain `go test ./...` only tests v1.
   Institutionalized — AGENTS.md Commands list the nix apps first.

3. **More structural test for v2 streaming.** The streaming tests use
   `bytes.Buffer` which is the simplest case. A test with a non-buffer reader
   (e.g., `strings.Reader` or `io.PipeReader`) would catch over-read issues
   that only manifest with non-buffer readers. — ~~identified~~ done at
   `2c98116` (`TestStreaming_JSONNonBufferReader`, `TestStreaming_JSONByteAtATimeReader`).

### Code improvements

4. **`JSONEncoder` allocates `[]byte{'\n'}` per `Encode` call.** This is a
   per-call allocation for a single byte. Extract to a package-level `var` or
   use `io.WriteString(w, "\n")` which may avoid the allocation. — ~~open~~
   done at `2c98116` (v2 encoder uses `io.WriteString(e.w, "\n")`).

5. **`EncodePooled` error wrapping is inconsistent with the thin-wrapper
   convention.** It wraps the encode error with `fmt.Errorf("codec: pooled
   encode failed: %w", err)` but passes the callback error through unwrapped.
   This is intentional (the callback owns its error) but could confuse users
   who expect all errors from `EncodePooled` to be wrapped. — resolved by the
   error contract: encode failures wrap with `codec.pooled_encode` (Rejection),
   callback errors pass through unclassified (`docs/error-codes.md`, ADR-0001
   at `888cf8b`).

6. **No streaming benchmark for JSON.** CBOR streaming has no benchmark
   either, but JSON streaming has a measurable overhead from the newline
   write and the v2 `jsontext.Decoder` internal buffer. A
   `BenchmarkStreamingJSON` would help quantify this. — ~~open~~ done at
   `699fad9` (`BenchmarkStreamingJSON_Encode/Decode`, CBOR siblings too).

7. **`BenchmarkCBORReflectionCache` doesn't isolate cold cache.** Running
   with `-benchtime=1x` shows the cold cost, but the cache is warm for all
   subsequent runs in the same process. A true cold-cache benchmark would need
   a fresh process per measurement (or type manipulation to evict from cache). —
   done (covered) — FEATURES marks the cold figure "one-off, not in the
   baseline suite"; methodology documented in `docs/benchmark-baseline.md`.

8. **`realisticOrderKeyInt` struct in benchmarks is missing `keyasint` on the
   `Items` field.** The `Items` field is `[]orderItem` and doesn't have a
   `cbor:"3,keyasint"` tag. This means the items slice is encoded with a
   string key "Items" instead of integer key 3, making the benchmark not
   fully representative of a pure keyasint payload. — ~~open~~ done —
   verified in code (`benchmark_test.go:517`, `Items` carries the tag).

### Documentation improvements

9. **`doc.go` performance numbers are hardcoded.** The size/speed tradeoff
   table in `doc.go` is derived from benchmark runs on one machine. These
   numbers will vary by CPU/architecture. A note saying "measured on AMD
   Ryzen AI MAX+ 395, your results may vary" would be more honest. —
   NOT-DO/DUPLICATE — README ("measured numbers on your hardware") and the
   FEATURES baseline note carry the caveat.

10. **No godoc example for `EncodePooled`.** `ExampleEncodePooled` would help
    users discover the API. The `ExampleBufferEncoder` exists but doesn't
    show the pooled variant. — ~~open~~ done at `2c98116`
    (`ExampleEncodePooled` in `example_test.go`).

---

## f) Up to 50 things to get done next

### High impact (should do soon)

1. ~~Add `ExampleEncodePooled` godoc example to `example_test.go`~~ done at `2c98116`
2. ~~Add streaming benchmark `BenchmarkStreamingJSON_Encode/Decode` in
   `benchmark_test.go`~~ done at `699fad9` (`streaming_benchmark_test.go`)
3. ~~Fix `JSONEncoder.Encode` to avoid per-call `[]byte{'\n'}` allocation (use
   `io.WriteString` or a package-level `var`)~~ done at `2c98116`
4. ~~Fix `realisticOrderKeyInt` to add `cbor:"3,keyasint"` on the `Items` field~~ done — verified in code (`benchmark_test.go:517`)
5. ~~Add a v2-specific streaming test using `strings.Reader` (non-buffer reader)
   to catch over-read issues that `bytes.Buffer` masks~~ done at `2c98116` (`TestStreaming_JSONNonBufferReader`)
6. ~~Run `golangci-lint run --build-tags goexperiment.jsonv2 ./...` to verify~~
   ~~v2 lint is clean for new files~~ done at `ef1f4f4` (v2 lint is a CI matrix leg)
7. ~~Add `nolint:wrapcheck` consistency audit — `EncodePooled` wraps the encode
   error but not the callback error; document this decision~~ done (covered) — codified in the error contract (`codec.pooled_encode` Rejection wrap; callback errors unclassified — `docs/error-codes.md`, ADR-0001 `888cf8b`)

### Medium impact (should do eventually)

8. ~~Add `DecodePooled` helper — pooled `bytes.Reader` for decode paths where
   callers have a `[]byte` and want to avoid materializing a reader~~ ← open — `ROADMAP.md` theme 2
9. ~~Add pre-warming helper for CBOR type cache (`WarmTypeCache[T]()` or
   similar) for startup-time consistency~~ ← open — `ROADMAP.md` theme 2
10. ~~Add streaming CBOR benchmark (`BenchmarkStreamingCBOR_Encode/Decode`)~~ done at `699fad9`
11. ~~Add streaming JSON benchmark for v2 specifically, comparing
    `jsontext.Decoder` vs `json.UnmarshalRead` performance~~ done at `699fad9` (`BenchmarkStreamingJSONV2_DecoderComparison`)
12. ~~Document the v1/v2 `JSONEncoder` implementation asymmetry in a code
    comment in `streaming.go` (currently only documented in `json_compat_v2.go`)~~ done (covered) — asymmetry documented in `json_compat_v2.go` + the AGENTS.md v2-NDJSON gotcha (`2c98116`)
13. ~~Add a fuzz target for `EncodePooled` (callback that copies bytes, verify
    round-trip)~~ ← open — nice-to-have (no consumer demand)
14. ~~Add a fuzz target for JSON streaming (NDJSON round-trip with random
    values)~~ done at `d9b30ff` (`FuzzStreamingJSON_NDJSONRoundtrip`)
15. ~~Add `go-snaps` golden snapshot for NDJSON streaming output (locks the
    wire format)~~ **Won't implement — framing is locked byte-exact by `FuzzStreamingJSON_NDJSONRoundtrip` (`d9b30ff`); golden prose would only add churn.**
16. ~~Add `io.Closer` or `Flush` method to `JSONEncoder` for non-buffer writers
    that need explicit flushing~~ ← open — nice-to-have
17. ~~Consider whether `JSONEncoder` should implement `BufferEncoder` interface
    (it doesn't return `[]byte`, but the interface could be extended)~~ **Won't implement — the codecs themselves implement `BufferEncoder`; the streaming wrapper is writer-oriented by design.**
18. ~~Add a `NewJSONEncoderWithOptions` variant for deterministic JSON output
    (v2 `json.Deterministic(true)`)~~ ← open — nice-to-have (v2-only concern)
19. ~~Benchmark `EncodePooled` vs `BufferEncoder` + manual pool to quantify the
    callback overhead~~ done at `094de50` (`BenchmarkEncodePooled`, pool vs plain; manual-pool variant not added)
20. ~~Add `PutBuffer` size guard — reject buffers larger than a threshold (e.g.,
    1MB) to prevent the pool from holding oversized buffers (memory leak
    prevention)~~ done at `2c98116` (`maxPoolBufferSize`, `TestPutBuffer_RejectsOversizedBuffers`)

### Documentation & polish

21. ~~Add `README.md` section for streaming JSON (currently only CBOR streaming
    is documented)~~ done at `699fad9`
22. ~~Add `README.md` section for `EncodePooled` (currently only in `doc.go`)~~ done at `699fad9`
23. ~~Add a "Performance" section to `README.md` summarizing the benchmark
    findings (cold vs warm cache, toarray/keyasint tradeoffs)~~ **NOT-DO/DUPLICATE — FEATURES.md owns the performance inventory; README cites benchmarks inline.**
24. ~~Update `doc.go` benchmark numbers with a "your results may vary" caveat~~ **NOT-DO/DUPLICATE — README and FEATURES carry the re-run-on-your-hardware guidance.**
25. ~~Add `AGENTS.md` gotcha about v2 `jsontext.Encoder` corrupting NDJSON (so
    future maintainers don't repeat the mistake)~~ done (`2c98116` era — the gotcha lives in AGENTS.md)
26. ~~Add `AGENTS.md` note about always running `nix run .#test` or both~~
    ~~`go test` + `GOEXPERIMENT=jsonv2 go test` after touching dual-build files~~
    done — covered by the pre-existing AGENTS.md dual-build gotcha ("Always run
    both modes")
27. ~~Add `CONTRIBUTING.md` note about testing streaming in both JSON modes~~
    done — covered by CONTRIBUTING.md's "always test in both v1 and v2 modes"

### CI & tooling

28. ~~Verify `nix run .#lint` passes for both modes (v2 lint wasn't run this~~
    ~~session)~~ done at `ef1f4f4` (CI) and re-verified in later sessions
29. ~~Add a CI check that runs benchmarks and compares against a baseline
    (regression detection)~~ ← open — `ROADMAP.md` theme 2
30. ~~Add `benchstat` to flake.nix for benchmark comparison~~ ← open — `ROADMAP.md` theme 2 (benchstat used for the baseline doc; flake packaging open)
31. ~~Consider adding a `nix run .#bench` app that runs all benchmarks in both
    modes~~ ← open — `ROADMAP.md` theme 2

### API hardening

32. ~~Add `io.WriterTo` / `io.ReaderFrom` implementations on `JSONEncoder` /
    `JSONDecoder` for `io.Copy` compatibility~~ ← open — nice-to-have
33. ~~Consider whether `JSONEncoder` should implement `io.Closer` to flush
    remaining buffer on close~~ **NOT-DO/DUPLICATE — same proposal as #16 (which stays open as the canonical row).**
34. ~~Add `Encoding()` method to `JSONEncoder` / `JSONDecoder` for consistency
    with the `Codec` interface (enables dispatch tables)~~ ← open — nice-to-have
35. ~~Consider a generic `StreamCodec` interface that unifies CBOR and JSON
    streaming~~ ← open — nice-to-have
36. ~~Add `EncodeStream` / `DecodeStream` convenience functions that take a
    `Codec` and `io.Writer` / `io.Reader` and return a streaming encoder /
    decoder (dispatch by `Encoding()`)~~ ← open — nice-to-have
37. ~~Consider whether `RawCodec` needs streaming (it's just `[]byte` — probably
    not, but worth documenting why not)~~ **Won't implement — raw passthrough has nothing to frame; no stream API applies.**

### Testing depth

38. ~~Add property test for `EncodePooled` — rapid-generated payloads, verify
    callback receives valid encoded bytes~~ ← open — nice-to-have
39. ~~Add concurrent `EncodePooled` test (multiple goroutines using the pool
    simultaneously with `-race`)~~ ← open — nice-to-have (`sync.Pool` is goroutine-safe by contract)
40. ~~Add stress test for JSON streaming — 10,000 values through NDJSON
    round-trip~~ done (covered) — `FuzzStreamingJSON_NDJSONRoundtrip` (`d9b30ff`) exercises arbitrary-length streams
41. ~~Add test for `JSONEncoder` with `io.PipeWriter` (non-buffer writer) to
    verify no buffering issues~~ ← open — nice-to-have
42. ~~Add test for `JSONDecoder` with partial reads (reader that returns one
    byte at a time) to verify the decoder handles fragmented input~~ done at `2c98116` (`TestStreaming_JSONByteAtATimeReader` + fuzz)
43. ~~Add test for empty JSON stream (zero values encoded, immediate EOF on
    decode)~~ ← open — nice-to-have
44. ~~Add test for malformed NDJSON (missing newline, double newline, value
    spanning newlines)~~ done (covered) — fuzz (`d9b30ff`) + `TestStreaming_JSONDecoderTruncatedInput` (`094de50`)
45. ~~Add benchmark comparing `BenchmarkRealisticPayload_Encode` results against
    the new `BenchmarkTagTradeoffs_Encode` to verify consistency~~ **NOT-DO/DUPLICATE — both benchmarks exist and FEATURES cross-cites them; a comparison harness adds no signal.**
46. ~~Add `BenchmarkStreamingCBOR` and compare against `BenchmarkStreamingJSON`
    for batch encode/decode throughput~~ done at `699fad9`
47. ~~Add test that `EncodePooled` works with a custom `BufferEncoder`
    implementation (not just the built-in codecs)~~ ← open — nice-to-have
48. ~~Add test that `JSONEncoder.Encode` produces output that `JSONDecoder.Decode`
    can read (cross-v1/v2 compatibility — v1 encode, v2 decode and vice versa)~~ **Won't implement — v1/v2 never coexist in one binary (build tags); framing equality is fuzz-locked.**
49. ~~Add benchmark for `JSONEncoder.Encode` overhead (the `MarshalWrite` +
    `\n` write) vs raw `json.MarshalWrite`~~ done (covered) — streaming benchmarks (`699fad9`) quantify the newline+MarshalWrite path; the v2 comparison bench shows the jsontext delta
50. ~~Consider adding `pprof` integration for benchmark memory profiling
    (detect allocation regressions automatically)~~ ← open — nice-to-have

---

## g) Questions I can NOT figure out myself

### 1. Should `EncodePooled` wrap the callback error?

Currently `EncodePooled` wraps the encode error (`fmt.Errorf("codec: pooled
encode failed: %w", err)`) but passes the callback error through unwrapped.
The reasoning: the callback owns its error semantics, and wrapping it with
"codec: pooled encode failed" would be misleading (the encode succeeded, the
callback failed). But this means errors from `EncodePooled` have inconsistent
wrapping. Should I wrap the callback error too, or document the asymmetry as
intentional?

**Resolved** — the asymmetry is now the documented contract: encode failures
carry the `codec.pooled_encode` code (Rejection family), callback errors pass
through unclassified (`docs/error-codes.md`, ADR-0001 at `888cf8b`).

### 2. Should the `JSONEncoder` / `JSONDecoder` types be exported from the package root or stay in the compat files?

Currently `JSONEncoder` and `JSONDecoder` are defined in `json_compat_v1.go`
and `json_compat_v2.go` (build-tagged). They are exported types. The
constructors (`NewJSONEncoder` / `NewJSONDecoder`) are in `streaming.go`. This
means the type definitions move with the build tag, but the constructors are
stable. Is this the right structural choice, or should the types be defined in
`streaming.go` with only the internal implementation in the compat files? The
CBOR streaming types (`*cbor.Encoder` / `*cbor.Decoder`) come from the external
library, so there's no precedent in this codebase for where to put
library-owned streaming types.

**Resolved** — kept in `json_compat_v*.go`: the type definitions must switch
with the build tag; constructors stay stable in `streaming.go` (decided during
the `2c98116` dual-build refresh).

### 3. Should I add a decode-side pool helper, or is that YAGNI?

The encode side has `EncodePooled` (callback-based, manages the buffer pool
lifecycle). The decode side takes `[]byte` directly — the caller owns the
input buffer. A decode-side pool would need to pool `bytes.Reader` or similar,
and the callback shape would be different (the caller needs to provide the
input `[]byte` anyway). Is there a real use case for this, or is it premature?
The ROADMAP lists it as a "remaining raw idea" but I'm not sure it's worth
building until a consumer actually needs it.

**Resolved — YAGNI confirmed.** Remains a raw idea in `ROADMAP.md` theme 2
until a consumer asks.

---

## Resolution (2026-08-14, docs-health pass; completed 2026-09-26)

Items 6, 26, 27, and 28 are resolved inline above; the 2026-09-26 docs-health
pass resolved the remainder inline (done at hash, Won't implement, or explicit
open-routing). Route summary: shipped — #1-5, #10-12, #14, #19-22, #25, #40,
#42, #44-46, #49; Won't implement — #15, #17, #33, #37, #48; NOT-DO/DUPLICATE
— #23, #24, #45; open → ROADMAP theme 2 — #8, #9, #29-31; open nice-to-haves —
#13, #16, #18, #32, #34-36, #38, #39, #41, #43, #47, #50. The v2
`jsontext.Encoder` NDJSON corruption gotcha lives in AGENTS.md.
