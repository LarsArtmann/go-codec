# TODO List

> Short-term, actionable, bounded work items. For long-term vision and unrefined
> ideas, see `ROADMAP.md`. Completed items are deleted from this list and logged
> in `CHANGELOG.md`.

## Status legend

| Status           | Meaning                                                 |
| ---------------- | ------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                               |
| 🟡 `IN_PROGRESS` | Actively being worked on.                               |
| 🔵 `BLOCKED`     | Cannot proceed; external dependency or decision needed. |

## Active items

| # | Task | Status | Impact | Effort | Notes |
| - | ---- | ------ | ------ | ------ | ----- |
| 1 | Decide release strategy and create the GitHub Release (`gh release create`) | 🔵 `BLOCKED` | High | 5min | User decision 2026-09-11: release is DEFERRED until the error-contract follow-up batch (this list) is done. The batch sits in `CHANGELOG.md [Unreleased]` (31 codes, family taxonomy, message-format change). Post-batch runbook: tag → `gh release create` with the `[Unreleased]` body → re-date to `## [v0.2.1]` → verify `go get` + pkg.go.dev → bump `go-cqrs-lite/codec/v4` + run its suite `GOWORK=off`. Gate: CI fully green, incl. the `error-audit` job (see #2). |
| 2 | Add `ERRAUDIT_PAT` secret so the CI `error-audit` gate activates | 🔵 `BLOCKED` | High | 5min | erraudit is a PRIVATE module: proxy.golang.org 404s (verified cold-cache 2026-09-11), so CI needs a fine-grained PAT (read: `github.com/larsartmann/erraudit`) stored as `ERRAUDIT_PAT`. The job in `.github/workflows/ci.yml` self-activates when the secret exists (skips with reason until then). Alternative: publish the erraudit repo (user call). Second blocker already handled in-job: erraudit v0.4.0 imports `encoding/json/v2`, so install needs `GOEXPERIMENT=jsonv2` under the go1.26.7 toolchain. Source: `docs/status/2026-09-11_06-45_….md` §f-4/6. |
| 3 | Create `docs/error-codes.md`: the error-code registry | 🔴 `TODO` | High | 45min | All 31 codes (list in `CHANGELOG.md [Unreleased]`), each with family, meaning, and emitted site (`file:line`). The machine vocabulary has no single home; this feeds #4 and the ADR. Source: report §f-11. |
| 4 | Tripwire: all `codec.*` error codes unique and registered in `docs/error-codes.md` | 🔴 `TODO` | Med | 45min | Script or test in the spirit of `scripts/check-features-planned.sh` (runs in CI). Source: report §f-12. Depends on #3. |
| 5 | Property test (rapid): every error escaping the public API is `*errorfamily.Error` with a `codec.`-prefixed code | 🔴 `TODO` | High | 1h | Mechanically catches future unclassified wraps at the API boundary. Source: report §f-13. |
| 6 | Make wrapcheck `ignore-sigs` explicit for all `errorfamily.Wrap*` variants (or document the matching mechanism) | 🔴 `TODO` | Med | 30min | The migrated wraps pass lint today for unexplained reasons (pattern looser than the configured strings) — fragile config-by-luck. Source: report §e-5/§f-14. |
| 7 | godoc `Example*` demonstrating `errors.AsType[*errorfamily.Error]` on a codec error | 🔴 `TODO` | Med | 30min | Doubles as a test. Source: report §f-19. |
| 8 | ADR: error taxonomy (families per layer), WrapOnce rule, generic_return decline rationale | 🔴 `TODO` | Med | 1h | Input: 2026-09-11 diagnosis — the 11→7 `generic_return` drop was an analyzer false-negative (`createsErrorInternally` exact-matches `Wrap`/`Wrapf`, missing `WrapOncef`/`WrapCorruptionf`/`WrapInfrastructuref`); the decline rationale itself stands. Source: report §f-20/§b-1. |
| 9 | Add erraudit to the nix devShell (+ `.#erraudit` flake app for pinned local runs) | 🔴 `TODO` | Med | 45min | Local devs currently lack the binary hermetically. Note: erraudit needs `GOEXPERIMENT=jsonv2` under go 1.26 (verified 2026-09-11). Source: report §f-21/§c-8. |
| 10 | Pin erraudit version management (renovate/dependabot for the ci.yml `go install`) | 🔴 `TODO` | Low | 15min | Source: report §f-22. |
| 11 | Sync the error policy to sibling repos (go-cqrs-lite, signing, encryption, storage) | 🔴 `TODO` | High | 30min/repo | Add the same `error-audit` CI job (same `ERRAUDIT_PAT` mechanism, see #2) and zero their findings under `--enforce-go-error-family --type-aware`. Source: report §f-23/24/§c-2. |
| 12 | Upstream erraudit improvements (verify-before-filing applies) | 🔴 `TODO` | Med | 30min + review | (a) `generic_return` misses `WrapOncef`/`WrapCorruptionf`/`WrapInfrastructuref`/`WrapOrchestrationf` as error creators (verified 2026-09-11 against v0.4.0 source: exact-name match on `Wrap`/`Wrapf` only). (b) Warn when `--enforce-samber-oops`/`--enforce-go-error-family` names a library absent from go.mod (the trap that produced the misleading audit report). Source: report §e-6/§f-29/30. |
| 13 | Auto-commit daemon policy: one-commit scoops vs separate docs/code commits | 🔵 `BLOCKED` | Med | — | User decision (daemon config change only the user can make); the daemon preempted the commit-split question once already (`f04d158`) — `docs/status/2026-08-14_20-58_…md` §g-3. |
| 14 | CI check that renders the README mermaid architecture diagram (catches syntax drift) | 🔴 `TODO` | Low | S | Recovered lost item 2026-08-15: flagged as 2026-08-14_20-07 follow-up report item #36 but never shipped (verified absent from `.github/workflows/ci.yml`) and never routed to any living doc. README also carries an ASCII fallback; a mermaid-cli (npx/docker) render step would protect the primary diagram. |
