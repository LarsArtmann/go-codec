# Status Report: Superb-Quality-Plan Execution (T1–T10 done, T11 partial) — ~~T11 partial~~ T11 completed same day

**Generated:** 2026-08-14 20:47 CEST
**Session type:** Pareto-planned quality sweep ("Make this project superb")
**Scope:** Execute every verified gap from the two 20:07 reports' next-task lists.
**Plan artifact:** `docs/planning/2026-08-14_20-16-SUPERB-quality-plan.html`
(D2 source + SVG alongside; graph inlined, self-contained).

---

## Executive Summary

Planned via pareto-planning (11 tasks, 38 fine steps), verified every candidate
gap against the tree first (~40% of the 20:07 §f lists was already shipped —
dropped), then executed **T1–T10 fully** and **T11 partially** (interrupted by
this report request). 10 of 11 tasks complete; the remainder is a small docs
sync (`CHANGELOG`/`FEATURES` entries) plus the final `nix flake check`.
~~T11 partially~~ T11 completed later on 2026-08-14: CHANGELOG + FEATURES
entries written and `nix flake check` passed (all checks).

All quality gates verified green after every task: build + race-test + lint in
**both JSON modes** (0 issues). Coverage rose **85.3% → 86.3% (v1)** and
**85.4% → 88.8% (v2)**. ~~86.3% (v1)~~ final v1 figure is **88.0%** (86.3 was
measured before the T7–T10 test files landed; FEATURES.md records 88.0/88.8).
The new NDJSON fuzz target earned its keep
immediately: it found `1e700` — a valid JSON number that overflows float64 on
decode-into-`any` (stdlib asymmetry, not a library bug) — now a committed
regression seed with a tightened framing contract.

The auto-commit daemon took T1–T6 as `d9b30ff`; T7–T10 sit in the working
tree (12 modified + 4 untracked paths), verified but uncommitted.

---

## a) FULLY DONE (with evidence)

| # | What | Evidence |
| - | ---- | -------- |
| P | Pareto plan (tiers 1%/4%/20%, 38 steps, D2 graph → SVG → HTML) | `docs/planning/2026-08-14_20-16-SUPERB-quality-plan.html` |
| T1 | **DeterministicCodec build-tagged satisfaction-matrix tests** — v1 asserts JSON NOT satisfied, v2 asserts satisfied; CBOR/CBORCompact always, Raw never. The four-doc claim now enforced by CI in both modes. | `deterministic_codec{,_v1,_v2}_test.go`; both modes PASS |
| T2 | **FEATURES.md drift tripwire** — `scripts/check-features-planned.sh` fails when a ⚪ PLANNED symbol resolves via `go doc`; self-tested (PASS on current tree, FAIL on injected fake row); legend-skip bug fixed; wired into CI (test job, v1 leg). Kills the 20:07 §d-1 drift class permanently. | script + `.github/workflows/ci.yml`; YAML validated; drift-exit=1 verified |
| T3 | **doc.go guidance gaps** — DeterministicCodec paragraph in Choosing-a-Codec + CBOR↔CBORCompact byte-incompatibility warning. | `doc.go`; gofmt/vet clean both modes |
| T4 | **godoc examples ×5** — `ExampleAutoDetect`, `ExampleTranscodeToJSON`, `ExampleDeterministicCodec` (compiles in BOTH builds via CBOR), `ExampleCBORCodec_time` (UTC round-trip), `ExampleMetricsSnapshot`. | `example_test.go`; both modes PASS |
| T5 | **fuzz ×2 + CI wiring** — `FuzzStreamingJSON_NDJSONRoundtrip` (framing + byte-at-a-time buffering independence = the v2 over-read class) and `FuzzObservableCodec_HookSafety` (hook fires exactly once/op; counters exact; no panics). Both added to CI fuzz matrix v1+v2. 10s runs clean in both modes (1.2M+ execs each). **Finding:** `1e700` crasher → target tightened to decode into `RawJSONValue`, regression seed committed as `seed-number-float64-overflow`. | `streaming_fuzz_test.go`, `observability_fuzz_test.go`, `testdata/fuzz/…/seed-number-float64-overflow`, ci.yml |
| T6 | **CI coverage reporting** — `-coverprofile` + `go tool cover -func` total per matrix leg. Locally verified: 86.3% v1 / 88.8% v2 (up from 85.3/85.4). | ci.yml test job; YAML validated |
| T7 | **benchmarks ×4** — `BenchmarkAutoDetect` (+Debug), `BenchmarkWrapEncode`/`BenchmarkUnwrapDecode`, `BenchmarkSize`, `BenchmarkCBORCompact_vs_Canon_Decode` (speed delta; only size existed). All `b.Loop()` style per repo convention. | `autodetect_benchmark_test.go`, `envelope_benchmark_test.go`, `size_benchmark_test.go`, `benchmark_test.go`; smoke-run green both modes |
| T8 | **edge tests ×5** — `TestNormalizeCOSEAlgorithm` (8-case table incl. uint64-overflow + non-integer errors), `TestStreaming_JSONEncoderWriterError`, `TestStreaming_JSONDecoderTruncatedInput`, `TestObservableCodec_HookEncodingTagWithForEncoding` (json/cbor/raw legs), `TestDiagnose_InvalidCBOR`. Plan deviation: the "PutBuffer Grow vs Write" 6th test was **skipped as duplicate** — the existing `TestPutBuffer_RejectsOversizedBuffers` already grows via `Grow`; the guard checks `Cap()` regardless. | `cose_test.go`, `streaming_test.go`, `observability_test.go`, `codec_test.go`; both modes PASS |
| T9 | **docs polish ×5** — AGENTS.md High-Value References rows (`codec.go` + tripwire script); `testdata/fuzz/README.md` `$GOCACHE/fuzz` cache-location section; CONTRIBUTING §Fuzzing (workflow, dual CI lists, crasher protocol); `.go-version` (1.26.5, matches go.mod); README ASCII text-summary fallback under the mermaid diagram. | file diffs |
| T10 | **ANNOTATE 17-29 + 18-09 reports** — every numbered item in §c/§e/§f/§g resolved inline with per-line (renderer-safe) strikethrough + verdicts (`done at <hash>` / `Won't implement` / `NOT-DO` / date-verdict for superb-session work); open items left untouched (17-29: signing wire, atomics, auto-commit drift; 18-09: sibling items, bench-regression CI, README badge, release). 108/108 verdicts applied under structural assertions. | both files; spot-checks in session log |

### Verification commands (all green this session)

```bash
$ go build ./... && GOEXPERIMENT=jsonv2 go build ./...
$ go test -race -count=1 ./...                      # ok (v1)
$ GOEXPERIMENT=jsonv2 go test -race -count=1 ./...  # ok (v2)
$ golangci-lint run ./...                           # 0 issues
$ GOEXPERIMENT=jsonv2 golangci-lint run --build-tags goexperiment.jsonv2 ./...  # 0 issues
$ bash scripts/check-features-planned.sh            # PASS
$ go test -fuzz FuzzStreamingJSON_NDJSONRoundtrip -fuzztime=10s   # PASS (1.2M execs, v1+v2)
$ go test -fuzz FuzzObservableCodec_HookSafety -fuzztime=10s      # PASS (v1+v2)
$ python3 yaml.safe_load(ci.yml)                    # OK
$ go tool cover -func                               # 86.3% (v1) / 88.8% (v2)
```

---

## b) PARTIALLY DONE

| Item | Done | Missing |
| ---- | ---- | ------- |
| T11 living-docs sync | Unreleased-section structure located; all session evidence gathered | ~~`CHANGELOG.md` entries for T1–T10 not yet written; `FEATURES.md` coverage figures stale (85.3/85.4 → 86.3/88.8) and new-evidence rows missing; `TODO_LIST.md` unchanged (still correct: 1 blocked item)~~ done 2026-08-14: CHANGELOG entries added; FEATURES.md at 88.0/88.8 with evidence rows; TODO_LIST re-confirmed unchanged |
| Final gates | test+lint green in both modes after every task | ~~**`nix flake check` not run this session** (treefmt over new files/scripts/HTML unverified; the canonical gate per AGENTS.md)~~ done 2026-08-14: `nix flake check` — all checks passed (treefmt runs gofumpt/goimports/nixfmt only; md/sh/HTML are not format-checked) |

---

## c) NOT STARTED

1. **Release `v0.1.1`** — unchanged: blocked on user decision (`TODO_LIST.md` #1).
2. **Sibling `signing` module `DeterministicCodec` integration** — outside this
   repo (ROADMAP theme 5); left open in both annotated reports.
3. **CI bench-regression job, longer fuzz cron, README badges, SizeResult JSON
   tags** — deliberately deferred (budget/API decisions), documented as
   NOT-DO/open verdicts in the annotations.

---

## d) TOTALLY FUCKED UP

1. **18-09 annotation applied to the wrong item numbers — caught in
   spot-check.** I built the 18-09 §f verdict map from the 20:07 *follow-up*
   report's numbering; the two lists order items differently (e.g. negative
   transcode tests are 18-09 #3 but follow-up #4). The script reported
   "44/44 applied" — counts matched because wrong verdicts landed on wrong
   items. Caught only because my spot-check grep for item 3's expected struck
   text returned nothing (exit 1). Recovered with `git restore` of that one
   file (my own uncommitted annotation — the sanctioned verb) and re-annotated
   from the file's true list (49 verdicts, verified item-by-item).
   **Lesson:** a count assertion cannot catch a numbering shift; the FIRST
   spot-check must compare item TEXT, not just markers.
2. **First fuzz target was contract-too-strict, not the library wrong.**
   `FuzzStreamingJSON_NDJSONRoundtrip` initially asserted any written value
   decodes into `any` — `1e700` falsified it (valid JSON, float64-overflow on
   decode). Correct fix: framing assertions decode into `RawJSONValue`
   (content-agnostic); the finding stays as a committed seed. Classifying a
   harness bug as a library bug would have meant "fixing" correct stdlib
   behavior.
3. **Two routine regressions, both caught by the gates and fixed same-task:**
   a `nlreturn` lint hit in the appended COSE test (heredoc formatting), and
   the hook-tag test's raw leg passed `nil` instead of `[]byte` to `RawCodec`
   (encode error). Also: three benchmark files were written with `range b.N`
   before I noticed the repo is 100% `b.Loop()` — sed'd to convention.
4. **Honest process gap:** T11 (docs sync) was started last and is the least
   finished part; the plan put it last deliberately, but an interruption at
   exactly that point leaves the session's own CHANGELOG record missing — the
   same "ship feature, forget FEATURES/CHANGELOG" pattern the 20:07
   self-review §d-1 criticized. The remainder is ~15 minutes of work.

---

## e) WHAT WE SHOULD IMPROVE

1. **Verify numbering before bulk annotate** — diff the map's first/last item
   text against the target list before running (cheap, catches shifts).
2. **Fuzz target contracts should assert the PROPERTY under test, not
   incidental semantics** — framing counts + error surface, not "decodes into
   `any`".
3. **Run `nix flake check` mid-session after file-tree changes** (new scripts,
   dotfiles, HTML), not only at the end — treefmt surprises are cheaper to fix
   early.
4. **CHANGELOG entries per task (or per tier), not per session** — the daemon
   may scoop the tree at any time; prose owed to CHANGELOG should land with
   the code it describes.

---

## f) Up to 50 things we should get done next

| # | Task | Impact | Effort |
| - | ---- | ------ | ------ |
| 1 | ~~Finish T11: `CHANGELOG [Unreleased]` entries for T1–T10~~ done 2026-08-14 (follow-up session) | High | 15min |
| 2 | ~~Finish T11: `FEATURES.md` — coverage 86.3/88.8 + evidence rows (matrix tests, fuzz ×2, tripwire, benchmarks)~~ done 2026-08-14 — final coverage 88.0/88.8 | High | 15min |
| 3 | ~~Run `nix flake check` (treefmt over new files; canonical gate)~~ done 2026-08-14 — all checks passed | High | 10min |
| 4 | User decides release: cut `v0.1.1` from HEAD + `gh release create` | Critical | 5min |
| 5 | Cross-repo: sibling `signing` module accepts `DeterministicCodec` | High | M |
| 6 | Cross-repo: sibling modules reuse `CBOREncMode()`/`CBORDecMode()` | Med | M |
| 7 | README badge (CI status + pkg.go.dev) | Low | 15min |
| 8 | CI bench-regression job (baseline + benchstat; ROADMAP theme 2) | Med | M |
| 9 | Longer/second fuzz cron after monitoring weekly 30s runs | Low | S |
| 10 | `SizeResult` JSON tags (deferred API change; pair with release decision) | Low | S |
| 11 | Record benchmark baselines (ns/op, B/op) in a docs file for regression eyeballing | Low | 15min |
| 12 | Consider `nix run .#bench` app for on-demand baselines | Low | M |
| 13 | Monitor first GitHub Actions run of the new CI steps (coverage, tripwire, fuzz matrix additions) | Med | S |
| 14 | gopls project diagnostics re-check after daemon commit (expect known stdversion + build-tag warnings only) | Low | S |
| 15 | Migrate any remaining multi-line `~~` spans in older annotated reports to per-line form (18-24 §e-2 follow-through) | Low | 30min |

(15 items — everything else from the 20:07 lists is either done in this
session, resolved NOT-DO with reasons in the annotations, or routed above.)

---

## g) Questions I cannot figure out myself

1. **Release:** cut `v0.1.1` from HEAD (recommendation unchanged — moving the
   published `v0.1.0` tag poisons the module proxy)? Also gates CHANGELOG
   dating and the GitHub Release body.
2. **Commit granularity for the working tree:** T7–T10 (benchmarks + tests +
   docs + annotations, 16 paths) is one uncommitted pile; the daemon may scoop
   it into one commit. Explicit split (tests / docs / annotations) or let it
   ride?
3. **Fuzz CI budget:** the matrix now runs 13 targets × 30s × 2 modes on the
   weekly cron. Is that runner-time budget acceptable, or should some targets
   get shorter budgets?

---

## Files touched this session

```
d9b30ff (daemon):  deterministic_codec{,_v1,_v2}_test.go
                    streaming_fuzz_test.go, observability_fuzz_test.go
                    scripts/check-features-planned.sh
                    doc.go, .github/workflows/ci.yml (tripwire + fuzz matrix)
working tree:       autodetect/envelope/size_benchmark_test.go (new)
                    benchmark_test.go (canon-vs-compact decode delta)
                    codec_test.go (TestDiagnose_InvalidCBOR)
                    cose_test.go (NormalizeCOSEAlgorithm table)
                    streaming_test.go (writer-error, truncated-input)
                    observability_test.go (hook tag × ForEncoding)
                    example_test.go (5 new examples)   [in d9b30ff? see git show]
                    AGENTS.md, CONTRIBUTING.md, README.md, testdata/fuzz/README.md
                    .go-version (new)
                    docs/status/17-29 + 18-09 (annotated)
                    docs/planning/2026-08-14_20-16-SUPERB-quality-plan.{html,d2,svg}
                    testdata/fuzz/FuzzStreamingJSON_NDJSONRoundtrip/seed-number-float64-overflow
```

---

## Closing

10/11 plan tasks shipped and gate-verified; coverage +1.0/+3.4 points; one
genuine fuzz finding converted into a permanent regression seed; both evening
reports fully annotated (108 verdicts); the FEATURES-drift and
DeterministicCodec contract classes are now CI-enforced. Outstanding: a
15-minute docs sync, one `nix flake check`, and the standing user decisions
(release, commit split, fuzz budget). Stopped here per instruction — awaiting
direction.
