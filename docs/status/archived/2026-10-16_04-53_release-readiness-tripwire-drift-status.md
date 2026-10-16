# Status Report — Release Readiness & Go-Version Tripwire Drift

**Date:** 2026-10-16 04:53 CEST
**Scope:** This session only (per operator instruction). Two exchanges: (1) release-readiness assessment for go-codec after v0.3.1; (2) correction of a wrong claim in that assessment, which exposed and fixed a long-red go-version tripwire.
**Branch:** master · Working tree at session end: clean (auto-commit daemon landed the fix as chore-noise)

---

## TL;DR

Asked "time for a new version?" I assessed: **no release required** — 26 commits since v0.3.1, all daemon noise, **zero production `.go` changes**. Recommended an optional v0.3.2 PATCH to live-fire the never-run `release.yml`. My assessment contained one wrong fact ("toolchain floor 1.26.7 → 1.27.1") which the operator caught. Investigating it revealed the **go-version tripwire had been failing since before the v0.3.1 tag** — go.mod said `go 1.27`, `.go-version` said `1.27.1`, and `.golangci.yml` had been re-quoted to `go: "1.27"` (unsatisfiable under the tripwire's raw-token compare). Fixed all three to `1.27.1`; tripwire PASS; both JSON build modes green. **v0.3.1 shipped with a red verification gate and nobody noticed.**

| Quick stats | |
| --- | --- |
| Production code changed this session | 0 files (metadata only: go.mod, .golangci.yml, AGENTS.md) |
| Verifications run | tripwire fail→PASS, dual-mode builds GREEN (2) |
| Releases cut | 0 |
| Wrong claims caught by operator | 1 |
| Latent breakage discovered | 1 (tripwire red since ≤ v0.3.1 tag) |

---

## Session narrative

### Exchange 1 — "Time for a new version soon?"

Ran go-release skill Phase 0:

- Tags: `v0.1.0`, `v0.2.0`, `v0.3.0`, `v0.3.1`.
- `v0.3.1..HEAD`: 26 commits, **all** `chore: auto-commit N changed file(s)` noise. Features: 0, Fixes: 0, Breaks: 0.
- Diff `v0.3.1..HEAD`: 29 files, +4697/−194 — **no production `.go` file at all.** Tests (`codec_test.go`, `export_test.go`, `testdata_test.go`), tooling (flake.nix/flake.lock, NEW `scripts/bench-compare.py` + `scripts/bench-gate.sh`), CI (NEW `release.yml` 79 lines, NEW `dependabot.yml`, ci.yml hardening), docs (CHANGELOG 87 lines, benchmark-baseline 520 lines, research/ + drafts/), dep bumps, `.go-version` 1.26.7→1.27.1, `.golangci.yml` 41 lines.
- Deps: cbor v2.9.4→v2.9.6, go-error-family v0.11.0→v0.12.0, gjson 1.19.0→1.20.0, x/net 0.59→0.61, x/text 0.42→0.43 (indirect).
- **Verdict delivered:** no release needed on strict reading; recommended optional **v0.3.2 PATCH** mainly to live-fire `release.yml` (written after v0.3.1, never executed) with low stakes, dep bumps riding along.

**My error:** I reported "Toolchain floor: 1.26.7 → 1.27.1" — that conflated `.go-version` (local asdf/mise pin, invisible to consumers) with the go.mod `go` directive (the actual consumer floor, which was `go 1.27`, unchanged).

### Exchange 2 — operator challenge: "not 1.27?"

Verified the facts, then the tripwire:

- go.mod: `go 1.27`; `.go-version`: `1.27.1`; tripwire contract = **identical version string** across go.mod / `.go-version` / `.golangci.yml` `run.go` (go.mod authoritative; CI builds via `go-version-file: go.mod`).
- `bash scripts/check-go-version.sh` → **FAIL, exit 1** on both non-go.mod declarations. The `.golangci.yml` failure is structural: awk extracts the raw token `"1.27"` **including quotes**, so a quoted value can never match — and quoting was introduced because unquoted 2-part YAML (`go: 1.27`) parses as a *float*, which golangci rejects. A quoted 2-part value and an unquoted 3-part value are the only viable shapes; the repo had picked the unsatisfiable one.
- Git archaeology: v0.1.0 `go 1.26.5` (no `.go-version` yet), v0.2.0 `go 1.26.6`+`1.26.6`, v0.3.0 `go 1.26.7`+`1.26.7`, `.golangci.yml` `go: 1.26.7` unquoted at v0.3.0/v0.3.1. **Precedent: full patch version, identical, everywhere.**
- **Drift predates this session:** at the `v0.3.1` tag itself go.mod already said `go 1.27` while the other two said `1.26.7` → the tripwire was red at tag time. Post-v0.3.1 daemon commits (`cb31c8f`, `3d1c096`) moved `.go-version`→`1.27.1` and `.golangci.yml`→quoted `"1.27"` — still red, in a new way.
- **Fix:** go.mod → `go 1.27.1`; `.golangci.yml` → `go: 1.27.1` (unquoted — safe because 3-part `1.27.1` is a YAML *string*, satisfying both golangci and the raw-token tripwire).
- **Verification:** `check-go-version.sh` PASS (all agree on `1.27.1`); `env -u GOEXPERIMENT go build ./...` GREEN; `GOEXPERIMENT=jsonv2 go build ./...` GREEN.

### Report-time maintenance

- Added an AGENTS.md Gotchas bullet documenting the full-patch contract, the YAML-float/quote trap, and that v0.3.1 shipped red (memory mandate: update at moment of discovery — done late, at report time).

---

## a) FULLY DONE

| # | Item | Evidence |
| --- | --- | --- |
| a1 | Release-readiness assessment (go-release Phase 0) delivered with honest verdict: no release required; v0.3.2 optional | tags, commit census, file diff census |
| a2 | Go-version drift root-caused across 5 tags | go.mod/.go-version/.golangci values at v0.1.0–HEAD |
| a3 | Drift **fixed**: all three declarations now `1.27.1`, precedent restored | `check-go-version.sh` → PASS |
| a4 | Dual-mode build verification post-fix | `env -u GOEXPERIMENT` and `GOEXPERIMENT=jsonv2` builds green |
| a5 | Trap documented in project memory (AGENTS.md gotcha): YAML-float quoting vs raw-token tripwire compare | AGENTS.md Gotchas |
| a6 | Consumer-visible delta since v0.3.1 enumerated precisely (deps + toolchain floor only) | diff census |

## b) PARTIALLY DONE

| # | Item | Gap |
| --- | --- | --- |
| b1 **→ TODO_LIST #14** | v0.3.2 release prep — assessed and recommended, **not executed** (awaits operator) | no CHANGELOG section, no tag, no full Phase 4 gates |
| ~~b2~~ done — 2026-10-16 — full dual-mode test+race, lint both modes, all four tripwires green | ~~Verification of the fix — tripwire + builds only~~ | ~~full test suite, race, lint both modes, remaining 3 tripwires (`nix run .#tripwires`) NOT run this session~~ |
| ~~b3~~ done — 2026-10-16 (TODO_LIST/ROADMAP/CHANGELOG updated; commit still harness-gated) | ~~Memory maintenance — AGENTS.md updated~~ | ~~TODO_LIST not yet updated with new items (HARVEST pending operator go-ahead); no descriptive commit~~ |
| b4 **→ TODO_LIST #16** | Release-workflow confidence — `release.yml` inspected via CHANGELOG/diff only | never executed, no dry-run attempted |

## c) NOT STARTED

| # | Item | Note |
| --- | --- | --- |
| c1 **→ TODO_LIST #14** | v0.3.2 tag/push/release itself | operator decision (see g-Q1) |
| ~~c2~~ done — 2026-10-16 — CI inspected: lint legs RED (golangci v2.12.2 panic on go1.27.1 stdlib); pin bumped v2.14.0; green pending push | ~~Post-fix CI greenness confirmation (`gh run list`)~~ | ~~fix landed via daemon; CI state unverified this session~~ |
| c3 **→ TODO_LIST #14 (consumer sweep rides on the release)** | Consumer sweep (go-ecosystem-upgrade): go-cqrs-lite codec/v4 → next tag | demand side, blocked on a release existing |
| c4 **→ TODO_LIST #8** | Backfill GitHub Releases for v0.1.0/v0.2.0 | TODO_LIST #8, BLOCKED on operator preference |
| c5 **→ TODO_LIST #2** | Filing erraudit drafts upstream | TODO_LIST #2, BLOCKED on explicit operator instruction |
| c6 **→ TODO_LIST #1** | `ERRAUDIT_PAT` secret decision | TODO_LIST #1, BLOCKED (deferred 2026-09-11) |
| c7 **→ TODO_LIST #3** | Auto-commit daemon policy decision | TODO_LIST #3, operator-only |
| c8 **→ TODO_LIST #10** | CONTRIBUTING/AGENTS secrets note for contributors | TODO_LIST #10 |

## d) TOTALLY FUCKED UP

| # | Item | Severity |
| --- | --- | --- |
| d1 | **My first verdict was issued without running any verification gate.** The tripwire was one command away (`nix run .#tripwires`) and it was RED. I blessed release-readiness on top of a failing verification state. Operator caught the symptom (floor claim); the disease (red gate since ≤ v0.3.1) was mine to find first. | High — process |
| d2 | **v0.3.1 shipped with a red go-version gate.** go.mod said `go 1.27` while `.go-version`/`.golangci.yml` said `1.26.7` **at tag time**. Pre-existing (not this session's doing), but it means the release process released anyway — the gate did not stop anything, which raises the question whether it gates CI at all. | High — systemic |
| d3 | **`.golangci.yml` was edited into an unsatisfiable state** — quoted `go: "1.27"` can never pass the tripwire's raw-token compare regardless of version. Someone "fixed" a YAML-float problem by creating a tripwire-dead config. | Medium |
| d4 | **26/26 commits since v0.3.1 are unreadable auto-commit noise.** The tripwire fix from this session already landed as `chore:` noise too. Project history narrative exists only in docs/status files. Recorded gotcha, still biting. | Medium — recurring |

## e) WHAT WE SHOULD IMPROVE (brutal self-review)

**What did I forget?**
- To run the gates **before** answering a release-readiness question. Phase 0 of the release skill says assess; it doesn't excuse skipping the repo's own one-command health check.
- Memory maintenance at moment of discovery (AGENTS.md bullet written at report time, a session late).
- That `check-go-version.sh` runs in CI (ci.yml:42) — a failing tripwire means a red CI, which I never checked (`gh run list`) before or after the fix.

**What is stupid that we do anyway?**
- Version strings maintained by hand in three files with an exact-match tripwire that *detects* drift but prevented none — and whose own config format (quoting) can defeat it.
- Releasing (v0.3.1) while a verification gate is red, because nothing makes red gates release-blocking.
- Auto-commit daemon scooping every human-meaningful change into `chore:` noise (TODO_LIST #3, still open).

**What could I have done better?**
- Verify-then-claim: gates first, verdict second. The correct first answer was "let me run the tripwires" not a verdict.
- Precision: distinguish go.mod consumer floor from `.go-version` tooling pin in the first answer, so the operator didn't have to catch it.

**Did I lie?** No. I stated one wrong fact (floor claim), corrected same session after pushback; verdicts otherwise were accurate and explicitly hedged.

**Ghost systems?** None created. `release.yml` is not a ghost but is **unexercised-in-anger** tooling; first real tag will be its first real run.

**Split brains?** The session's core finding *is* a split brain: the same version concept maintained in three places that drifted (and README claimed 1.27.1+ while go.mod said 1.27 — README was more honest than go.mod). Now aligned; the structural risk remains until sync is generated rather than tripped.

**Scope creep?** None: no HARVEST, no release execution, no CI research beyond session scope. Operator said report-then-wait; obeyed.

**Removed something useful?** Only the quoted `"1.27"` YAML value — which was new, broken, and unusable under the tripwire. Nothing useful lost.

**Tests?** Version-only edits: tripwire + dual-mode builds are proportionate; full suite deliberately not run this session. Honest gap listed in b2.

**How to be less stupid (structural):**
1. Make "gates green" a mandatory line item in any readiness verdict template.
2. Generate the three version declarations from one source instead of tripwiring after the fact (see f-#18).
3. Make the tripwire quote-proof and format-validating (see f-#13/14).

---

## f) UP TO 50 THINGS TO GET DONE NEXT (impact-ordered; brainstorm, not commitment)

### Release thread — v0.3.2

| # | Task | Impact | Effort |
| --- | --- | --- | --- |
| 1 **→ TODO_LIST #14** | Operator decision: tag v0.3.2 now (live-fire `release.yml`, low stakes) vs bundle with next code change (g-Q1) | High | 0 |
| ~~2~~ done — 2026-10-16 (dual-mode race, lint both modes, all four tripwires) | ~~Full go-release Phase 4 pre-tag verification: dual-mode `go test -race`, lint both modes, `nix run .#tripwires`~~ | ~~High~~ | ~~30min~~ |
| ~~3~~ done — 2026-10-16 (no replace directives; go mod verify: all modules verified) | ~~Phase 3 hygiene: `grep '^replace' go.mod` leak check; `go mod tidy` + `go mod verify` both modes~~ | ~~High~~ | ~~10min~~ |
| 4 **→ TODO_LIST #14** | Cut CHANGELOG `[Unreleased]` → `[0.3.2]` (dep bumps, toolchain floor 1.27.1, tripwire fix, release.yml, dependabot, least-privilege CI, bench scripts, tripwires app) | High | 20min |
| 5 **→ TODO_LIST #14** | Leave `[Unreleased]` with placeholder categories | Low | 2min |
| 6 **→ TODO_LIST #14** | Curate user-focused GitHub Release notes (no breaking changes — say so) | Med | 10min |
| 7 **→ TODO_LIST #14** | Annotated tag v0.3.2, push, watch `release.yml` **first-ever live run** end-to-end | High | 15min |
| 8 **→ TODO_LIST #14** | Verify GitHub Release created with correct notes, `--latest` semantics | Med | 5min |
| 9 **→ TODO_LIST #14** | Verify proxy propagation: pkg.go.dev rendering + `go get …@v0.3.2` from a scratch module | High | 10min |
| 10 **→ TODO_LIST #14** | Consumer sweep (go-ecosystem-upgrade): bump go-cqrs-lite codec/v4 to v0.3.2, prove with `GOWORK=off` build | High | 30min |
| 11 **→ TODO_LIST #8** | Operator decision: backfill Releases for v0.1.0/v0.2.0 (TODO #8) | Low | 5min |

### Tripwire / verification hardening (this session's lesson)

| # | Task | Impact | Effort |
| --- | --- | --- | --- |
| ~~12~~ done — 2026-10-16 (implemented + tested incl. quoted-value shapes) | ~~Quote-proof `check-go-version.sh` (`tr -d '"'` on extracted tokens) so quoting style can't defeat exact-match~~ | ~~High~~ | ~~10min~~ |
| ~~13~~ done — 2026-10-16 (implemented + tested with explanatory error) | ~~Reject 2-part versions in the tripwire with an explanatory error (document the YAML-float trap where it bites)~~ | ~~Med~~ | ~~10min~~ |
| ~~14~~ done — 2026-10-16 (run step defaults to bash -e; tripwire exit 1 fails the job) | ~~Verify ci.yml's tripwire step actually **fails the job** (gating vs advisory) and that branch protection enforces it~~ | ~~High~~ | ~~15min~~ |
| ~~15~~ done — 2026-10-16 — CI inspected: lint legs RED, root-caused, pin bumped v2.14.0; green pending push | ~~Confirm CI green after the fix (was red on this gate pre-fix; `gh run list` / fresh run)~~ | ~~High~~ | ~~5min~~ |
| ~~16~~ done — 2026-10-16 (CHANGELOG Fixed + AGENTS gotcha + this archive annotation) | ~~Annotate (docs-health ANNOTATE, never rewrite) that v0.3.1 shipped with the gate red~~ | ~~Med~~ | ~~10min~~ |
| 17 **→ ROADMAP theme 5 (single-source version generation)** | Consider single-source version generation (one input → writes all three files) replacing detect-after-drift | Med | 1-2h |
| ~~18~~ done — 2026-10-16 (all four scripts PASS) | ~~Run full `nix run .#tripwires` — only check-go-version ran this session; error-codes / features-planned / erraudit-version unexercised~~ | ~~Med~~ | ~~5min~~ |
| 19 **→ TODO_LIST #16** | Consider `workflow_dispatch` dry-run mode for `release.yml` so the first real tag isn't its first run at all | Med | 30min |
| ~~20~~ done — 2026-10-16 (dependabot.yml covers gomod + github-actions) | ~~Check whether dependabot covers the `go modules` ecosystem (only `github-actions` entry seen)~~ | ~~Low~~ | ~~10min~~ |

### Docs / memory

| # | Task | Impact | Effort |
| --- | --- | --- | --- |
| ~~21~~ done — already present (README "Go 1.27.1+ (see go.mod)", verified this pass) | ~~README: one explicit toolchain-floor sentence (1.27.1 floor; GOTOOLCHAIN auto-switch absorbs patch bumps)~~ | ~~Med~~ | ~~5min~~ |
| ~~22~~ done — 2026-10-16 (CHANGELOG Fixed entries) | ~~If release delayed: add tripwire-fix entry under CHANGELOG `[Unreleased]` now~~ | ~~Low~~ | ~~5min~~ |
| ~~23~~ done — 2026-10-16 (this harvest: TODO_LIST rows 14-23, ROADMAP themes 2/5/7) | ~~docs-health HARVEST: route this report's section (f) into TODO_LIST.md / ROADMAP.md~~ | ~~High~~ | ~~20min~~ |
| 24 **→ TODO_LIST #18** | Consider an ADR line on the go-version policy (full-patch floor in go.mod is a consumer commitment — make it written policy) | Med | 20min |

### Commit hygiene

| # | Task | Impact | Effort |
| --- | --- | --- | --- |
| ~~25~~ **Won't implement — harness forbids commits without explicit operator instruction.** | ~~Hand-made descriptive commit for the tripwire fix + AGENTS.md bullet + this report (daemon already landed them as chore-noise; needs operator commit instruction, g-Q3)~~ | ~~Med~~ | ~~5min~~ |

### Pre-existing backlog (unchanged this session, from TODO_LIST)

| # | Task | Impact | Effort |
| --- | --- | --- | --- |
| 26 **→ TODO_LIST #1** | `ERRAUDIT_PAT` secret: PAT vs publish-erraudit decision (TODO #1, deferred 2026-09-11) | High | operator |
| 27 **→ TODO_LIST #2** | File the two prepared erraudit drafts upstream (TODO #2, verified v0.6.1-6-g7773253) | Low | 5min + instruction |
| 28 **→ TODO_LIST #3** | Auto-commit daemon policy: scoops, go.work writes, mid-session wipes (TODO #3) | Med | operator |
| 29 **→ TODO_LIST #10** | CONTRIBUTING/AGENTS note on PAT-style secrets for contributors (TODO #10) | Low | 10min |
| 30 **→ TODO_LIST #8** | Decide v0.1.0/v0.2.0 Release backfill (TODO #8) | Low | 5min |

### Worth considering (ROADMAP fuel)

| # | Task | Impact | Effort |
| --- | --- | --- | --- |
| ~~31~~ **Won't implement — declined (the dual-mode lint legs already catch .golangci.yml build-tags drift).** | ~~Tripwire for `.golangci.yml` build-tags drift vs the dual-mode docs~~ | ~~Low~~ | ~~30min~~ |
| 32 **→ TODO_LIST #19** | Non-nix contributor path for the four check scripts (document outside flake) | Low | 20min |
| ~~33~~ done — adopted (gh run list checked as this pass's step 0) | ~~Habit: `gh run list` glance at session start of any release/verification work~~ | ~~Med~~ | ~~1min~~ |
| ~~34~~ **Won't implement — declined (AGENTS.md gotcha already carries it).** | ~~Reminder bullet in release checklist: quiet-window gate before any benchmark validation (AGENTS.md gotcha)~~ | ~~Med~~ | ~~—~~ |
| 35 **→ TODO_LIST #14 (adoption-proof re-verify rides on the release)** | After next release: verify go-cqrs-lite adoption proof still holds against the new tag (AGENTS.md claims proxy-build with `GOWORK=off`) | Med | 15min |
| 36 **→ ROADMAP theme 5 (job-summary idea)** | Consider CI job-summary step posting gate statuses for at-a-glance red detection | Low | 30min |
| ~~37~~ done — 2026-10-16 (all 8 open reports annotated; 13-file cell-strike debt routed TODO_LIST #21) | ~~Sweep docs/status archive: annotate stale claims older reports made about go-version state~~ | ~~Low~~ | ~~30min~~ |
| ~~38~~ done — adopted (gates-first applied this pass) | ~~Add "run repo gates as step 0" to the release-assessment workflow notes (skill or checklist)~~ | ~~Med~~ | ~~10min~~ |
| 39 **→ ROADMAP theme 5 (folded into single-source version generation)** | Decide whether `.go-version` should gain a `go.mod`-sync check in the flake (tripwire currently bash-only in CI) | Low | 20min |
| 40 **→ TODO_LIST #14 (stance folded into the v0.3.2 row)** | Eventually: retract-or-tolerate stance for v0.3.1 (shipped with red gate; content is fine — likely tolerate, document) | Low | 10min |

---

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Tag v0.3.2 now, or wait?** My recommendation stands (live-fire the never-run `release.yml` with low stakes, dep bumps riding along) — but tags are immutable once the proxy sees them, so timing is your call, not mine. **→ TODO_LIST #14**
2. **Ratify the go-version policy I just re-established?** All three files carry the identical **full patch** version (`1.27.1`) — consumer floor is honestly 1.27.1, precedent-consistent. The alternative is a loose `go 1.27` floor + a major.minor-only tripwire compare. I chose full-patch; confirm or override, because every future bump inherits this decision. **→ TODO_LIST #18 (full-patch contract stands; written policy pending)**
~~3. **Commit instruction?** The tripwire fix, the AGENTS.md bullet, and this report landed via the daemon as `chore:` noise — the exact history-flattening gotcha AGENTS.md warns about. Want me to hand-make one descriptive commit for them (I don't commit without your explicit say-so)?~~ **Won't implement — harness-gated (operator commit instruction).**

---

## Verification log (this session)

| Command | Result |
| --- | --- |
| `bash scripts/check-go-version.sh` (pre-fix) | FAIL exit 1 — `.go-version` 1.27.1 vs go.mod 1.27; `.golangci.yml` `"1.27"` vs 1.27 |
| `bash scripts/check-go-version.sh` (post-fix) | PASS — all declarations agree on 1.27.1 |
| `env -u GOEXPERIMENT go build ./...` | GREEN |
| `GOEXPERIMENT=jsonv2 go build ./...` | GREEN |
| `git status` (end of session) | clean — daemon committed the fix |

*Not run this session (deliberately, scope): full test suite, race, lint, the other three tripwires, CI run inspection.*

---

**NEXT: WAITING FOR INSTRUCTIONS.** Most natural next moves: answer g-Q1..3, or say "harvest" to route section (f) into TODO_LIST.
