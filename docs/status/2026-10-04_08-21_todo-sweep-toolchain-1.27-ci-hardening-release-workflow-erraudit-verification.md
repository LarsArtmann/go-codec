# Status Report — TODO sweep: Go 1.27 toolchain bump, CI hardening, release workflow, erraudit verification

> Point-in-time snapshot, 2026-10-04 08:21. Session executed open TODO_LIST items
> (research-first, one at a time). Auto-commit daemon has already scooped all
> working-tree changes into `3d1c096`/`8baae09` (expected noise; item 3 policy
> still open).

## a) FULLY DONE

1. **Research & verification foundation.** TODO harvest sources read (three
   archived status reports, all tripwire scripts, flake.nix, ci.yml,
   dependabot.yml, benchmark baseline doc). Latest Go verified via go.dev/dl
   JSON: **go1.27.1** stable (go1.26.8 is the 1.26 line). Locked nixpkgs
   (`c59305ba`, 2026-10-01) already packages `go_1_27` = 1.27.1 — no nixpkgs
   bump needed. Local erraudit checkout found at `v0.6.1-6-g7773253`.
2. **Item 12 — deliberate Go 1.27.1 toolchain bump (declaration moves).**
   Session-start discovery: the daemon had AGAIN silently bumped go.mod to
   `go 1.27` (`2e70454`, Sep 29) — tripwire RED. Completed the bump properly:
   go.mod `1.27.1` (via `go mod edit`), `.go-version` `1.27.1`,
   `.golangci.yml` `go: 1.27.1`, flake.nix `go_1_27`. All four declarations
   agree (`check-go-version.sh` PASS). Both JSON modes build; **full test
   suite green on go1.27.1** in the devShell; `GOEXPERIMENT=jsonv2` still
   accepted under 1.27.1.
3. **Item 5 — CI hygiene.** Verified at execution time via `gh api`: all five
   pinned actions (checkout v7.0.1, setup-go v7.0.0, upload-artifact v7.0.1,
   setup-node v7.0.0, gitleaks-action v3.0.0) are the **latest releases** and
   already run **node24** (`using:` read from each action.yml at the pinned
   SHA) — the digest-refresh half was already satisfied; documented instead of
   churned. Added least-privilege top-level `permissions: contents: read` to
   ci.yml.
4. **Item 6 — dependabot + caching.** `github-actions` ecosystem entry added
   (weekly, grouped, labeled). `cache: true` pinned explicitly on all six
   setup-go steps (v7 default already `true` — verified; now intent-pinned).
5. **Item 11 — `.#tripwires` nix app.** Runs the four check scripts in CI-leg
   order; executed green locally (`nix run .#tripwires`).
6. **Item 9 — tag-push release workflow** (`.github/workflows/release.yml`).
   Codifies the manual v0.3.0 runbook: scoped `contents: write`; release gates
   (replace/pseudo-version leak checks, `go mod verify`, build + test both
   JSON modes); notes cut from the CHANGELOG section for the tag (v-prefix
   tolerant) with `--generate-notes` fallback; idempotent create-or-refresh;
   `--latest` for stable tags, `--prerelease` for rc/alpha/beta (repo
   convention: 0.x ships as full Latest, per v0.3.0).
7. **Item 2 — upstream erraudit claims re-verified with live evidence**
   (against local HEAD v0.6.1+6, supersedes the 2026-09-11 v0.4.0 check):
   - (a) `internal/ast/predicates.go:95-127` — `createsErrorInternally`
     recognizes `errors.New`/`errors.Join`/`fmt.Errorf` and **exact-name**
     `Wrap`/`Wrapf` selector calls only → go-error-family creators
     `WrapOncef`/`WrapCorruptionf`/`WrapInfrastructuref`/`WrapOrchestrationf`
     are invisible.
   - (b) `--enforce-go-error-family`/`--enforce-samber-oops` are boolean
     pass-throughs (`cmd/erraudit/cmd/appconfig.go:35-36`,
     `internal/portfolio/audit_args.go:59-63`) — no go.mod dependency check.
   - Live repro (`/tmp/erraudit-repro`): `controlCreator` (errors.New) flagged,
     exit=2; `wrapOncefCreator` (WrapOncef) silently NOT flagged.

## b) PARTIALLY DONE

1. **Item 2 — issue drafts.** Verification complete; github-voice skill loaded
   (actual location `/home/lars/projects/SKILLS/github-voice`). The two draft
   bodies are **not yet written** to `docs/drafts/` — evidence is ready, this
   is the immediate next step. Filing stays gated on explicit instruction.
2. **Item 12 — hermetic tail.** `nix build` / `nix flake check` not yet re-run.
   The flake `vendorHash` predates the daemon's Sep-29 dependency bumps
   (gomega 1.43→1.44 + indirects), so a hash mismatch is **expected** and
   needs the `got:` hash pasted in.
3. **Final verification sweep** not yet run: actionlint (new release.yml,
   edited ci.yml), golangci-lint both modes, `go test -race` both modes,
   `nix fmt` drift check.

## c) NOT STARTED

1. **Items 4+7 — benchmark re-baseline.** Deliberately sequenced after the
   toolchain settle (a baseline on the retired toolchain would be instantly
   stale); nothing run yet. Includes: 10-count runs v1+v2 on go1.27.1,
   benchstat, recreating `scripts/bench-compare.py` (raw-vs-raw, per-metric,
   0.33–3 sanity bounds), the raw-output retention decision, and the
   `docs/benchmark-baseline.md` rewrite.
2. **Docs sync.** TODO_LIST.md (delete 5/6/9/11/12, annotate 2/4/7),
   CHANGELOG.md Unreleased entries, AGENTS.md (`.#tripwires` command, Go 1.27
   references), README Go floor line (1.26.7+ → 1.27.1+), SECURITY.md check.
3. **Item 10** — gated on item 1's direction (deferred user decision).
4. **Blocked items untouched:** 1 (ERRAUDIT_PAT), 3 (daemon policy),
   8 (release backfill).

## d) TOTALLY FUCKED UP!

Nothing destructive. Honest waste and mistakes:

1. **rg flag misuse twice** (`rg -rn`, `rg -rln` — `-r` is *replace*, not
   recursive): mangled output; one accidental broadened search over `/mnt` and
   the module cache burned a large chunk of context.
2. **Context-dumping fetch:** the raw proxy.golang.org toolchain list (~100KB)
   was fetched whole to answer "latest 1.27 patch?" — should have been a
   filtered/agentic query.
3. **Critical-path sequencing:** the benchmark re-run (longest task, ~45min)
   still has not started because CPU-heavy verification (erraudit builds,
   runs) was front-loaded; benchmarks must now wait for the remaining heavy
   verification or accept noise.
4. **(Pre-existing, restated because it recurred):** the daemon's second
   silent go.mod bump (`2e70454`), same failure mode as `ed85a26`. No guard
   exists between daemon and go.mod.
5. **Side-finding:** `github-voice` is advertised in available_skills but its
   fan-out symlink `~/.config/crush/skills/github-voice` is **dangling**
   (actual skill at `/home/lars/projects/SKILLS/github-voice`). The
   session-start fan-out guard evidently did not catch it.

## e) WHAT WE SHOULD IMPROVE!

1. **Pre-commit tripwire:** hook `check-go-version.sh` (and friends) into
   pre-commit so daemon/user commits fail fast instead of leaving red CI +
   session-start repair. Two silent go.mod bumps in two weeks.
2. **CHANGELOG drift:** tag `v0.3.1` exists with **no matching CHANGELOG
   section** (Unreleased jumps to [v0.3.0]). release.yml would fall back to
   generated notes for it. Reconcile.
3. **Tag-push duplication:** ci.yml and release.yml both fire gates on tag
   push. Accepted for now (documented in workflow header); could be slimmed
   into one workflow later.
4. **Fan-out integrity:** run `bash ~/.config/crush/scripts/check-skill-fanout.sh`
   after this session; fix the dangling github-voice link.
5. **rg discipline:** `-r` replaces. Use `-l`/`-n` separately.
6. **Reproducible bench tooling:** provision benchstat via pinned nixpkgs or a
   dedicated GOBIN, not ad-hoc installs.

## f) Next tasks

1. Write `docs/drafts/erraudit-issue-generic-return-creators.md` (repro + file:line evidence ready)
2. Write `docs/drafts/erraudit-issue-enforce-missing-library.md`
3. Run `check-draft.py` over both drafts (github-voice mechanical gate)
4. Provision benchstat reproducibly
5. Run v1 10-count benchmark suite (quiet machine, background, go1.27.1)
6. Run v2 10-count benchmark suite
7. Recreate `scripts/bench-compare.py` (raw-vs-raw, per-metric, sanity bounds 0.33–3)
8. Validate bench-compare.py against the two fresh raw outputs
9. Rewrite `docs/benchmark-baseline.md` (v1+v2 sections, go1.27.1, re-date, retention policy)
10. `nix build` → paste `got:` vendorHash into flake.nix
11. `nix flake check` hermetic green
12. actionlint over `.github/workflows/*.yml`
13. golangci-lint both modes
14. `go test -race` both modes
15. `nix fmt` + zero drift
16. TODO_LIST.md: delete items 5, 6, 9, 11, 12
17. TODO_LIST.md: annotate items 2 (drafts prepared), 4/7 (in flight)
18. CHANGELOG.md Unreleased entries for everything shipped this session
19. AGENTS.md: `.#tripwires` command; Go 1.27 references
20. README Go floor line → 1.27.1+
21. Check SECURITY.md for stale version mentions
22. lsp diagnostics on touched files
23. Hand-made descriptive commit for the toolchain bump before any push
24. Reconcile v0.3.1 CHANGELOG section (drift)
25. File both erraudit issues upstream on explicit instruction
26. Consider erraudit pin bump v0.4.0 → v0.6.x (both pin sites + tripwire)
27. BLOCKED: ERRAUDIT_PAT vs publish decision (item 1)
28. BLOCKED: daemon policy incl. go.mod/go.work exclusions (item 3)
29. BLOCKED: v0.1.0/v0.2.0 release backfill preference (item 8)
30. GATED: CONTRIBUTING/AGENTS PAT-secrets note (item 10)
31. Run check-skill-fanout.sh; repair the dangling github-voice symlink
32. ROADMAP: CI bench job with artifact retention (per retention decision)
33. ROADMAP: release checklist script (09-17 §f-48) — mostly covered by release.yml; keep as candidate

## g) Questions I can NOT figure out myself

1. **Item 1/10 direction:** store an `ERRAUDIT_PAT` secret (CI gate
   self-activates) or publish the erraudit repo? Deferred 2026-09-11; gates
   both the CI error-audit gate and the contributor docs note.
2. **Item 8:** backfill GitHub Releases for v0.1.0/v0.2.0 (v0.3.0 treatment
   retroactively) or leave tags-only?
3. **Item 3:** daemon commit policy — one-commit scoops vs split commits, and
   should go.mod/go.work writes be excluded entirely? Two silent go.mod bumps
   have now occurred (`ed85a26`, `2e70454`).
