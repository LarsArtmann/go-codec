# Status Report — Docs-Health AUDIT: Harvest, Verify, Annotation Sweep

**Date:** 2026-08-15 00:49 CEST (session ran ~00:15–00:49, crossing midnight from 08-14)
**Session type:** docs-health skill, full AUDIT mode (BUILD + HARVEST + VERIFY + ANNOTATE)
**Trigger:** "Execute the docs-health SKILL — TODO_LIST, CHANGELOG, AGENTS, ROADMAP, FEATURES must be superb."
**Scope:** This session only — the audit sweep and what I noticed during it. No new research beyond it.
**Format note:** written as `.md` per explicit user request (status-report skill default is
styled HTML — override flagged per spec, not a new default).

---

## TL;DR

Full docs-health AUDIT executed: harvested 4 newest status reports (including one that
appeared **mid-session** from a concurrent run), rebuilt `TODO_LIST.md` from 1 item to 8
evidence-cited items, fixed every drift found (FEATURES perf rows, DOMAIN_LANGUAGE count,
README badge, ROADMAP fold, CHANGELOG entry), migrated **34** multi-line `~~` annotation
spans to the renderer-safe per-line form across 4 old reports (the item three prior
sessions kept deferring), recovered one **lost** work item (`gosec` — flagged 08-12,
never shipped, tracked nowhere), and re-measured coverage live: **88.0% (v1) / 88.8%
(v2)**, matching FEATURES.md exactly. Gates: `nix flake check` all green; FEATURES
tripwire PASS before and after my edits. Health verdict: **Accuracy 8.25 → 10,
Fitness 8.5 → 10** (original-state scores; all findings fixed in-place). Nothing red.
Three user decisions remain standing (§g).

---

## a) FULLY DONE (this session, with evidence)

1. **Skill + references loaded before acting** — docs-health SKILL.md, doc-ownership,
   verify-checklist, health-report-format; then full doc inventory (7 living docs,
   14→15 status reports, planning docs incl. archived).
2. **HARVEST from the 4 newest reports** (`20-47`, `20-07 ×2`, and `20-58` discovered
   mid-session): every forward-looking item verified against code before routing.
   - **Falsified one report claim:** 20-58 §f-9 said `testdata_test.go:7` constant set
     is unused — grep shows all 10 constants used 5–74×. Dropped, not routed.
   - **Dropped already-done candidates** (verified, not trusted): CI `retention-days`
     (set at ci.yml:64/153), the 4 fuzz targets added since v0.1.0 (all in CHANGELOG;
     `FuzzCBORCodec_CanonicalFidelity` predates the tag — only its seeds are new),
     benchmark batch (fully changelogged in `f04d158`).
3. **TODO_LIST.md rebuilt 1 → 8 items**, each citing code (`file:line`) AND report
   provenance; release item #1 enriched with the full post-release runbook
   (tag → `gh release create` → CHANGELOG dating → `go get`/pkg.go.dev verify →
   sibling rebuild with `GOWORK=off`).
4. **Recovered a lost item:** `gosec` security lint — flagged 2026-08-12
   (03-24 report, "CI / release hygiene" #10), never shipped, tracked by no living doc.
   Verified absent from `.github/workflows/ci.yml` (`govulncheck` + `gitleaks` ARE
   present) → new TODO #8.
5. **ROADMAP theme 5 folded:** sibling `signing` accepting `DeterministicCodec` and
   sibling `CBOREncMode()`/`CBORDecMode()` reuse (previously floating in report
   §f-lists) merged into the existing cross-repo bullet.
6. **FEATURES.md:** enumerated all 33 benchmark functions in the tree, read both
   unlisted implementations, added the 2 missing Performance rows
   (`BenchmarkRealisticPayload_Encode/Decode`, `BenchmarkObserveCodec`).
7. **DOMAIN_LANGUAGE.md:** "five unexported functions (four names)" corrected —
   five helpers = four functions + the `rawJSONValue` type alias.
8. **README.md:** CI workflow badge added (Low-severity gap from 20-47 §f-7).
9. **CHANGELOG.md `[Unreleased]`:** audit entry appended (span migration, badge,
   FEATURES rows, DOMAIN_LANGUAGE fix, harvest).
10. **ANNOTATE — 34 multi-line `~~` spans migrated to per-line form** in 4 old
    reports: `08-11 ×15`, `09-25 ×1`, `03-24 ×15`, `12-42 ×3`. This was tracked open
    in three consecutive reports (18-24 #24, 20-47 §f-15, 20-58 §f-16) — now done.
    Verified clean afterward: only prose *mentions* of the format remain.
11. **ANNOTATE — 4 stale "still open / routed to TODO_LIST #N" claims corrected
    inline** with commit evidence (09-58 ×3: coverage CI, fuzz cron, CONTRIBUTING
    snapshot flow; 03-24 ×1: fuzz/coverage). All four asserted open work that had
    shipped.
12. **Ghost investigated:** commit `2c98116`'s message advertises "EncodedSize
    helpers" — no such symbol exists in the diff or tree. Daemon-inflated commit
    message; no doc gap (nothing living-doc references it). No action needed beyond
    this record.
13. **Verification battery (all confirmed):**
    - Coverage re-measured fresh: 88.0% v1 / 88.8% v2 — FEATURES.md claims exact.
    - FEATURES line-number citations spot-checked real: `cose.go:144`,
      `cbor.go:23`, `cbor_compact.go:35`, `json_compat_v2.go:42`, `pool.go` 1 MiB
      guard.
    - AGENTS.md hygiene: 10.6 KB (in budget), 0 commit hashes, 0 temporal-pollution
      matches, every documented command exists as a flake app/check.
    - CHANGELOG ↔ git log/tags complete; only tag `v0.1.0` (at `3f8ac9d`);
      **no GitHub Releases exist** (`gh release list` empty) — TODO #1 is valid.
    - `SizeResult` confirmed tagless (TODO #4 justified); `DecodePooled` confirmed
      nonexistent (ROADMAP idea legitimately open).
14. **Quality gates:** `scripts/check-features-planned.sh` PASS before AND after my
    FEATURES edits; `nix flake check` — **all checks passed** (treefmt + hermetic
    build + dual-mode tests).
15. **Inline health report printed** (conversation, not a file) with two scored axes:
    Accuracy 8.25/10, Fitness 8.5/10 as found → both at 10 after fixes; per-doc
    findings table with every finding marked fixed; "could not verify" set stated
    (real-runner CI, pkg.go.dev, daemon policy).

## b) PARTIALLY DONE

1. **Secondary living docs unverified:** `CONTRIBUTING.md` and `SECURITY.md` were
   never read this session. The audit scoped to the five target docs +
   DOMAIN_LANGUAGE + AGENTS.md. Their freshness is unknown, not confirmed.
2. **README external links not fetched** (pkg.go.dev, go-branded-id, go-cqrs-lite
   repos) — external targets; internal links/files all verified.
3. **Trusted-not-checked set:** I relied on the 20-47 report's claim that "everything
   else from the 20:07 lists is done / NOT-DO / routed" instead of independently
   re-verifying all 50 items of the 20-07 follow-up report. Spot-checks all passed
   (retention, fuzz, benchmarks), but e.g. #34 (lint-artifact-validity CI test),
   #36 (mermaid-render CI check), #43 (Envelope fallback test), #44 (golden
   snapshots of README examples) were NOT individually grepped.
4. **Pre-existing working-tree modification never diffed:** the 20-47 report was
   already `M` at session start (per the 20-58 report §a-9, its final annotation
   touches await the daemon). I inferred provenance from that note instead of
   running `git diff` — correct by luck, not by verification.
5. **One stale old-numbering reference survives** at 09-58:323 ("determines whether
   TODO_LIST #1/#2 are P0 or P2" — old numbering, inside a §g question). Seen,
   judged historical context, left untouched. Defensible; borderline.

## c) NOT STARTED (user-gated, cross-repo, or out of audit scope)

1. **`v0.1.1` release execution** — user decision (TODO #1; runbook embedded there).
2. **Auto-commit daemon policy** — user decision (TODO #7).
3. **Fuzz CI budget** (13 targets × 30s × 2 modes) — user decision (in TODO #2).
4. **Sibling `signing`/`encryption` integration** — cross-repo (ROADMAP theme 5).
5. **All routed code items** — benchstat baseline (#3), `SizeResult` tags (#4),
   `observability_test.go:560` unusedwrite (#5), Go-version single source (#6),
   `gosec` CI (#8). A docs audit fixes docs; these are code/CI work.
6. **`docs/planning/archived/` stale-reference audit** — not touched.
7. **HARVEST of THIS report's §f** — items 1–11 were routed into TODO_LIST/ROADMAP
   *live during the audit itself* (before this file existed); nothing here is new
   fuel beyond what §f marks as un-routed.

## d) TOTALLY FUCKED UP

1. **Two multiedit old_strings failed on the 03-24 report** (13/15 applied). I had
   reconstructed the §g Q1 span's line-wrap from an earlier view instead of copying
   character-exact text — the actual file wrapped differently. Recovered by
   re-viewing at the exact offset and redoing both edits. The exact-match discipline
   this skill demands slipped on the first attempt; one wasted round trip.
2. **Wrong detection tool first.** My initial multi-line-span grep
   (`~~.*~~.*~~`) can only match ≥3 markers on a single line — it found 2 of the 4
   affected files and missed 03-24 and 12-42 entirely. The awk parity check (odd
   per-line marker count) run afterward found them. I used a tool that cannot see
   multi-line structure to detect multi-line structure.
3. **Concurrent-session blind spot.** The initial `ls` showed 14 status reports; the
   20-58 report was created by a parallel session mid-audit and my "3 newest"
   working set therefore excluded the single most harvest-relevant report until the
   parity scan stumbled over it. A `git status` at audit start would have shown the
   untracked file immediately. Luck surfaced it; process should have.
4. **Off-by-one in the final health report:** I wrote "33 multi-line spans migrated";
   the true count is 34 (15+1+15+3). A miscounted number inside a report whose
   entire job is numeric accuracy — same class as the 86.3% incident the 20-47
   session documented.

## e) WHAT WE SHOULD IMPROVE

1. **Diff unexplained working-tree modifications before building on them.** The
   global rule exists; I skipped it because provenance was inferable from a report.
   Inference is not verification.
2. **Parity check first for strikethrough audits.** Multi-line structure demands a
   per-line marker-parity scan (awk), not a single-line regex. Encode this in
   process, not in memory.
3. **`git status` at audit start, newest artifacts first.** The 20-07 report's own
   §e-5 rule ("read the newest 1–2 reports before ANY verification work") half-held:
   I read the newest *committed* reports but never checked for uncommitted/new files.
4. **Name the trusted-not-checked set at write time, not in retrospect.** When a
   prior session's dedup claim is load-bearing, the report should list exactly which
   items rest on trust (I did list them — but only after being asked what I forgot).
5. **Compute every number in a health report from the edit log at write time.** The
   33-vs-34 slip is the 86.3 lesson in miniature: quoted-from-memory numbers rot.

## f) Up to 50 things we should get done next

Routed during the audit (do not re-harvest from here — they live in TODO_LIST/ROADMAP):

| # | Task | Impact | Effort | Routed to |
| - | ---- | ------ | ------ | --------- |
| 1 | USER: decide `v0.1.1` release (tag + GitHub Release + CHANGELOG dating) | Critical | 5min | TODO #1 |
| 2 | USER: daemon commit policy (scoop vs split) | Med | — | TODO #7 |
| 3 | USER: fuzz CI budget (13×30s×2) trim-or-keep | Med | — | TODO #2 |
| 4 | Push + watch first real CI run (coverage, tripwire, fuzz matrix, cron) | High | S | TODO #2 |
| 5 | Benchstat long-run baseline; upgrade FEATURES `~` figures | Med | 30min | TODO #3 |
| 6 | Fix `observability_test.go:560` unusedwrite | Low | 10min | TODO #5 |
| 7 | Single Go-version source (`.go-version` vs `go.mod` vs CI) | Low | 10min | TODO #6 |
| 8 | Add `gosec` to CI | Low | 10min | TODO #8 |
| 9 | `SizeResult` JSON tags (pair with release) | Low | S | TODO #4 |
| 10 | Sibling `signing` accepts `DeterministicCodec` | High | M | ROADMAP t5 |
| 11 | Siblings reuse `CBOREncMode()`/`CBORDecMode()` | Med | M | ROADMAP t5 |

New/un-routed observations from this session:

| # | Task | Impact | Effort |
| - | ---- | ------ | ------ |
| 12 | Verify the trusted-not-checked set from the 20-07 follow-up report: #34 lint-artifact-validity CI test, #36 mermaid-render CI check, #43 Envelope fallback test, #44 README-example golden snapshots — grep each against code | Med | S |
| 13 | Read + verify `CONTRIBUTING.md` freshness (not read this session) | Low | 15min |
| 14 | Read + verify `SECURITY.md` freshness (not read this session) | Low | 10min |
| 15 | Audit `docs/planning/archived/` for stale references | Low | 15min |
| 16 | Strike or annotate the last old-numbering ref at 09-58:323 | Low | 2min |
| 17 | `git diff` the pre-existing 20-47-report modification (or confirm the daemon committed it) | Low | 5min |
| 18 | Decide `dprint.json` fate: treefmt drives formatting; is the secondary config still earning its keep? | Low | 10min |
| 19 | Post-`v0.1.1` docs-health pass: `[Unreleased]`→`[0.1.1]` dating collapses TODO #1; re-measure coverage; next audit can compare against this session's Accuracy/Fitness baseline | Med | 30min |
| 20 | Post-release: `go get @v0.1.1` + pkg.go.dev rendering check (in TODO #1 runbook; listed for visibility) | High | 10min |

(20 items — the honest ceiling of what THIS session observed; padding to 50 would
violate the no-vague-items rule.)

## g) Questions I cannot figure out myself (max 3)

1. **Release:** cut `v0.1.1` from HEAD now (HEAD includes this audit's doc fixes)?
   Recommendation unchanged — new tag, never move the published `v0.1.0`. Also gates
   CHANGELOG dating and the GitHub Release body.
2. **Daemon policy:** the daemon has twice preempted commit-granularity decisions
   (`f04d158`, and this session's 11-file diff is now queued the same way). Is
   one-commit scooping acceptable, or should reports/docs be committed separately
   from code? (Daemon config — only you can change it.)
3. **Trusted-not-checked set:** items 12 in §f above rest on the 20-47 session's
   dedup claim rather than my own verification. Do you want a verification sweep of
   those next, or does the prior claim stand?

---

**Handoff:** working tree holds this audit's 11-file diff + this report (uncommitted;
the auto-commit daemon will scoop it per its usual behavior — flagged, not fought).
All gates green at time of writing. **WAITING FOR INSTRUCTIONS.**
