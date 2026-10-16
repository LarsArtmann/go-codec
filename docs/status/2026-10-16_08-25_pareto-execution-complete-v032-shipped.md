# Status Report — Pareto Plan Executed: v0.3.2 Shipped, M2–M27 Swept, All Gates Green

**Date:** 2026-10-16 08:25 CEST
**Scope:** Full execution of `docs/planning/2026-10-16_06-40_pareto-ship-v0.3.2-execution-plan.md`
under the operator's "GET SHIT DONE — the WHOLE TODO LIST" instruction, M2→M27
minus operator-gated items, plus everything the execution surfaced.
**Branch:** master (HEAD `80bd82e`, pushed, CI green) · **Load at write time:** 86 (storm; correctness-only work all session, zero local benchmarks)
**Supersedes:** the 08:13 same-session snapshot (kept below, banner-marked) — this report is the rotation source going forward.

---

## TL;DR

**v0.3.2 is SHIPPED and verified end-to-end.** The first live fire of
`release.yml` ran green; the GitHub Release is cut from the CHANGELOG and is
`/releases/latest`; the module proxy serves the tag; a cold scratch module
`go get`s it and round-trips CBOR; and the five core go-cqrs-lite modules
(event, signing, encryption, storage/pebble, kv) adopted it with `GOWORK=off`
verify+build+test proof. Two independent CI-red incidents were found and fixed
on the way — the handoff's "CI green" claim was **wrong** (the plan push was
red on a NEW failure: govulncheck flagging four go1.27.1 stdlib CVEs), fixed
by the honest toolchain bump to **go 1.27.2** across the tripwire contract.
Then the plan's entire tooling tier landed: release dry-run mode, pre-commit
tripwire hooks, ADR-0003 + the `write-go-version` generator, ADR-0002, the
golangci pin floor guard (fifth tripwire), erraudit v0.7.0, the mermaid cache,
bench-gate `--selftest`, the **self-relative CI benchmark gate** (seeded,
chain verified), the **per-gate job summary** (rendered on a real run), the
non-nix contributor path, the AGENTS trim, and the archived corpus normalized
to **32/32 check-rows complete** (including a genuine gate bug fix). Old
releases v0.1.0/v0.2.0/v0.3.1 were backfilled. The 06:35 report was rotated
(39 verdicts → archived + manifest). **Final state: master pushed at
`80bd82e`, every CI run green, tripwires 5/5, all 27 plan items dispositioned.**

| Quick stats | |
| --- | --- |
| Plan items executed | M2–M26 except operator-gated (M1 was already done — actually was NOT: CI was red again) |
| CI runs this session | 12+ triggered, ALL green at session end |
| Releases | v0.3.2 shipped live · v0.1.0/v0.2.0/v0.3.1 backfilled (`--latest=false`) |
| TODO_LIST | 14 rows → 5 (11 closed into CHANGELOG, 1 new) |
| New tooling | `write-go-version`, `check-golangci-pin`, `repair-partial` (docs-health skill), pre-commit hook |
| ADRs | 0003 Accepted (go-version contract) · 0002 Proposed (compress-then-sign) |

## a) DONE — the ship thread (P0/P1, the 1%→64%)

1. **M1-contingency (CI red #1, discovery):** the handoff said CI was green on
   `e8e061c`; `gh run list` showed **failure**. Root cause: the lint legs were
   green (the v2.14.0 pin held) but the **vulncheck** leg had gone red —
   govulncheck@latest flagged GO-2026-6617/6613/6611/6603 (net/http, stdlib
   go1.27.1, all fixed in go1.27.2). Fixed by bumping go.mod + `.go-version` +
   `.golangci.yml` to **1.27.2** atomically (tripwire green; nixpkgs
   `go_1_27` already shipped 1.27.2, so hermetic checks followed with no
   vendorHash change). The library ships no HTTP surface; the fix was the
   floor, not a gate suppression.
2. **M2 dry-run:** `release.yml` gained `workflow_dispatch` + a `version`
   input; dispatch exercises gates + note-cutting and publishes nothing
   (release step gated on tag pushes; dry-run summary step added). Dispatched
   BEFORE the tag: gates green, the no-CHANGELOG-section fallback exercised,
   `DRY RUN OK`. actionlint clean.
3. **M3 ship:** `[Unreleased]` cut to `[v0.3.2] — 2026-10-16` ("no breaking
   changes, no Go file changes" stated explicitly), pre-tag gates green
   (replace-grep, `go mod verify`, dual-mode build/test, local awk simulation
   of the workflow's note-cut → 166-line section), annotated tag `v0.3.2`
   pushed — **first live release.yml run GREEN** (38854807936) with the full
   CI matrix green on the tag (38854807927).
4. **M4 verification:** Release not draft/prerelease, `/releases/latest` →
   v0.3.2, notes cut from the CHANGELOG section, proxy serves v0.3.2
   (`go list -m -versions`), scratch module cold `go get @v0.3.2` + CBOR
   round-trip run (17 bytes, `enc: cbor`). pkg.go.dev indexing lagged all
   session (404 on the version URL) — proxy artifact is the load-bearing
   layer and is confirmed; the render check stays owed (§f-1).
5. **M5 consumer sweep (go-ecosystem-upgrade protocol):** baseline at v0.3.1
   green (GOWORK=off, go.work stashed) → `go get @v0.3.2` in event, signing,
   encryption, storage/pebble, kv → `go mod tidy` **force-raised each
   consumer's `go` directive to 1.27.2** (dependency floor — the
   normalize-to-minor guidance loses to MVS truth) → verify+build+test green
   on go1.27.2 (kv build-verified, no test files) → go.work raised to 1.27.2
   → adoption entry in the sibling CHANGELOG → committed per-repo
   (`09ac30e81` daemon scoop + `6cc57bb51` descriptive companion; **not
   pushed** — sibling push awaits instruction). The remaining ~57 modules are
   TODO #24 (next sibling release train).

## b) DONE — tooling, policy, structure (P2/P3)

6. **M6 hooks:** `.githooks/pre-commit` runs all five tripwires with
   fast-fail fix hints; `core.hooksPath` active in this clone (it fired on
   every subsequent commit); negative test in a fresh scratch clone **BLOCKED**
   a commit with a drifted `.go-version`. First attempt taught the failure
   mode honestly: the un-committed hook made the clone's hooksPath point at
   nothing and the "block" was vacuous — the hook had to exist in the clone
   before the negative test meant anything.
7. **M7 erraudit v0.4.0 → v0.7.0** (latest; the plan guessed v0.6.x):
   ci.yml + flake app moved in lockstep (`check-erraudit-version.sh` PASS),
   local smoke ran the exact CI invocation set: zero findings, baselines hold.
8. **M8 mermaid cache:** actions/cache v4.3.0 over the npx tree + Chromium,
   keyed on the CLI version.
9. **M9 bench-gate:** `--selftest` (four synthetic load histories through a
   round-robin snapshot sampler: flat-quiet and 5..8-band must converge,
   ≥8-plateau and a rising sub-8 storm must never), `--help`, flake app
   `nix run .#bench-gate`, shellcheck in the devShell, shellcheck clean.
   Two selftest bugs were caught BEFORE landing green: an unbound arithmetic
   var under `set -u`, and a counter incremented inside a command-substitution
   subshell (every sample re-read snapshot 1, making the storm scenario pass
   for the wrong reason). Fixed via a counter file; probe path re-verified
   standalone.
10. **M10 + M22:** **ADR-0003** (one full-patch string across go.mod /
    `.go-version` / `.golangci.yml`; go.mod authoritative; mechanical
    enforcement; the YAML-float/quote trap and 2-part rejection recorded as
    incidents) + `scripts/write-go-version.sh` + `nix run .#write-go-version`
    (scratch-copy tested: rewrites all three for 1.27.3, rejects 2-part).
11. **M19 golangci pin guard:** `scripts/check-golangci-pin.sh` — the ci.yml
    pin must be ≥ the known-good floor (2.14.0, the buildir-panic fix),
    wired as the **fifth tripwire** (ci.yml v1 leg, `.#tripwires`,
    pre-commit), negative-tested against a simulated v2.12.2 pin.
12. **M20 bench gate:** a `bench` job on the weekly cron + workflow_dispatch:
    v1 suite count=10, 90-day `bench-results` artifact, self-relative compare
    against the previous run's artifact via `bench-compare.py` (wide
    0.33..3.0 bounds, GOMAXPROCS-normalized). Design honesty: comparing
    against the committed LOCAL baseline would be cross-machine fiction for
    ns/op, so the chain seeds itself — **the dispatch run completed success
    and seeded the artifact (7,906 bytes)**; the first real comparison
    happens on the next scheduled run. A seed-vs-cron event-filter bug was
    caught and fixed before it could silently re-seed forever (query all
    successful runs, attempt download per run).
13. **M21 job summary:** a `summary` job renders every gate's result into
    `$GITHUB_STEP_SUMMARY` (skipped gates labeled: fuzz/bench are
    schedule/dispatch-only, error-audit waits for ERRAUDIT_PAT) and fails on
    any always-on gate failure. **Verified on a real run: `Gate Summary:
    success`** (38860410413), all nine gates success/skipped as expected.
14. **M14 AGENTS trim:** 19,887 → 15,599 bytes with every distinct concept
    preserved (go-version gotcha → ADR pointer, nix toolchain trio merged
    into one numbered gotcha, references table de-padded); additions landed
    only against cuts. The pairing rule is now written into the file itself.
15. **M15 corpus normalization:** the archived status/planning corpus went
    from **16/32 flagged to 32/32 complete (gate exit 0)**:
    - 544 PARTIAL rows across 16 files repaired — bare ID/effort cells
      struck with every existing verdict preserved byte-for-byte, via a NEW
      docs-health tool (`annotate.go repair-partial`, 9 fixture assertions)
      built because `repair` would have double-struck already-struck cells.
    - 15 CLEAN rows inside struck tables got REAL verdicts, each verified
      against today's repo: per-target fuzz runs are covered by the ci.yml
      fuzz job; govulncheck + gitleaks live in ci.yml (gosec never added);
      flake.lock exists and `nix flake check` runs hermetically; the Codec
      godoc Examples exist (`example_test.go`); the v0.1.0-era
      project-discovery-sdk check declined as stale (adoption proven via
      go-cqrs-lite proxy builds); the Codec registration API routes to
      ROADMAP theme 1 where the idea already lives.
    - The v0.3.0 report's TOP-50 brainstorm table ledger-exempted (open rows
      are the point; dispatched rows carry verdicts).
    - Tool fix: `check-rows` recognized only 2+-dash separators — GFM
      single-dash separators were misclassified as CLEAN data rows,
      producing phantom failures. Regex fixed (`^:?-+:?$`), fixtures added,
      31/31 assertion suite green.
16. **M16 + M23 + M26:** CONTRIBUTING "Repository Checks Without Nix" (five
    scripts + the `env -u GOEXPERIMENT` pinning discipline); **ADR-0002**
    compress-then-sign (Proposed candidate; ROADMAP pointer updated); the
    indicative-figures propagation policy written into the baseline doc and
    cited from FEATURES.md.
17. **M18 backfill:** v0.1.0/v0.2.0/v0.3.1 GitHub Releases created from their
    CHANGELOG sections with `--latest=false`; release list shows all five
    tags with v0.3.2 still Latest.
18. **M13 rotation:** the 06:35 report annotated (39 verdicts — §b/§c/§f/§g;
    operator-gated items left bare), `check-rows` complete, archived with a
    manifest row; its §f harvested (TODO #24 new; ROADMAP theme 7 +2 ideas).
19. **Doc sync:** TODO_LIST 14 → 5 rows (11 closed, all logged in CHANGELOG;
    #1/#2/#3/#10 operator-gated remain; #24 new); CHANGELOG `[Unreleased]`
    carries ONLY post-tag work (a double-count of in-v0.3.2 items was caught
    and removed); this report supersedes the 08:13 snapshot.

## c) VERIFICATION LEDGER (all fresh, this session)

| Gate | Result |
| --- | --- |
| Five tripwires (`nix run .#tripwires`) | PASS (final + at every commit) |
| Dual-mode build/test (go1.27.2) | ok both modes; `nix flake check` all checks passed |
| Dual-mode golangci-lint | 0 issues both modes |
| erraudit v0.7.0 smoke (3 CI invocations) | zero findings |
| actionlint / shellcheck | clean on every workflow/script touched |
| release.yml live run (tag v0.3.2) | success (38854807936) |
| CI on the tag (full matrix) | success (38854807927) |
| Dispatch fuzz v1+v2 | success |
| Dispatch bench seed | success, `bench-results` artifact 7,906 bytes |
| Summary job first live run | success (38860410413) |
| Pre-commit hook negative test | commit BLOCKED with fix hint (scratch clone) |
| `go mod verify` (both repos) | clean |
| check-rows archived corpus | 32/32 complete, exit 0 |
| GitHub Release semantics | v0.3.2 = Latest; four other tags released |
| Proxy | serves v0.3.2 (`go list -m`), scratch cold go-get + run OK |

## d) TOTALLY FUCKED UP

1. **The handoff's "CI confirmed GREEN" was false and I almost shipped on
   it.** The 06:35 session's `gh run watch` rc=0 had missed that the vulncheck
   leg failed AFTER the watch (or was misread). Lesson recorded: trust
   `gh run list` state, not a watch exit code from an earlier moment.
2. **The daemon raced me FOUR times** (go bump batch, hooks, ADR batch,
   rotation batch) — each scooped uncommitted work into `chore:` noise before
   my descriptive commit landed, requiring soft-reset re-commits of unpushed
   history. The fix is written into AGENTS.md now (commit within the same
   breath as the edit), but I violated it while running verification between
   edits, three more times after learning it once.
3. **The bench selftest initially passed for the wrong reason.** The subshell
   counter bug made every scenario read snapshot 1 — the two "never converge"
   scenarios were actually crashing into timeout (right verdict, broken
   mechanism) and the storm scenario "converged" on a crash. Only running the
   probe standalone surfaced it. A selftest that cannot distinguish "passed"
   from "crashed" is the vacuous-test class again.
4. **`repair` nearly mangled 544 archived rows.** The obvious tool for the
   PARTIAL class appends verdicts and strikes ALL cells — double-striking the
   already-struck ones. Reading `strikeRepairRow` before batch-applying is
   the only reason the corpus is uniform instead of mangled.
5. **Two CHANGELOG structure fumbles** while inserting the Unreleased block
   (duplicate `### Changed`/`Nothing yet` tails, then double-counting
   in-v0.3.2 items as Unreleased). Both caught by re-reading the section
   structure before commit; neither shipped.

## e) WHAT WE SHOULD IMPROVE

1. **Verify CI state at the MOMENT of claiming it** (`gh run list`, not a
   stale watch rc) — the handoff-class failure.
2. **Commit in the same breath as the edit; verify after the commit.** The
   daemon's scoop window is seconds on both repos.
3. **Selftests must execute their real invocation path standalone at least
   once** — fixtures can pass while the entry path crashes.
4. **Read a transform tool's code before batch-applying it to a new damage
   shape**; extend the tool with a fixture instead of hand-rolling.
5. **Cross-machine benchmark comparisons stay banned.** Self-relative on the
   same runner class is the honest CI gate; the local baseline remains the
   reference for human-reviewed perf work only.

## f) NEXT (owed / operator-gated)

1. **pkg.go.dev render check for v0.3.2** — indexing lagged all session
   (proxy confirmed serving; pure upstream lag). One fetch of
   `pkg.go.dev/github.com/larsartmann/go-codec@v0.3.2` closes it.
2. **First real bench comparison** on the next scheduled run (Sunday 02:00
   UTC) — the chain is seeded; eyeball the `bench-compare.txt` artifact.
3. **Push the go-cqrs-lite adoption commits** (`774556d17`, `6cc57bb51` on
   its master) when the operator says so; then the ~57-module sweep rides the
   next sibling release train (TODO #24).
4. **Operator decisions unchanged:** ERRAUDIT_PAT vs publish (TODO #1; M27),
   file the erraudit drafts (TODO #2; re-verify against v0.7.0 first), daemon
   policy (TODO #3 — re-proved four times today across both repos),
   contributor secrets note (TODO #10), multi-arch builders (M24), the
   LiteStream+LTX product call (M25).
5. **No v0.3.3 warranted:** everything after the tag is docs + CI only.
6. The stale LSP gocognit warning on `json_streaming_v2_bench_test.go`
   remains stale (CLI lint 0 issues both modes) — ignore as documented.

---

**State: ALL EXECUTABLE PLAN ITEMS DONE AND VERIFIED. WAITING FOR OPERATOR
INSTRUCTIONS** (the four standing decisions + the two owed checks above).

---

## Addendum (09:10 same morning): automated sweep broke the version contract — fixed

Twenty minutes after this report, an automated formatter/lint sweep (daemon
commit `397b3b2`, 38 files) rewrote go.mod/.golangci.yml to 2-part versions
(`go 1.27`), added an unreviewed `lll` linter, and bumped the pinned
actions/cache v4.3.0 → v6.1.0; the daemon pushed it and master went red on
every push (the go-version tripwire caught it in CI, exactly as designed).
Restored in `d543914` (full-patch 1.27.2 floor, reviewed linter set, reviewed
cache pin; pure formatting kept) — CI green again (run 38864602942, all 9
gates). Sweep hazard + foreign-sweep checklist recorded in AGENTS.md Gotchas.
Also fixed on sight: CHANGELOG compare links (v0.3.1/v0.3.2 refs missing,
[Unreleased] stale) in `faab46c`.

**Owed checks update:** pkg.go.dev still 404 on `@v0.3.2` at 09:00 (proxy
serves it; indexing lag). The bench self-relative comparison was NOT left to
Sunday's cron: a workflow_dispatch (run 38864642165) was fired at 09:01 to
exercise the compare chain against the seeded artifact immediately — outcome
recorded in CHANGELOG [Unreleased] and any successor report.

*Verification sources: fresh CLI runs this session — five tripwires,
dual-mode build/test/lint on go1.27.2, `nix flake check` (job log), actionlint,
shellcheck, `go mod verify` on both repos, the scratch go-get round-trip,
`gh run` logs for every CI claim (38853654809, 38854028654, 38854057247,
38854403846, 38854629170, 38854807927, 38854807936, 38859475364, 38860315047,
38860410413), the GitHub Releases API, and the proxy via `go list -m`. No
benchmarks were run locally (loadavg 15–86 all session); the only benchmark
output produced is the CI runner's seeded artifact.*
