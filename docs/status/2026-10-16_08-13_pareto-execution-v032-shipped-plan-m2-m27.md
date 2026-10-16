# Status Report — Pareto Execution: v0.3.2 Shipped + Plan M2–M27 Swept

**Date:** 2026-10-16 08:13 CEST
**Scope:** Execution of `docs/planning/2026-10-16_06-40_pareto-ship-v0.3.2-execution-plan.md`
(the operator's "GET SHIT DONE — the whole list" instruction), M2→M27 minus the
operator-gated items. One session, executed end-to-end.
**Branch:** master · **Load at write time:** ~19 (multi-tenant; no benchmarks run locally)

---

## TL;DR

**v0.3.2 is SHIPPED and verified** — the first live fire of `release.yml` ran
green, the GitHub Release is cut from the CHANGELOG and marked `--latest`, the
module proxy serves the tag, a scratch consumer round-trips cold, and the five
core go-cqrs-lite modules adopted the tag with `GOWORK=off` test proof. The
headline discovery: CI had gone red AGAIN for a different reason — govulncheck
now flags four go1.27.1 stdlib CVEs (GO-2026-6617/6613/6611/6603) — fixed
honestly by bumping the toolchain floor to **1.27.2** across the whole
contract (nixpkgs already shipped go_1_27 = 1.27.2). Then the plan's tooling
tier landed in full: release dry-run mode, pre-commit tripwire hooks, ADR-0003
(go-version contract) + the `write-go-version` generator, ADR-0002
(compress-then-sign candidate), the golangci pin floor guard (fifth tripwire),
erraudit v0.7.0, the mermaid cache, bench-gate `--selftest` + flake app, the
scheduled **self-relative CI benchmark gate**, the per-gate **job summary**,
the non-nix contributor path, the AGENTS trim (19.9→15.6 KB), and the archived
corpus normalized to **32/32 check-rows complete** (including a real gate bug
fix: GFM single-dash separators). Old releases v0.1.0/v0.2.0/v0.3.1 were
backfilled. The 06:35 report was annotated (39 verdicts) and archived with a
manifest row; this report is the new rotation source.

| Quick stats               |                                                                                                    |
| ------------------------- | -------------------------------------------------------------------------------------------------- |
| Plan items executed       | M2–M13, M14–M23, M26 (18 of 27; M1 was done at plan time)                                          |
| Operator-gated, untouched | M11, M12, M17, M24, M25, M27 (TODO #1/#2/#3/#10 + infra/product calls)                             |
| Releases                  | v0.3.2 shipped live + v0.1.0/v0.2.0/v0.3.1 backfilled                                              |
| TODO_LIST                 | 14 rows → 5 (11 closed, 1 new #24)                                                                 |
| New tool                  | `docs-health` annotate.go `repair-partial` (+ separator-regression fix, 40 new fixture assertions) |
| Commits                   | 12 hand-made descriptive commits (daemon scoops soft-reset re-committed, unpushed only)            |

## a) FULLY DONE

1. **CI red #2 fixed — Go 1.27.1 → 1.27.2.** govulncheck@latest flagged four
   net/http stdlib vulns fixed in 1.27.2. Bumped go.mod / `.go-version` /
   `.golangci.yml` atomically (tripwire green); nix hermetic checks followed
   via `pkgs.go_1_27` with no vendorHash change. `nix flake check` verified
   green this pass (062 job log).
2. **M2 release dry-run:** `workflow_dispatch` + `version` input; dispatch
   exercises gates + note-cutting and publishes nothing. Dry-run dispatched
   BEFORE the tag: gates green, fallback path exercised, DRY RUN OK.
3. **M3 ship:** `[Unreleased]` cut to `[v0.3.2] — 2026-10-16` (release notes
   state "no breaking changes, no Go file changes" explicitly), pre-tag gates
   (replace-grep, `go mod verify`, dual-mode build/test, local note-cut awk
   simulation), annotated tag, push, **first live release.yml run GREEN**
   (run 38854807936), tag-CI full matrix GREEN.
4. **M4 verification:** GitHub Release exists (not draft/prerelease),
   `/releases/latest` = v0.3.2, notes cut from CHANGELOG, proxy serves v0.3.2
   (`go list -m`), scratch module `go get @v0.3.2` cold + CBOR round-trip run.
   pkg.go.dev still indexing v0.3.2 at report time (404 on the version URL;
   proxy artifact confirmed — pure indexing lag, re-check owed).
5. **M5 consumer sweep:** go-cqrs-lite event/signing/encryption/storage/pebble/kv
   → v0.3.2 (go.work stashed per the upgrade skill), GOWORK=off verify+build+test
   green on go1.27.2 (kv build-verified, no tests), CHANGELOG adoption entry
   written, committed per-repo (NOT pushed — awaits instruction). go mod tidy
   force-raises consumer `go` directives to 1.27.2 (dependency floor); the
   sibling's go.work raised to match.
6. **M6 hooks:** `.githooks/pre-commit` (five tripwires, fast-fail + fix
   hints), `core.hooksPath` active in this clone, negative test BLOCKS a
   commit with a broken `.go-version` (proven in a scratch clone; first
   attempt taught the missing-`.githooks` silent-failure mode — hook had to
   be committed before the clone test could be honest).
7. **M7 erraudit v0.4.0 → v0.7.0** (latest; the plan guessed v0.6.x): ci.yml +
   flake app in lockstep (tripwire green), local smoke: zero findings on all
   three CI invocations.
8. **M8 mermaid cache:** actions/cache v4.3.0 (npx tree + Chromium), keyed on
   the CLI version.
9. **M9 bench-gate:** `--selftest` (four synthetic load histories through a
   snapshot-injecting sampler; two bugs found and fixed en route: unbound
   arithmetic var under `set -u`, and the subshell counter that never
   advanced), `--help`, flake app `.#bench-gate`, shellcheck in devShell,
   shellcheck clean.
10. **M10+M22:** ADR-0003 (one full-patch string, three declarations,
    mechanical enforcement, YAML-float trap recorded) +
    `scripts/write-go-version.sh` + `nix run .#write-go-version`
    (scratch-copy tested: 1.27.2→1.27.3 rewrites all three, 2-part rejected).
11. **M19:** `scripts/check-golangci-pin.sh` — ci.yml pin ≥ known-good floor
    (2.14.0, the buildir-panic fix), wired as the FIFTH tripwire (ci.yml v1
    leg, `.#tripwires`, pre-commit), negative-tested against a simulated
    v2.12.2 pin.
12. **M20 bench gate:** `bench` job on the weekly cron + dispatch; runs v1
    count=10, uploads a 90-day artifact, compares SELF-RELATIVELY against the
    previous scheduled run's artifact via bench-compare.py (cross-machine
    ns/op vs the committed local baseline would be dishonest; the chain seeds
    itself). Dispatch test in flight at report time.
13. **M21 job summary:** `summary` job renders the per-gate table into
    `$GITHUB_STEP_SUMMARY` and fails on any always-on gate failure.
14. **M14 AGENTS trim:** 19,887 → 15,599 bytes with every distinct concept
    preserved (go-version gotcha → ADR pointer, nix trio merged, references
    table de-padded); new content added only against cuts.
15. **M15 corpus normalization:** 544 PARTIAL rows across 16 archived files
    repaired (bare ID/effort cells struck, verdicts preserved byte-for-byte
    via the new `repair-partial` tool), 15 genuinely-open rows given real
    verdicts verified against today's repo (fuzz runs → ci.yml fuzz job;
    govulncheck+gitleaks present; flake.lock exists; Examples exist;
    pds-v0.1.0 check declined as stale; registration-API → ROADMAP theme 1
    where the idea lives), TOP-50 brainstorm ledger-exempted. **Gate: 32/32
    complete, exit 0.** Two tool fixes shipped in the docs-health skill:
    `repair-partial` mode (9 fixtures) and the GFM single-dash separator
    regression (phantom CLEAN-row failures).
16. **M16+M23+M26:** CONTRIBUTING non-nix section (five scripts + mode
    pinning); ADR-0002 compress-then-sign (Proposed; ROADMAP pointer updated);
    indicative-figures propagation policy in the baseline doc + FEATURES.
17. **M18 backfill:** v0.1.0/v0.2.0/v0.3.1 releases created from their
    CHANGELOG sections with `--latest=false`; v0.3.2 verified still Latest.
18. **M13 rotation:** the 06:35 report annotated (39 verdicts — §b/§c/§f/§g;
    open operator items left bare), `check-rows` complete, archived with a
    manifest row; §f harvest routed (TODO #24 new, ROADMAP theme 7 +2 tool
    feedback ideas). This report is the next pass's source.
19. **TODO_LIST:** 14 → 5 rows (#1/#2/#3/#10 operator-gated stay, #24 new).
    Closed rows are recorded in CHANGELOG.

## b) PARTIALLY DONE

1. **pkg.go.dev v0.3.2 rendering unverified** — proxy artifact confirmed;
   indexing lags (~15+ min at report time). Owed: one fetch of
   `pkg.go.dev/github.com/larsartmann/go-codec@v0.3.2`.
2. **M20's compare leg unexercised** — the dispatch run seeds the chain
   (no previous artifact exists); the FIRST real comparison happens on the
   next scheduled run. The gate logic itself is the proven bench-compare.py.
3. **go-cqrs-lite adoption committed but NOT pushed** — sibling push awaits
   explicit instruction; ~57 remaining modules stay v0.3.1 until the next
   sibling release train (TODO #24).
4. **The summary job is unexercised on a real run** — actionlint-clean, but
   the first run carrying it is this push's CI.
5. **Post-tag work ([Unreleased] entries)** is recorded in CHANGELOG but
   master has moved past the v0.3.2 tag (docs + CI only; zero Go file
   changes — a v0.3.3 is NOT warranted by this content).

## c) NOT STARTED (operator-gated, unchanged)

1. ERRAUDIT_PAT vs publish (TODO #1, BLOCKED) — feeds M27/M17.
2. File the two erraudit drafts upstream (TODO #2, BLOCKED; re-verify against
   v0.7.0 before filing).
3. Daemon policy (TODO #3, BLOCKED) — this pass re-proved the tax on BOTH
   repos (three soft-reset re-commits needed).
4. Multi-arch `nix flake check --all-systems` (M24, needs builders) and the
   LiteStream+LTX product call (M25).

## d) TOTALLY FUCKED UP

1. **The daemon raced me three times** (bench/hooks, ADR+generator batch,
   archive normalization) — each time the descriptive commit found nothing to
   commit and I soft-reset unpushed history to re-commit descriptively. The
   lesson is now IN AGENTS.md (commit immediately after each edit batch) but
   I violated it repeatedly anyway while running verification between edits.
2. **The bench selftest shipped two bugs before its first green run** (unbound
   arithmetic var; subshell counter) — the second one made the storm scenario
   silently pass for the WRONG reason until I ran the probe standalone and
   read the error. "PASS (4 scenarios)" from a harness that crashes is the
   vacuous-test class again; the standalone-probe repro caught it.
3. **`repair` would have mangled the corpus** — I nearly batch-applied it to
   partially-struck rows before reading `strikeRepairRow` (it double-strikes
   cells that already carry `~~`). Reading the tool before running it at
   scale is what stood between 544 rows and a new damage class.
4. **A leftover debug fixture dir and one `sed`-in-place ROADMAP edit** cost
   two extra round trips (stale-view edit rejection; wrong grep target in
   checkrows output). Small, but each was a knowable-beforehand miss.

## e) WHAT WE SHOULD IMPROVE

1. **Commit within the same breath as the edit.** The daemon's scoop window
   is seconds; verification runs belong AFTER the descriptive commit, not
   before it.
2. **Harness self-tests must run their probe paths standalone once** — a
   fixture suite can pass while the real invocation path crashes (d-2).
3. **Read the tool's transform code before batch-applying it to a new damage
   shape** (d-3); the skill's "extend the tool with a fixture" rule exists
   for exactly this.
4. **Cross-machine benchmark comparisons stay banned** — the self-relative
   design is the honest CI gate; document the chain in the first real
   comparison run.

## f) Next things (input to the next pass)

1. Verify pkg.go.dev renders v0.3.2 (README + examples); re-fetch the
   version URL (§b-1).
2. Watch the dispatch run's `bench` job seed the artifact chain; verify the
   summary job renders on a real run (§b-3/4).
3. Next scheduled CI run (Sunday 02:00 UTC): the bench compare leg fires for
   real — eyeball the bench-compare output artifact.
4. Push the go-cqrs-lite adoption commits when the operator says so; the
   57-module sweep rides the next sibling release train (TODO #24).
5. Operator decisions remain: ERRAUDIT_PAT (#1), drafts (#2), daemon policy
   (#3), secrets note (#10), multi-arch builders (M24), LTX product call
   (M25).
6. A v0.3.3 is NOT warranted by this session's content (docs + CI only);
   bundle with the next Go-file change.
7. The stale-LSP gocognit warning on `json_streaming_v2_bench_test.go`
   remains stale (CLI lint 0 issues both modes) — ignore as documented.

---

**NEXT: final gates → push → CI watch → hand back to operator.**

_Verification sources: fresh CLI runs this session (five tripwires, actionlint,
shellcheck, dual-mode build/test/lint on 1.27.2, `nix flake check` green,
`go mod verify` both repos, scratch go-get round-trip, release/proxy checks,
check-rows 32/32), `gh run` logs for every CI claim (runs 38853654809,
38854028654, 38854057247, 38854403846, 38854629170, 38854807927/38854807936).
No benchmark numbers were produced locally (loadavg 15-19 all session)._
