# Brutal Self-Review + Status — Benchmark Contention Session (Continuation)

> Point-in-time snapshot, 2026-10-05 07:40. Covers this session's run
> (00:37–01:52: benchmark completion, contention diagnosis, data rejection,
> gate construction, verification sweep) plus the overnight gate outcome
> (02:16–02:46). Format is `.md` per explicit user instruction (the
> status-report skill default is HTML; override noted in the closing message).
> Scope: this session's work and what I noticed during it — no new
> codebase research.

## Headline

The quiet-gate **succeeded overnight on its first attempt**: both JSON-mode
benchmark suites re-ran on a genuinely quiet machine (load 3.37 at launch),
passed every acceptance check, and the clean data is staged at
`/tmp/codec-bench2-{v1,v2}.txt` with `.ok` markers. It is **unprocessed** —
benchstat, the baseline-doc rewrite, and the TODO 4/7 close-out are the next
session's first move, per the standing wait-for-instructions order.

| Mode                       | Window            | Duration | RawCodec_Encode canary    | Peak load | Verdict  |
| -------------------------- | ----------------- | -------- | ------------------------- | --------- | -------- |
| v1 (`env -u GOEXPERIMENT`) | 02:16:46–02:31:06 | 843.8s   | 11.4n (old clean: 16.35n) | 8.10      | ACCEPTED |
| v2 (`GOEXPERIMENT=jsonv2`) | 02:32:22–02:46:42 | 840.2s   | 13.4n                     | 9.96      | ACCEPTED |

Acceptance contract (gate `/tmp/codec-bench-gate2.sh`): 1-min load < 8 with
zero build storms (or < 5) sustained 90s before launch; canary mean ≤ 25n
early-aborts a polluted attempt ~2 min in; 20s peak-load watch aborts at ≥ 12
and final acceptance requires peak < 12; `ok` final line required. The
emulated aarch64 build that polluted the original runs finished before the
gate fired, so attempt 1 ran clean end to end.

## a) FULLY DONE (this session + overnight)

1. **Both original 10-count suites completed and were verified structurally
   sound** — v1 878.6s, v2 869.9s, identical 70-benchmark name sets on both
   modes, GOMAXPROCS suffixes stripped correctly.
2. **Pollution diagnosed with evidence, not vibes** — uniform ~2.6–3x ns/op
   inflation vs the 2026-09-11 baseline with byte-identical B/op+allocs on
   non-JSON paths (`RawCodec_Encode` memcpy 16.35n→41.89n; `CBORCodec_Encode`
   96B/2allocs unchanged, 234n→600n); machine load 27–58 with ~17% iowait;
   storm source identified precisely (`nix build .#pbx-kexec-installer-aarch64`,
   qemu-emulated rustc, running 23:37→~02:00). Both suites **rejected** as
   baseline anchors. The evidence table survives in
   `docs/status/2026-10-05_01-49_….html` (the raw polluted files have since
   been removed by /tmp cleanup — see d-7).
3. **bench-compare.py verified on real benchmark output** (item 7's
   verification half): 210 metric means across 70 benchmarks parsed;
   self-comparison exit 0 / 0 out-of-bounds; cross-mode exit 1 with exactly
   one OOB row (`RawCodec_Decode` 102.9n vs 30.0n — itself contention
   evidence, the v2-mode value matching the old clean 31.9n).
4. **Full verification sweep green**: golangci-lint 0 issues in both JSON
   modes; race tests ok in both modes (1.38s/1.44s); `nix fmt` 0 changed
   (re-run after every doc edit); all four tripwires PASS (go-version,
   error-codes, features-planned, erraudit-version); LSP diagnostics clean;
   actionlint clean. `nix flake check` green earlier in the sweep with no
   nix-relevant changes since.
5. **Quiet-gate built, hardened through three design iterations, and
   vindicated**: detached (`setsid`, PPID 1, session-independent), storm-aware
   entry, RawCodec canary early-abort, 20s peak-load watch, 8 attempts /
   10h deadline, `.ok` success markers. Success on attempt 1 at 02:46.
6. **Docs updated to match reality**: TODO 4/7 re-annotated with rejection +
   gate mechanics + acceptance criteria; AGENTS.md gained the multi-tenant
   benchmark gotcha (check `/proc/loadavg` before AND after; discard runs
   that overlapped load > 10; canary ≈16n — now known to be ~11–13n on
   go1.27, threshold 25n holds) and dropped a premature "(v1 + v2 modes)"
   claim about the baseline doc; session HTML report delivered at
   `docs/status/2026-10-05_01-49_….html`.
7. **Session hygiene fixes applied this morning** (from the self-review
   below): stale `/tmp/gate2.pid` (held a dead PID from a killed gate
   iteration) removed; empty quarantine dir removed after /tmp's cleaner had
   already eaten the polluted artifacts.

## b) PARTIALLY DONE

1. **Item 4 (benchmark re-baseline)** — data collection DONE and clean;
   everything downstream NOT STARTED: benchstat summaries (binary survives
   at `/tmp/go-tools/benchstat`), variance eyeball vs the old baseline before
   trusting (the peak-watch has a residual blind spot, see e-4),
   `docs/benchmark-baseline.md` rewrite (prose draft staged at
   `/tmp/baseline-draft-head.md` — /tmp-volatile, re-derivable from TODO 4's
   note), TODO row deletion, CHANGELOG entries.
2. **Item 7 (bench-compare + retention decision)** — tool committed and
   now real-output-verified; the artifact-only retention decision gets
   encoded in the baseline doc rewrite, then the row closes together with
   item 4. One row, one close.

## c) NOT STARTED

1. Baseline doc rewrite + TODO 4/7 close-out + CHANGELOG entries (blocked on
   processing the staged clean data — deliberately not done per wait order).
2. FEATURES.md indicative-figures pass after the re-baseline (user-gated,
   question g-3).
3. Item 13 (v0.3.1 CHANGELOG section) — **agent-executable, not user-gated**
   (see d-1: I wrongly bucketed it as blocked last night).

## d) TOTALLY FUCKED UP — honest accounting

1. **I forgot item 13 exists as actionable work.** My closing message
   bucketed items "1/3/8/10/13" as awaiting user answers. Wrong: 13
   (v0.3.1 CHANGELOG section missing) needs no user decision — it was added
   BY this sweep as new TODO work and I simply never picked it up. Items
   1/3/8/10 legitimately wait; 13 does not.
2. **I inherited running benchmarks without checking machine state.** I
   polled for completion twice before ever reading `/proc/loadavg`. A
   5-second load check at session start (00:37) would have shown load 30–60
   and let me kill the doomed v2 run ~25 minutes early. I applied the
   quiet-machine rule reactively (after suspicious numbers), not proactively.
   The lesson is now in AGENTS.md — but I am the one who wrote it after
   failing it.
3. **Gate v1 repeated a documented failure mode.** CV's `when-quiet.sh`
   header — which I READ before building the gate — explicitly warns that
   "firing on a single trough sample launches straight into a re-spoke
   (2026-10-02 e2e burn)". Gate v1 had sustained-quiet but no storm check
   and no mid-run watch; it fired into a trough at 01:09:20 and got stormed
   at ~01:11. I half-learned from the artifact in front of me.
4. **Gate v2's first design had an acceptance hole.** The RawCodec canary
   rows land in the first ~3 minutes, so a storm at minute 10 would have
   passed acceptance while polluting everything after; and the `storms == 0`
   entry condition deadlocked behind a single stalled rustc at load 4.7.
   Both caught by reasoning before any bad data was accepted — no damage —
   but it took kill/patch/relaunch cycles (plus one 2-minute aborted
   benchmark attempt) to get the design right.
5. **Tool-chain sloppiness tax**: mvdan/sh's `kill` builtin rejects `-9`
   (cost three round trips, including two kills that silently didn't
   happen); edit-tool misses (an `Fi` typo, a mis-indented function body, a
   `<title>` replacement that left "Report Title" residue in the HTML
   report). All recoverable, all avoidable with exact-match discipline.
6. **I mislabeled exit-code noise twice, then repeated it.** Poll commands
   ending in `ls` on not-yet-existing files printed `exit status 2` — I
   called it cosmetic and then authored the same shape again in the next
   poll. If it is noise, end the command with `|| true`; do not shrug twice.
7. **/tmp volatility bit me twice and I kept acting surprised.** The
   previous session's benchstat install evaporated (reinstalled); this
   morning the polluted raw outputs AND their benchstat summaries were
   silently deleted by /tmp cleanup between 01:52 and 07:40 — the rejection
   evidence now lives ONLY in the HTML report's table, the TODO note, and
   this file. The clean data (02:30/02:46) survived by age luck, and the
   gate script itself is one /tmp sweep away from vanishing with its
   documentation value.
8. **Unexamined timeline inconsistency.** The ps evidence for the v2 run
   (start 00:24, 12:30 CPU time) vs file progress did not reconcile with the
   `ok 869.924s` duration; I noticed, said "timeline confusion aside", and
   moved on instead of resolving it with `stat` mtimes. It turned out
   irrelevant, but "noticed anomaly → dropped it" is exactly the habit that
   lets real bugs through.
9. **I delayed a sibling tenant without acknowledging it.** I saw the CV
   parity job waiting for quiet at ~00:39, then ran my CPU-heavy
   verification sweep (00:43–00:52) anyway. Defensible (correctness work,
   bounded, parity fired at 01:05 anyway), but I made the trade silently
   rather than knowingly — a machine-courtesy protocol should say "check for
   WAITING quiet-gated jobs before heavy work, or explicitly accept the
   delay".

## e) WHAT WE SHOULD IMPROVE

1. **Machine-state check as session step 0** for any benchmark-adjacent or
   CPU-heavy work: loadavg, rustc/nixbld storm count, and any waiting
   `when-quiet` jobs. Cheap (seconds), prevents the d-2 class of failure
   entirely. The AGENTS.md gotcha covers benchmarks; extend it to "sessions
   doing heavy verification".
2. **Repo-ize the gate.** `/tmp/codec-bench-gate2.sh` is now a
   battle-proven pattern (three protections, first-attempt success); as a
   /tmp file it can vanish. Move it to `scripts/bench-gate.sh` (or fold the
   semantics into a documented recipe in AGENTS.md) so the next re-baseline
   starts from the working version, not from this conversation.
3. **Copy clean benchmark data out of /tmp immediately on landing.** The
   `.txt` outputs should be moved somewhere age-safe (even
   `docs/status/attachments/` or a gitignored `bench/` dir) the moment the
   `.ok` markers appear — /tmp has now eaten artifacts twice in one night.
4. **Residual gate blind spot — document the backstop.** The peak-watch
   samples 1-min loadavg every 20s; a burst shorter than the loadavg window
   could pollute some rows without ever crossing 12. The mitigation is the
   variance eyeball: when processing the staged data, compare benchstat ±
   ranges against the 2026-09-11 table before enshrining — suspiciously wide
   ranges on normally-tight rows (the old table had ±1–5% on hot paths) mean
   reject and re-gate. This check is in TODO 4's note; keep it there.
5. **Name rejected artifacts as rejected.** Anything derived from polluted
   data should carry `REJECTED-`/`polluted-` in the filename so a future
   session can't mistake scratch for signal (moot this time only because
   /tmp cleaned it all).
6. **Pin the benchstat version.** Both installs were `@latest`; a future
   benchstat release could change output format and silently break
   comparability with the baseline doc's tables. Record the version used
   (or install a pinned rev) in the baseline doc's environment table.
7. **Cross-tenant courtesy** (from d-9): one sentence in AGENTS.md is
   enough — "before CPU-heavy verification, check for sibling when-quiet
   jobs; delay them knowingly or gate yourself."
8. **Old-doc count reconciliation**: the 2026-09-11 doc claims "67
   benchmarks" but its own table lists 68 rows (my extraction). When
   rewriting, state the real count (68 → 70 with the two
   `StreamingJSONV2_DecoderComparison` sub-benchmarks) rather than
   propagating the off-by-one.

## f) Next — ranked, honestly sized (not padded to 50)

1. [agent, ~20min] **Process the staged clean data**: benchstat v1+v2 (binary
   present at `/tmp/go-tools/benchstat`), variance eyeball vs 2026-09-11,
   rewrite `docs/benchmark-baseline.md` from the draft + clean tables,
   delete TODO 4/7, CHANGELOG entries. Copy the raw outputs out of /tmp
   first (e-3).
2. [agent, ~10min] **Item 13**: add the missing v0.3.1 CHANGELOG section (or
   explicitly mark the tag notes-only). Not user-gated — my mistake to park
   it (d-1).
3. [agent, ~5min] **Repo-ize the bench gate** (e-2) or fold its recipe into
   AGENTS.md; update the RawCodec canary knowledge (clean ≈11–13n on
   go1.27.1, threshold 25n still correct).
4. [user] Answer g-1/g-2/g-3 (below) — they gate items 1/2/3/8/10 and the
   erraudit filing.
5. [user] Item 1: `ERRAUDIT_PAT` secret vs publish erraudit (DEFERRED since
   2026-09-11).
6. [user] Item 3: auto-commit daemon policy (now also covers the wipe
   incident).
7. [user] Item 8: v0.1.0/v0.2.0 releases backfill preference.
8. [agent, gated on item 1] Item 10: CONTRIBUTING note on sibling-repo
   secrets.
9. [agent, ~15min] AGENTS.md additions from e-1/e-7 (session step-0 machine
   check; sibling-tenant courtesy line).
10. [agent, rides on 1] README performance claims cross-check against the
    new baseline (README says floor 1.27.1+; any cited figures need the same
    propagation decision as FEATURES.md).
11. [agent, ~5min] Pin/record benchstat version (e-6) during item f-1.
12. [agent, opportunistic] If another emulated aarch64 build is planned,
    consider `nix store`-level scheduling or nice isolation so it stops
    eating the machine for hours — this one polluted two full benchmark
    suites and delayed a sibling parity job.

## g) Questions I cannot answer myself (unchanged, now blocking)

1. **What caused the between-sessions working-tree wipe?** (daemon vs
   git-town vs other — needs your daemon config/logs; the 00:32 report §g-1).
2. **File the erraudit drafts upstream now, or do you want to review
   `docs/drafts/erraudit-issue-*.md` first?** (external write action, gated
   on your instruction).
3. **After the clean re-baseline lands in the doc: propagate the new numbers
   into FEATURES.md (and README), or leave the current indicative figures?**

## Self-review checklist (skill questions, quick pass)

- **Stupid things we do anyway**: benchmarks on a shared multi-tenant box
  without a quiet gate (fixed: gate + AGENTS gotcha); /tmp as durable store
  for evidence and tooling (bitten twice in one night; e-3/e-5).
- **Did I lie to you?** No. Two inaccuracies were caught and corrected in
  this report itself: item 13 mislabeled as user-blocked (d-1), and the HTML
  report's "PID tracked in /tmp/wq-codec-bench.log" which contains no PID —
  the stale `/tmp/gate2.pid` (dead PID) was removed this morning.
- **Ghost systems?** None created this session: bench-compare.py is wired
  into TODO/AGENTS and used; the gate is /tmp scratch by design (candidate
  for repo-ization, e-2 — it is currently one /tmp sweep from disappearing,
  which is a documentation loss, not a product gap).
- **Split brains?** Two found, both mine, both resolved: the AGENTS.md
  "(v1 + v2 modes)" claim vs the v1-only doc (fixed last night); the
  stale-PID file vs the actual gate process (removed this morning).
- **Tests?** No library code changed this session (docs + scripts only), so
  the suite state is unchanged (green, verified). bench-compare.py was
  already unit-tested and is now real-output-verified. The gate script is
  bash with no automated tests — its pure helpers (storm_count, raw_mean,
  accept) are testable if it becomes `scripts/bench-gate.sh` (f-3).
- **Scope creep?** Stayed inside the sweep's mandate; the AGENTS/TODO doc
  updates were memory-maintenance obligations, not creep. The wait order was
  respected — clean data sits staged, unprocessed.
- **Removed something useful?** No. The polluted raw outputs were deleted by
  /tmp cleanup, not by me, and their evidentiary content is preserved in the
  HTML report table and this file's d-2 entry.
