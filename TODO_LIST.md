# TODO List

> Short-term, actionable, bounded work items. For long-term vision and unrefined
> ideas, see `ROADMAP.md`. Completed items are deleted from this list and logged
> in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                      |
| ---------------- | ------------------------------------------------------------ |
| 🔴 `TODO`        | Not started. Needs doing.                                    |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                    |
| 🔵 `BLOCKED`     | Cannot proceed; external dependency or decision needed.      |

## Active items

| #  | Task                                                                                             | Status        | Impact | Effort | Notes                                                                                                                     |
| -- | ------------------------------------------------------------------------------------------------ | ------------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------- |
| 1  | Decide release strategy and create the GitHub Release (`gh release create`)                      | 🔵 `BLOCKED`  | High   | 5min   | Recommendation: cut `v0.1.1` from HEAD — `v0.1.0` tag at `3f8ac9d` predates the v1-build fix `d871122`. Moving a published tag poisons the module proxy. The two related open questions are resolved: (a) fuzz corpus is artifact-only, no auto-commit; (b) `DeterministicCodec` is documented in README and code for sibling signing integration. Requires user confirmation + remote push/tag. Post-release runbook: tag → `gh release create` with the `[Unreleased]` body → re-date `CHANGELOG.md [Unreleased]` to `## [0.1.1]` → verify `go get github.com/larsartmann/go-codec@v0.1.1` + pkg.go.dev rendering → rebuild `go-cqrs-lite/codec/v4` against the tag with `GOWORK=off`. |
| 2  | Watch the first real GitHub Actions run of the new CI additions (coverage summary, FEATURES drift tripwire, fuzz-matrix entries, weekly fuzz cron) once pushed; review the fuzz budget (13 targets × 30s × 2 modes) against actual runner time | 🔴 `TODO` | Med | S | New steps have never run on a real runner — YAML is valid, runner behavior unverified (`docs/status/2026-08-14_20-58_t11-completion-brutal-self-review.md` §b-3). Adjust cron length/budget from observations (same report §f-2, §f-14). |
| 3  | Record a benchstat long-run benchmark baseline (ns/op, B/op) in a docs file; upgrade FEATURES `~` figures to baseline-grade | 🔴 `TODO` | Med | 30min | FEATURES performance numbers are indicative (quick `-benchtime=200ms` runs), not regression-grade — `docs/status/2026-08-14_20-58_…md` §b-2, §f-7. CI bench-regression job itself stays a ROADMAP theme-2 idea. |
| 4  | Add JSON tags to `SizeResult` (`size.go`)                                                        | 🔴 `TODO`     | Low    | S      | API change, deferred deliberately; pair with the `v0.1.1` release decision — `docs/status/2026-08-14_20-58_…md` §f-13. |
| 5  | Resolve the unusedwrite smell at `observability_test.go:560` (`snap1.EncodeBytes = -1` is written but never read) | 🔴 `TODO` | Low | 10min | Assert the mutated snapshot field or drop the write — `docs/status/2026-08-14_20-58_…md` §f-8. Verified present this session. |
| 6  | Single source of truth for the Go version: `.go-version` vs `go.mod` vs CI `go-version-file`     | 🔴 `TODO`     | Low    | 10min  | Three declarations, one consumer (CI reads `go.mod` only; `.go-version` is `1.26.5`, matching). Point CI at `.go-version` or delete the file — `docs/status/2026-08-14_20-58_…md` §f-10. |
| 7  | Auto-commit daemon policy: one-commit scoops vs separate docs/code commits                      | 🔵 `BLOCKED`  | Med    | —      | User decision (daemon config change only the user can make); the daemon preempted the commit-split question once already (`f04d158`) — `docs/status/2026-08-14_20-58_…md` §g-3. |
| 8  | Add `gosec` security lint to CI                                                                 | 🔴 `TODO`     | Low    | 10min  | Flagged 2026-08-12 (`docs/status/2026-08-12_03-24_…md` §"CI / release hygiene" #10, per how-to-golang policy) but never shipped — `govulncheck` and `gitleaks` are in CI, `gosec` is not (verified absent from `.github/workflows/ci.yml` this session). |
