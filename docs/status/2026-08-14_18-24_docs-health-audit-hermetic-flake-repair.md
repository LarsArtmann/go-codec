# Status Report: Docs-Health Full Audit + Hermetic Flake Repair

> 2026-08-14 18:24 CEST — full docs-health AUDIT (BUILD + HARVEST + VERIFY +
> ANNOTATE) across all living and historical docs, plus an unplanned critical
> fix: `nix flake check` / `nix build` were broken and are now hermetic and
> green.

---

## Executive Summary

Executed the docs-health skill end to end with all references loaded. All 7
living docs verified against code and repaired where drifting; TODO_LIST was
structurally rotten (trophy section + unharvested 30-item report list) and was
rebuilt to 22 evidence-cited items; ~350 numbered items across 6 historical
documents were resolved inline and ~90 stale `open — TODO_LIST #N` routing
markers in 4 older documents were corrected with `done at <hash>` verdicts. The
audit's biggest non-docs finding: **`nix flake check` was red**
(`/homeless-shelter` sandbox failure) and **plain `nix build` had no default
package** — both fixed by converting the flake to hermetic `buildGoModule`
checks. All gates green at session end: `go test -race` (both JSON modes),
`golangci-lint` 0 issues (both modes), `nix flake check` all checks passed.

Found-state health: **Accuracy 6.25/10, Fitness 4.94/10** (visible math:
10 − 2·1 Critical − 3·0.5 Medium − 1·0.25; 10 − 6·0.75 structural findings −
0.56 structural-ratio penalty for the ~53% non-job TODO_LIST). Baseline for
comparison: 2026-08-12 09:58 audit (9.25/10) — the decay came from the 08-14
session's unharvested report and accumulated unresolved historical items. All
findings were fixed in this session; post-fix state re-verified by gates and
cross-file checks.

---

## a) FULLY DONE

| #  | What                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | Evidence                                              |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| 1  | Loaded docs-health SKILL.md + all 8 references (doc-ownership, harvest-guide, build-guide, agents-quality-guide, verify-checklist, resolving-items, annotation-placement, health-report-format) before touching anything                                                                                                                                                                                                                                                                                                                                                                                                                           | Session log                                           |
| 2  | Full inventory + read: 7 living docs + 8 status reports + 3 planning docs + CI workflow + flake.nix + targeted source greps (~60 verification greps total)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Session log                                           |
| 3  | Claim verification vs code: exposed `DefaultCodec()` ghost in README (no such symbol), missing `ExampleEncodePooled`/`ExampleSize`/`BenchmarkObserveCodec` (grep → 0), fuzz corpus contents, CI matrix structure, flake apps, `go 1.26.5`, `PrepareCOSESetup` line ref, test-constant names, lint/coverage numbers                                                                                                                                                                                                                                                                                                                                 | Session greps                                         |
| 4  | **Quality gates measured, all green:** `go build` + `go test -race` v1+v2; `golangci-lint` 0/0 both modes; coverage re-measured 85.3% (v1) / 85.4% (v2) statements — FEATURES.md numbers confirmed current                                                                                                                                                                                                                                                                                                                                                                                                                                         | Test/lint output                                      |
| 5  | **`nix flake check` repaired (Critical, unplanned):** old `checks` used `runCommand` + `GOCACHE`/`HOME` assumptions → `mkdir /homeless-shelter: permission denied`. Rewrote as hermetic `buildGoModule` with FOD-fetched modules; real `vendorHash` obtained via fixed-output derivation error iteration (`sha256-+JW5…`)                                                                                                                                                                                                                                                                                                                          | `flake.nix`; `nix flake check` → "all checks passed!" |
| 6  | **`nix build` repaired:** flake had no `packages.default` (a known multi-session-old gap). Added via `goModule.overrideAttrs { doCheck = false; }`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | `nix build` now resolves                              |
| 7  | **TODO_LIST.md rebuilt** (1 active item + 20-line trophy section → 22 verified items): trophy deleted, 08-14 report's 30 next-tasks harvested/routed/dropped with code verification, every item cites `file:line` or report evidence, BLOCKED items carry the blocking reason                                                                                                                                                                                                                                                                                                                                                                      | `TODO_LIST.md`                                        |
| 8  | **ROADMAP.md rebuilt:** "Completed:" lists deleted (they were status markers, forbidden in ROADMAP), remaining raw ideas kept + enriched (benchmark regression detection, lazy normalization, codec-cli, cross-repo shim retirement ref), new explicit non-goal (JOSE envelope) citing the 08-14 architecture review                                                                                                                                                                                                                                                                                                                               | `ROADMAP.md`                                          |
| 9  | **CHANGELOG.md `[Unreleased]` gap-filled:** CI hardening entry (SHA-pinned actions, explicit golangci-lint v2.12.2, govulncheck job), hermetic-flake fix entry, architecture-review doc entry, duplicate split Added block merged                                                                                                                                                                                                                                                                                                                                                                                                                  | `CHANGELOG.md`                                        |
| 10 | **FEATURES.md fixed:** `encode/deode` typo, duplicate `## Performance`/`## Performance benchmarks` sections merged (one table), `BenchmarkTranscodeToJSON_*` row added, **`DeterministicCodec` added as ⚪ PLANNED** (verified no code exists), missing blank line                                                                                                                                                                                                                                                                                                                                                                                 | `FEATURES.md`                                         |
| 11 | **README.md:** ghost row "Stack `DefaultCodec()` returns CBOR" replaced (symbol does not exist in this package)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | `README.md`                                           |
| 12 | **AGENTS.md refreshed:** `nix build`/`flake check` descriptions corrected to match the new flake; temporal pollution removed (dated "verified 2026-08-14", "committed at HEAD" phrasing → current-tense constraint); stale `vendorHash` gotcha restored (it had been dropped!); v2 NDJSON `jsontext.Encoder`-corruption gotcha added                                                                                                                                                                                                                                                                                                               | `AGENTS.md` (9.9 KB, in budget)                       |
| 13 | **ANNOTATE — 6 historical docs, every numbered item resolved inline** (strikethrough + `done at <hash>` / `Won't implement` / `NOT-DO` / still-open-with-owner): 09-58 (50 items + TL;DR inline correction), 10-04 execution plan (11 Pareto prose + 21 M-rows + 81 F-rows, incl. F15 Won't-implement-gosec nuance and F57/F60 closed by this session's flake fix), 12-42 perf report (4 inline + routing appendix), 13-55 self-critique (39 verdicts + corrected appendix counts), 20-05 observability report (45 verdicts incl. c)-table status corrections), 08-14 harvest report (30 verdicts + working-tree/commit-intent inline corrections) | `git diff --stat` on docs/                            |
| 14 | **~90 stale routing markers corrected** in the 4 previously-annotated docs: `← open — TODO_LIST #N` markers pointing at the OLD numbering (items since shipped at `094de50`/`ef1f4f4`/`eba9f80`/`d871122`) now read `~~routed, since executed~~ done at <hash>`; genuinely-open ones re-pointed to the NEW TODO_LIST/ROADMAP locations                                                                                                                                                                                                                                                                                                             | grep: 0 old-form markers remain in those files        |
| 15 | Cross-file consistency + link check: every internal file reference in living docs resolves (3 flagged strings are intentional: a TODO-filename, a sibling-repo path, a `*_test.go` glob); no feature simultaneously PLANNED+DONE; no completed item in both TODO_LIST and CHANGELOG                                                                                                                                                                                                                                                                                                                                                                | Link-check script                                     |
| 16 | Health report printed inline with both scores and visible math (per skill: not written to a file)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Previous message                                      |
| 17 | **Archives: 0 new — deliberately.** The skill's archive bar is "EVERY item resolved." Every historical file retains legitimately-open items (release decision, ROADMAP-routed ideas, unowned nice-to-haves). Archiving would have lied; annotation was the correct terminal state. The one fully-executed plan was already archived on 08-12                                                                                                                                                                                                                                                                                                       | `docs/planning/archived/` unchanged                   |

### Verification commands (all green, session end)

```bash
$ go test -race -count=1 ./...                      # ok 1.097s
$ GOEXPERIMENT=jsonv2 go test -race -count=1 ./...  # ok 1.124s
$ golangci-lint run ./...                           # 0 issues
$ golangci-lint run --build-tags goexperiment.jsonv2 ./...  # 0 issues
$ nix flake check                                   # all checks passed!
$ go tool cover -func (both modes)                  # 85.3% / 85.4% statements
```

---

## b) PARTIALLY DONE

| Item                    | Done                                                                         | Missing                                                                                                                                                                                                                                          |
| ----------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Historical annotation   | Every NUMBERED item in 6 docs has a verdict; 4 older docs' markers corrected | A handful of deliberately-unmarked prose lines (niche nice-to-haves) are listed in per-file appendices rather than struck — documented, but a strict reader must read the appendix to know they were considered                                  |
| Strikethrough rendering | Applied consistently with the repo's existing annotation style               | Several `~~…~~` spans cross line breaks in prose lists; GitHub's renderer may not strike across newlines. Matches repo precedent (all prior annotations have the same shape), but I never verified in a real renderer                            |
| flake.nix modernization | Hermetic checks + `packages.default` + working vendorHash                    | `checks.build` duplicates the `packages.default` build (harmless redundancy); `version = "unstable"` placeholder instead of deriving from a tag; `src` uses `gitTracked ./.` (broad — includes docs; works, arguably fine for a checks-only FOD) |
| TODO_LIST breadth       | 22 items, all verified                                                       | The old reports' very-long-tail micro-ideas (e.g. detail-string hex test, `AutoDetectDebug` benchmark, envelope-wrapped detection test) were consciously left OUT to keep the list actionable — they live only in annotated reports now          |

---

## c) NOT STARTED

1. **No commit made** — per standing rules (user did not request; auto-commit
   daemon owns the working tree). 18 files modified, +638/−548.
2. **No GitHub Release** — blocked on the user's release-strategy decision
   (`TODO_LIST.md` #1), unchanged from prior sessions.
3. **No `.go-version` / dependabot / codec-cli / website work** — routed to
   TODO_LIST/ROADMAP, correctly not started in a docs session.

---

## d) TOTALLY FUCKED UP

### 1. Bulk table-surgery scripts mangled columns — twice

My first 09-58 annotation script used `(.+?) \|` which captured only the FIRST
table cell, destroying the Why/Est columns (turned the 50-row table into
2-column mush). Caught it in the sed spot-check, recovered via `git restore`
(allowed — my own changes), and rewrote with full-column preservation. Then the
08-14 script left dangling `| |` empty cells — needed a third cleanup pass.

**Lesson:** every bulk table edit needs a structural assertion (column count
per row) BEFORE writing the file, not just an eyeball afterward.

### 2. Marker-update regex silently half-matched

The script that corrected `← open — TODO_LIST.md #N` markers used `#[0-9]`
(single digit) instead of `#[0-9]+`. First run "succeeded", updating only
single-digit references (27 of ~90 in one file) while reporting mere
"changed". Caught only because I re-grepped for remaining markers afterward.

**Lesson:** for every bulk replace, print and compare BEFORE/AFTER match
counts. A silent partial apply is worse than a failure.

### 3. Unverified numbers in an appendix

The 13-55 resolution appendix initially claimed "34 closed, 16 open" — I wrote
the counts from vibes, then noticed they didn't match my own verdict table and
had to rewrite the paragraph in a second pass. Numbers in resolution appendices
must be counted from the actual verdicts.

### 4. Sloppy generated code

One annotation script contained `{93 if False else '93e68f3'}` — a leftover of
mid-thought editing that happened to evaluate correctly. Harmless output, but
it's exactly the "stacked guesses" anti-pattern the 08-14 report flagged in a
prior session. Slow down or hand-edit.

---

## e) WHAT WE SHOULD IMPROVE

1. **Bulk-edit hygiene:** structural assertions + match-count deltas for every
   scripted docs edit (the two failure classes in d-1/d-2 both evade "applied
   OK" outputs).
2. **Renderer-aware strikethrough:** keep `~~` spans single-line (or per-line)
   so GitHub actually renders the resolution; migrate the handful of multi-line
   spans when those files are next touched.
3. **Keep the flake honest:** add the vendorHash-update step to memory done
   (AGENTS.md gotcha restored); consider deriving `version` from the tag and
   deduping `checks.build` vs `packages.default` at the next flake touch.
4. **Harvest immediately:** the 08-14 session wrote a 30-item report without
   harvesting it — this audit found it two days later as structural rot. Run
   HARVEST at report-writing time, not "next session".
5. **Verify counts before printing them** — appendix arithmetic included.

---

## f) Up to 50 things we should get done next

> From the rebuilt TODO_LIST (22 items) + ROADMAP raw ideas + this session's
> process notes, ranked by impact.

| #  | Task                                                                                                                                                                                                                                                                                             | Impact   | Effort |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------ |
| 1  | **Decide release strategy + `gh release create`** (v0.1.1 from HEAD recommended; moving the published v0.1.0 tag poisons the module proxy)                                                                                                                                                       | Critical | 5min   |
| 2  | Implement `DeterministicCodec` marker interface (approved proposal, ~15 lines; `CBORCodec`/`CBORCompactCodec` always, `JSONCodec` v2-build only, `RawCodec` never)                                                                                                                               | High     | 30min  |
| 3  | CI fuzz job (cron, short fuzztime) + committed seed corpus for `FuzzAutoDetectDebug_Consistency`                                                                                                                                                                                                 | Medium   | 1-2h   |
| 4  | Convert `normalizeForJSON` depth error to `go-error-family` with stable code (`codec.normalize_depth_exceeded`) — last non-categorized error                                                                                                                                                     | Medium   | 30min  |
| 5  | `BenchmarkObserveCodec` (decorator overhead) — decides the atomics-vs-RWMutex question                                                                                                                                                                                                           | Medium   | 30min  |
| 6  | Observability edge-case test bundle: wrap `CBORCompactCodec`, `EncodeToBuffer` inner-error propagation, `buf.Write` failure fallback, `MetricsSnapshot` immutability, nested `ObservableCodec` no-double-count, `ObserveCodec(nil)`, `EncodePooled` composition, hook byte counts on error paths | Medium   | 2h     |
| 7  | v2 streaming test with non-buffer reader (`strings.Reader` / byte-at-a-time) — over-read class bit us once                                                                                                                                                                                       | Medium   | 30min  |
| 8  | `ExampleEncodePooled` godoc example                                                                                                                                                                                                                                                              | Low      | 15min  |
| 9  | `ExampleSize` godoc example (`SizeResult`)                                                                                                                                                                                                                                                       | Low      | 15min  |
| 10 | Make `ExampleObserveCodec` output size-independent (drop hardcoded `bytes=12`)                                                                                                                                                                                                                   | Low      | 15min  |
| 11 | Fix v2 `JSONEncoder` per-call `[]byte{'\n'}` allocation                                                                                                                                                                                                                                          | Low      | 15min  |
| 12 | Add `cbor:"3,keyasint"` to `realisticOrderKeyInt.Items` benchmark struct                                                                                                                                                                                                                         | Low      | 5min   |
| 13 | Measure or hedge README/doc.go "19-43% smaller / 25-72% faster" claims with a benchmark citation                                                                                                                                                                                                 | Low      | 30min  |
| 14 | Rename opaque test constants (`testField`, `testFieldE`, `testMapKey`)                                                                                                                                                                                                                           | Low      | 15min  |
| 15 | Annotate/nolint `json_helpers_v2_test.go` gopls `stdversion` warnings                                                                                                                                                                                                                            | Low      | 15min  |
| 16 | Add `dependabot.yml`/`renovate` config                                                                                                                                                                                                                                                           | Low      | 15min  |
| 17 | Prometheus/OpenTelemetry exporter example (blocked: backend choice — see g-2)                                                                                                                                                                                                                    | Low      | 30min  |
| 18 | CI step: `golangci-lint --out-format json` artifact (LSP-vs-CLI truth)                                                                                                                                                                                                                           | Low      | 15min  |
| 19 | README architecture diagram (codec → store/event/signing/encryption)                                                                                                                                                                                                                             | Low      | 20min  |
| 20 | Streaming benchmarks (`BenchmarkStreamingJSON_*` + CBOR; v2 decoder variants)                                                                                                                                                                                                                    | Low      | 45min  |
| 21 | `PutBuffer` size guard (reject >1 MiB buffers from the pool)                                                                                                                                                                                                                                     | Low      | 15min  |
| 22 | README sections: Streaming JSON (NDJSON), `EncodePooled`, `Size`/`SizeResult`                                                                                                                                                                                                                    | Low      | 30min  |
| 23 | Dedupe flake `checks.build` vs `packages.default`; derive `version` from tag                                                                                                                                                                                                                     | Low      | 15min  |
| 24 | Migrate multi-line `~~` strikethrough spans to per-line form (renderer-safe)                                                                                                                                                                                                                     | Low      | 30min  |
| 25 | Commit this session's output (daemon or explicit — see g-3)                                                                                                                                                                                                                                      | High     | 2min   |
| 26 | ROADMAP theme 5: website launch via `website-launch` skill                                                                                                                                                                                                                                       | Med      | 1h+    |
| 27 | ROADMAP theme 5: worked end-to-end example consuming go-codec from go-cqrs-lite                                                                                                                                                                                                                  | Med      | 30min+ |
| 28 | ROADMAP theme 5: pkg.go.dev polish — godoc example per exported symbol                                                                                                                                                                                                                           | Med      | 30min+ |
| 29 | ROADMAP theme 5: cross-repo PR — wire `ObservableCodec` into go-cqrs-lite event store                                                                                                                                                                                                            | High     | M      |
| 30 | ROADMAP theme 5: cross-repo PR — `AutoDetectDebug` in mixed-stream diagnostics                                                                                                                                                                                                                   | Med      | M      |
| 31 | ROADMAP theme 5: retire deprecated `codec/v4` shim (mechanical, start at `event/codec.go`)                                                                                                                                                                                                       | Med      | M      |
| 32 | ROADMAP theme 2: decode-side buffer pool                                                                                                                                                                                                                                                         | Med      | M      |
| 33 | ROADMAP theme 2: CBOR type-cache pre-warm helper                                                                                                                                                                                                                                                 | Low      | M      |
| 34 | ROADMAP theme 2: benchmark regression detection in CI (baseline + benchstat + `.#bench` app)                                                                                                                                                                                                     | Med      | M      |
| 35 | ROADMAP theme 2: lazy normalization (normalize only when v1 marshaler chokes)                                                                                                                                                                                                                    | Med      | L      |
| 36 | ROADMAP theme 4: `LastEncodeTime`/`LastDecodeTime`, size histograms, per-encoding aggregation                                                                                                                                                                                                    | Low      | M      |
| 37 | ROADMAP theme 4: atomics-based `CodecMetrics` (if #5 justifies)                                                                                                                                                                                                                                  | Med      | M      |
| 38 | ROADMAP theme 4: configurable `maxAutoDetectSize`                                                                                                                                                                                                                                                | Low      | M      |
| 39 | ROADMAP theme 1: MessagePack codec                                                                                                                                                                                                                                                               | Med      | 2h     |
| 40 | ROADMAP theme 1: protobuf/flatbuffers adapter                                                                                                                                                                                                                                                    | Low      | 3h     |
| 41 | ROADMAP theme 1: user-registration dispatch table for `ForEncoding`                                                                                                                                                                                                                              | Med      | 45min  |
| 42 | ROADMAP theme 3: wire-shape diff tooling (reorder/renumber/rename detection)                                                                                                                                                                                                                     | Med      | 3h     |
| 43 | ROADMAP theme 3: migration helper (`UnwrapDecode` + codec swap)                                                                                                                                                                                                                                  | Med      | 2h     |
| 44 | ROADMAP theme 3: `toarray` field-order lint                                                                                                                                                                                                                                                      | Med      | 1h     |
| 45 | ROADMAP theme 5: `codec-cli` diagnostic tool (CBOR dump → JSON/EDN)                                                                                                                                                                                                                              | Low      | M      |
| 46 | Depth-cap integration test through full `TranscodeToJSON` path                                                                                                                                                                                                                                   | Med      | 20min  |
| 47 | Full-chain integration test: encode → envelope → detect → decode                                                                                                                                                                                                                                 | Med      | 30min  |
| 48 | `.go-version` file for gvm/asdf users                                                                                                                                                                                                                                                            | Low      | 5min   |
| 49 | Go 1.27 bump when released (drop dual-build)                                                                                                                                                                                                                                                     | Low      | 30min  |
| 50 | API-stability guarantee doc (SemVer commitments for the Codec seam)                                                                                                                                                                                                                              | Low      | 30min  |

---

## g) Questions I cannot figure out myself

### 1. Release strategy — v0.1.1 from HEAD, or move the v0.1.0 tag?

The published `v0.1.0` tag sits at `3f8ac9d`, which predates the v1-build
corruption AND its fix (`d871122`), the hermetic flake, and all `[Unreleased]`
additions. My strong recommendation: **cut `v0.1.1` (or `v0.2.0` given the
`SizeResult` breaking change) from HEAD and leave `v0.1.0` untouched** —
moving a published tag poisons the module proxy checksums. This blocks
TODO_LIST #1, CHANGELOG dating, and the GitHub Release.

### 2. Exporter example backend (TODO_LIST #17)?

For the Prometheus/OpenTelemetry hook example: a dependency-free
pseudo-metrics example (keeps go-codec dep-light) or a real
`prometheus/client_golang` example accepting the dev-dependency? This decides
whether the example lives in `example_test.go` or a separate doc.

### 3. Commit granularity for this session's output?

18 files modified (+638/−548): the hermetic flake fix is a behavioral change;
the rest is docs. Should the auto-commit daemon take it all as-is, or do you
want an explicit split (`fix(nix): hermetic checks` vs `docs: full docs-health
audit`)? I won't commit without your call.

---

## Files touched this session

```
AGENTS.md                | de-temporalized, 2 gotchas restored/added, commands fixed
CHANGELOG.md             | [Unreleased] gap-filled (CI, flake, arch review)
FEATURES.md              | typo, merged perf sections, DeterministicCodec PLANNED
README.md                | ghost DefaultCodec row removed
ROADMAP.md               | rebuilt (completed lists pruned, themes refreshed)
TODO_LIST.md             | rebuilt (22 items, trophy deleted)
flake.nix                | hermetic buildGoModule checks + packages.default
docs/status/*.md (7)     | annotated/corrected (09-58, 12-42, 13-55, 20-05, 08-14 heavy;
                         |  23-38, 03-24, 09-25 marker corrections)
docs/planning/*.md (2)   | 10-04 fully resolved inline; 09-07 markers corrected
```

---

> **Bottom line:** docs are now internally consistent and every historical
> claim is resolved; the repo's canonical verification path (`nix flake check`)
> went from red to green as a side effect of the audit. The only true blockers
> remaining are user decisions (release, exporter, commit split).
