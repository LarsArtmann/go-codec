# Status Report: T11 Completion + Brutal Self-Review (superb-plan session closed)

> Session: 2026-08-14 ~20:48–20:58. Scope: resumed T11 of the 11-task
> SUPERB-quality plan (`docs/planning/2026-08-14_20-16-SUPERB-quality-plan.html`),
> completed it, then ran this self-review. Based on this session's run and what
> it noticed — no new research beyond it.

## TL;DR

T11 done: CHANGELOG entries, FEATURES.md sync (coverage + evidence + measured
benchmark rows), first `nix flake check` of the session-pair (all green), and
inline annotation of the 20-47 report. The superb plan is now **T1–T11 all
complete**. The auto-commit daemon then scooped the entire 18-path tree into
one commit (`f04d158`, 20:55:53), preempting the user's commit-granularity
question. One self-caught honesty bug (unverified causal claim in an
annotation) was fixed on the spot. Nothing is red; the project is waiting on
three user decisions, one of which just went moot.

---

## a) FULLY DONE (this session)

1. **Todo-state correction** — handoff list was stale; set T1–T10 `completed`,
   T11 `in_progress`, closed T11 at the end. The plan is 11/11.
2. **Fresh coverage measurement, both modes** — **88.0% (v1) / 88.8% (v2)**.
   Caught that the handoff's claimed 86.3% (v1) does not reproduce on the final
   tree. Fresh numbers were used everywhere; FEATURES.md records 88.0/88.8.
3. **CHANGELOG.md `### Added`** — 10 entries covering T1–T10 (matrix tests,
   NDJSON fuzz + `1e700` seed, hook-safety fuzz, drift tripwire, doc.go
   guidance, CI coverage/fuzz additions, 6 benchmark functions, edge/error-path
   tests, 5 godoc examples, docs batch). Every symbol and file citation was
   verified against the tree before writing (grep of function names, daemon
   diff of `d9b30ff`, corpus directory listing).
4. **FEATURES.md** — coverage figures updated; CI drift-guard note added to the
   header; 4 new Performance rows with **actually measured** numbers
   (AutoDetect ~71–104ns/1-alloc first-byte vs ~174ns/4-allocs trial-decode;
   Wrap ~287ns / Unwrap ~840ns; Size ~239ns; Compact-vs-Canon decode parity
   ~320ns); evidence citations appended to 3 feature rows (DeterministicCodec
   matrix, NDJSON fuzz incl. `1e700` seed, hook-safety fuzz).
5. **Drift tripwire re-verified** — `scripts/check-features-planned.sh` PASS
   after my own FEATURES edits (guards against annotating myself into drift).
6. **`nix flake check` — ALL CHECKS PASSED** — the canonical gate that had
   never been run this session-pair (treefmt + hermetic build + dual-mode
   tests). Scope discovery: treefmt runs **gofumpt/goimports/nixfmt only** —
   markdown, shell, and HTML are NOT format-checked, so the handoff's worry
   about the shell script and HTML plan was moot.
7. **20-47 report annotated inline** — title, TL;DR coverage correction, T11
   and final-gates table rows, f-items 1–3 all marked done with evidence; the
   commit-granularity question marked moot (daemon scoop).
8. **Honesty amendment** — removed my own unverified causal claim from an
   annotation (see d-1); annotation now reads "cause unverified".
9. All of the above is preserved in daemon commit `f04d158` (except the last
   two annotation touches to the 20-47 report, which sit in the working tree
   for the daemon to pick up).

## b) PARTIALLY DONE

1. **Superb plan overall** — T1–T11 code and docs are complete and committed,
   but the plan's release-gated follow-through is open: `v0.1.1` not cut,
   `[Unreleased]` undated, `SizeResult` JSON tags deferred pending release.
2. **FEATURES.md benchmark numbers are indicative, not baseline-grade** —
   measured with `-benchtime=200ms` quick runs (fine for orders of magnitude
   and allocation counts; not a benchstat baseline for regression detection).
3. **New CI steps unexercised by a real runner** — coverage summary, FEATURES
   drift tripwire, and the two new fuzz-matrix targets have never run on
   GitHub Actions (nothing pushed since they landed). YAML is valid; runner
   behavior is unverified.
4. **20-47 report's open f-items 4–15 remain open** (release, sibling
   adoption, badges, bench-regression job, baselines, …) — routed there, not
   re-listed blindly here (see f).

## c) NOT STARTED (deliberately, user-gated or cross-repo)

1. `v0.1.1` release (tag + GitHub Release + CHANGELOG dating) — waiting on user.
2. Sibling `signing` module accepting `DeterministicCodec` (cross-repo).
3. Sibling modules reusing `CBOREncMode()`/`CBORDecMode()` (cross-repo).
4. README badges (CI status, pkg.go.dev).
5. CI bench-regression job + benchmark baseline document.
6. Real-runner verification of the new CI steps (needs a push).

## d) TOTALLY FUCKED UP (brutally honest)

1. **(Mine — caught and fixed during this review.)** I wrote an *unverified
   causal claim* into a historical annotation: "86.3 was measured before the
   T7–T10 test files landed." That story is not just unverified — it cannot be
   right, because the 20-47 report was written *after* T7–T10 landed and still
   claims 86.3. I pattern-matched to a plausible explanation instead of
   stopping at the verified fact (my fresh measurement = 88.0). Fixed by
   amending the annotation to "cause unverified, do not trust the original v1
   number." Root cause: wanting the annotation to *explain* rather than just
   *resolve*. Rule going forward: **verdicts + citations only; no causal
   narratives without evidence.**
2. **(Inherited, prior session, already documented in the 20-47 report.)** The
   18-09 report annotation initially used the wrong numbering (copied list
   order from a different report); recovered via `git restore` and re-annotated
   from the true list. Lesson on record: first spot-check must compare item
   TEXT, not marker counts. Same failure *class* as d-1: trusting a plausible
   story over verified content.
3. **(Process/infrastructure, not mine to fix.)** The auto-commit daemon
   scooped the entire 18-path tree — code, tests, docs, annotations, and my
   just-written T11 entries — into a single commit `f04d158` at 20:55:53,
   ~8 minutes after the user was asked how to split it. No content was lost
   and all gates were green before the commit, but the user's decision was
   preempted by infrastructure. Also: the daemon's commit credits
   "Assisted-by: Crush:MiniMax-M3" — not the model that did this work.

## e) WHAT WE SHOULD IMPROVE

1. **Annotation discipline** — verdicts + evidence citations only, never
   explanatory narratives without proof. Two incidents in one session-pair
   (d-1, d-2) is a pattern, not bad luck.
2. **Measure at write-time** — coverage (and any performance figure quoted in
   a report) should be measured when the report/doc entry is written, not
   carried forward from mid-session (the 86.3 incident).
3. **Benchmark provenance** — FEATURES performance numbers should eventually
   come from a benchstat long-run baseline; until then they are indicative
   (marked `~`), which is honest but weaker.
4. **Dead test code smells visible in diagnostics and ignored because tests
   are green** — `observability_test.go:560` (gopls unusedwrite: `EncodeBytes`
   is written but never read — the test records a field it never asserts) and
   `testdata_test.go:7` (an entire constant set unused). Small, but shipped in
   a "superb" session; somebody should look.
5. **Go-version split brain** — `.go-version` (new, `1.26.5`) vs `go.mod` vs
   CI `go-version-file: go.mod`. Three places, one consumer (CI reads go.mod).
   Pick one source of truth or wire `.go-version` in.
6. **Daemon behavior vs open questions** — asking the user about commit
   granularity while the daemon runs is asking a question with an 8-minute
   fuse. Either pause/serialize the daemon around pending decisions, or accept
   that the daemon decides.

## f) Things we should get done next

| # | Task | Impact | Effort |
| - | ---- | ------ | ------ |
| 1 | USER: decide `v0.1.1` release from `f04d158` (tag + GitHub Release + date the `[Unreleased]` section) | Critical | 5min |
| 2 | USER: fuzz CI budget — 13 targets × 30s × 2 modes weekly cron, acceptable or trim? | Med | — |
| 3 | USER: daemon commit policy — is one-commit scooping fine, or should docs/reports be committed separately from code? (the preempted question) | Med | — |
| 4 | Push and watch the first real CI run: coverage step, FEATURES tripwire, new fuzz-matrix entries | High | S |
| 5 | Sibling `signing` module: accept `DeterministicCodec` (compile-time gate for non-deterministic codecs) | High | M |
| 6 | Sibling modules: reuse `CBOREncMode()`/`CBORDecMode()` instead of rebuilding modes | Med | M |
| 7 | Benchstat long-run baseline → record in a docs file; upgrade FEATURES `~` numbers to baseline-grade | Med | 30min |
| 8 | Fix `observability_test.go:560` unusedwrite: assert `EncodeBytes` or drop the write | Low | 10min |
| 9 | Remove or use the unused constant set at `testdata_test.go:7` | Low | 10min |
| 10 | Single source of Go version: point CI at `.go-version` or delete `.go-version` | Low | 10min |
| 11 | CI bench-regression job (baseline + benchstat; ROADMAP theme 2) | Med | M |
| 12 | README badges (CI status + pkg.go.dev) | Low | 15min |
| 13 | `SizeResult` JSON tags — API change, pair with the next release decision | Low | S |
| 14 | Longer/second fuzz cron once weekly runs are monitored | Low | S |
| 15 | pkg.go.dev verification after the release (`go get` check against the proxy) | Med | 10min |
| 16 | Migrate remaining multi-line `~~` spans in older annotated reports to per-line form | Low | 30min |

(16 items — deduped against the 20-47 report's still-open f-items 4–15; those
remain the routing source. Per the status-report skill, this list is HARVEST
fuel for the next docs-health run, not a commitment list.)

## g) Questions I cannot figure out myself

1. **Release:** cut `v0.1.1` from `f04d158` now? Recommendation unchanged —
   yes, new tag, never move the published `v0.1.0`. Also gates CHANGELOG
   dating and the GitHub Release body.
2. **Fuzz CI budget:** 13 targets × 30s × 2 modes on the weekly cron — is that
   runner-time budget acceptable, or should some targets get shorter budgets?
3. **Daemon commit policy:** the daemon preempted your commit-split decision
   by committing everything as `f04d158` within 8 minutes of the question.
   Going forward: one-commit scoops acceptable, or should reports/docs be
   committed separately from code (a daemon-config change only you can make)?

---

*Report format note: written as `.md` per the user's explicit instruction —
the status-report skill's canonical format is styled HTML; this is a
deliberate, flagged override, not a new default.*
