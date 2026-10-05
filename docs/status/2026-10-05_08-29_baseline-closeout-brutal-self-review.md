# Status Report — Baseline Close-out + Brutal Self-Review (2026-10-05 08:29)

> Scope: this session only (08:00–08:30 CEST) — the post-overnight close-out
> of benchmark items 4+7, item 13, gate repo-ization, and what I noticed while
> doing it. Point-in-time snapshot; evidence links inline. Companion reports:
> `2026-10-05_01-49_benchmark-contention-rejection-quiet-gate.html`
> (contention diagnosis) and
> `2026-10-05_07-40_brutal-self-review-overnight-gate-success.md` (overnight
> success). Format note: user explicitly demanded `.md` here; the
> status-report skill's HTML default was overridden for this file.

## Verification snapshot (end of session)

| Check                                | Result                                        |
| ------------------------------------ | --------------------------------------------- |
| `nix fmt`                            | 0 changed                                     |
| `go build ./...` (v1 + v2 modes)     | OK / OK                                       |
| `nix run .#tripwires` (4 scripts)    | 4/4 PASS (32 codes, versions agree, etc.)     |
| bench-compare self (clean v1 vs v1)  | exit 0, 210/210 in bounds                     |
| bench-compare cross (clean v1 vs v2) | exit 0, 210/210 in bounds, worst ratio 1.32   |
| Race/test suite                      | NOT re-run — no Go files touched (deliberate) |

## a) FULLY DONE (this session)

1. **Clean benchmark data secured out of /tmp** before the next cleaner pass:
   `~/bench-artifacts/go-codec/` now holds `codec-bench2-{v1,v2}.txt`, both
   `.ok` markers, both `.loadpeak` files (8.10 / 9.96), and a copy of the
   benchstat binary used (`x/perf@v0.0.0-20260929162123-406019bb8b68`,
   recorded in the baseline doc).
2. **benchstat summaries generated** for v1, v2, and the A/B view; 70 distinct
   benchmarks confirmed in each suite (68 → 70 reconciled: the two
   `StreamingJSONV2_DecoderComparison` sub-benchmarks).
3. **Variance eyeball vs the 2026-09-11 table — ACCEPTED**: hot paths ±1–4%,
   canary `RawCodec_Encode` 11.27n (v1) / 13.10n (v2) vs the 25n threshold,
   peak load under 12, no uniform inflation. Geomean 471.2n → 385.0n (v1) /
   411.2n (v2), explained by go1.27 codegen + JSON engine alloc shifts
   (e.g. `JSONCodec_Encode` 192B/6 → 88B/4 — v1 and v2 builds now
   byte-identical per row, a nice go1.27 v1-on-v2-engine confirmation).
4. **`docs/benchmark-baseline.md` rewritten**: dual-mode environment table,
   quiet-machine protocol note (gate mechanics + acceptance values + link to
   the contention-evidence report), go1.27 alloc-shift warning, cross-mode
   caveat, artifact-only retention decision (closes item 7's remainder),
   "supersedes 2026-09-11".
5. **TODO items 4 and 7 closed** (rows deleted per file convention) with
   CHANGELOG Unreleased entries (re-baseline entry + bench tooling).
6. **Item 13 closed — v0.3.1 CHANGELOG section added** from the actual tag
   diff (go-floor minor-form rationale, dep bumps, flake systems inlining,
   SECURITY.md, xargs guard). Bonus finding: **two Unreleased "Fixed" bullets
   were misattributed** — SECURITY.md 0.x and the xargs -r guard had already
   shipped IN v0.3.1; moved them into the new section. "No Go file changes"
   claim verified against the tag diff (empty).
7. **`scripts/bench-gate.sh` repo-ized** (the /tmp prototype was eaten — see
   d-2): storm-aware sustained-idle entry, canary early-abort, peak-load
   watch, 8 attempts / 10h deadline, import-safe guard. Logic unit-tested:
   5 `float_lt` cases + 4 `wait_quiet` scenarios (storm reset, boundary-8
   reset, sub-5 fast path, sustained count).
8. **AGENTS.md updated**: step-0 `/proc/loadavg` check (proactive, not
   reactive), sibling-tenant quiet-window courtesy, canary updated ≈16n →
   ≈11–13n on go1.27, `scripts/bench-gate.sh` + baseline-doc pointers,
   stale `CV/scripts/when-quiet.sh` reference removed.
9. **Gate discipline honored**: stopped at the user-gated items; no erraudit
   drafts filed; no pushes.

## b) PARTIALLY DONE

1. **bench-compare "real-output verification"** — the prior session verified
   it on the REJECTED suites; this session I extended the CHANGELOG claim
   before re-running it on the clean data (see d-1). Caught during
   self-review prep, re-run immediately: self exit 0, cross exit 0, bullet
   corrected. Right outcome, wrong order of operations.
2. **`scripts/bench-gate.sh`**: tested only via ad-hoc sourced-bash probes
   (not shellcheck'd — unavailable in devshell and on PATH; not wired as a
   flake app). It is a faithful reconstruction of the design INTENT from the
   session summary; the original prototype is gone, so byte-fidelity is
   unverifiable by construction.
3. **`benchstat-clean-ab.txt`** was generated early but never actually read;
   the bench-compare cross run ended up superseding it as the read artifact.

## c) NOT STARTED (all user-gated, by standing instruction)

- Item 1: `ERRAUDIT_PAT` secret vs publishing erraudit (user decision).
- Item 2: filing the two erraudit issue drafts upstream (external write).
- Item 3: auto-commit daemon policy + 2026-10-04 working-tree wipe root cause.
- Item 8: GitHub Releases backfill for v0.1.0/v0.2.0 (or explicit skip).
- Item 10: contributor secrets note (gated on item 1's direction).
- FEATURES.md / README benchmark-number propagation (old `~` values now cite
  a superseded table).
- A hand-made descriptive commit for today's human-meaningful changes
  (baseline rewrite, v0.3.1 section, gate script) — the daemon has been
  flattening them; per AGENTS.md a real commit should precede any push.

## d) TOTALLY FUCKED UP

1. **Introduced a factual error in the CHANGELOG and shipped it green.** I
   wrote that the bench-compare exit-1 on `RawCodec_Decode` (ratio 0.292)
   was "a genuine mode gap" — it was actually clean-v2 vs
   contention-POLLUTED-v1, i.e. the exit-1 was itself contention evidence
   (the prior session's own TODO note said exactly this). Clean data proves
   it: cross-mode ratio 0.957, exit 0. Root cause: I extended an inherited
   verified claim without re-anchoring it to the new evidence base I had
   JUST accepted. Found + fixed within the hour during self-review — but a
   reader in between would have been misled. Process rule extracted: **when
   the evidence base is replaced, every claim derived from the old base must
   be re-derived, not copied.**
2. **/tmp trusted for the third time this incident.** Overnight cleaner ate
   the staged baseline-draft prose, the bench-name extractions, AND the
   working gate script (`codec-bench-gate2.sh`). The repo-ized gate is a
   reconstruction from a conversation summary — its fidelity is a belief,
   not a fact. The data that mattered survived by luck of timing, not by
   design (the copies happened at 08:03; the cleaner ran between 01:52 and
   07:40 and could as easily have taken the results).
3. **Silent-green test harness.** My first `wait_quiet` test "passed" with
   `samples=0` (want 6) because the stub lost side effects across command
   substitutions AND my assertions printed `(want N)` annotations instead of
   failing on mismatch. Green output was about to be accepted; only reading
   the numbers caught it. Assertions must FAIL, never annotate.

## e) WHAT WE SHOULD IMPROVE

1. Re-anchor inherited claims when their evidence is replaced (d-1 rule).
2. Record the machine-local artifact location (`~/bench-artifacts/go-codec/`)
   somewhere durable (AGENTS.md line or a pointer in the baseline doc) —
   right now only this report knows it exists.
3. Quiet-gate split brain: CV's `when-quiet.sh` and our `bench-gate.sh`
   encode the same concept in two repos — consolidate into a shared sibling
   (collector-utils-style home) when a third repo needs one.
4. `bench-gate.sh` header does not document the known sub-60s-burst blind
   spot (1-min loadavg lag) that the 07-40 report called out; add it.
5. Ad-hoc test discipline: fail-on-mismatch assertions (d-3).
6. Wire `bench-gate.sh` as `nix run .#bench-gate` (tripwires-style) and get
   shellcheck into the devshell; both are 10-minute items.
7. The CHANGELOG Unreleased section now carries real user-facing content
   (release workflow, re-baseline, tooling) — a v0.3.2 tag is getting due;
   cut it deliberately rather than letting Unreleased grow unbounded.

## f) Next things (ordered by impact; ⏳ = user-gated)

1. ⏳ Answer g-1/g-3 below → unblocks items 1/2/3/8/10 + number propagation.
2. Hand-make one descriptive commit for today's baseline/v0.3.1/gate changes
   before any push (daemon flattening otherwise erases the narrative).
3. Add the artifact-location pointer (e-2) to AGENTS.md (2 min).
4. Add the sub-60s blind-spot note to `bench-gate.sh` header (2 min).
5. Cut v0.3.2 from Unreleased (release.yml is ready; notes cut from
   CHANGELOG — verify the workflow end-to-end on a scratch tag first).
6. Propagate new baseline numbers into FEATURES.md/README once g-3 answered.
7. Wire `.#bench-gate` flake app; add shellcheck to devShell.
8. Consolidate quiet-gate scripts across CV + go-codec (e-3).
9. Item 10 (contributor secrets note) once item 1's direction is set.
10. Re-run the FULL test suite (race ×2 modes) at the next natural Go change;
    today's scope-down was justified but must not become the norm.
11. Record `benchstat` version pinning strategy (it was installed @latest —
    a future run may diff against a different benchstat; consider `go run`
    with a pinned x/perf version in the flake).
12. Consider a tiny self-test mode (`bench-gate.sh --selftest`) encoding the
    four wait_quiet scenarios so the logic tests live with the script.
13. Sweep FEATURES.md status claims against the new baseline table (drift
    check) — folds into 6.
14. Check whether `docs/drafts/` still holds the two erraudit drafts in the
    working tree (git status showed them added-then-deleted at session
    start; confirm they are committed before any filing work resumes).
15. Next session step-0: `/proc/loadavg` (now encoded in AGENTS.md — follow
    it).
    16–25. (No further honest items without researching beyond this session's
    scope, per instruction. Items 16–25 intentionally not padded.)

## g) Questions I cannot answer myself

1. **Daemon wipe root cause (item 3):** the 2026-10-04 working-tree wipe
   (reverted in-flight drafts + flake fixes, recovered via dangling commit
   `d75c8a1`) — do you want the one-commit-scoop vs split-commits policy
   settled now, and is the wipe itself explainable from the daemon's config
   on your side? I cannot inspect the daemon's configuration.
2. **erraudit drafts:** file the two prepared upstream issues now, or do you
   want to read them first? (External write; drafts were voice-checked
   2026-10-04 but their working-tree state needs the check in f-14.)
3. **Benchmark numbers in user-facing docs:** propagate the new dual-mode
   table into FEATURES.md/README now, or leave the `~` values until each
   doc is next touched for other reasons?

---

_Written 2026-10-05 08:29 CEST. Auto-commit daemon will pick this up; no
manual commit made (harness contract: no commits without explicit request)._
