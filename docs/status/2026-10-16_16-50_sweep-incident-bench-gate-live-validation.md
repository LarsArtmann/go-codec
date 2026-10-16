# Status — Sweep Incident Response & Bench-Gate Live Validation (2026-10-16 16:50 CEST)

**Verdict: MASTER GREEN AND HARDENED. Both incidents closed with root-cause
fixes, the self-relative benchmark gate is validated end-to-end, and
pkg.go.dev serves v0.3.2. Session grew out of the 08:25 report's addendum.**

| Quick stats |                                                                |
| ----------- | -------------------------------------------------------------- |
| Incidents   | 2 (version-contract sweep ×2 strikes, bench-gate silent no-op) |
| Fixes       | `d543914` + re-restore (via `1128d12`), bench-gate `1128d12`   |
| Bench gate  | COMPARES live: 210/210 verdicts OK (run 38915649152)           |
| pkg.go.dev  | v0.3.2 indexed and served (was 404 this morning)               |
| Master      | `70d7040` pushed, all runs green, tripwires 5/5                |

## a) Incident 1 — automated sweep broke the version contract (twice)

- 08:19 (`397b3b2`, daemon-pushed): go.mod/.golangci.yml downgraded to
  2-part `1.27`, unreviewed `lll` linter added, pinned actions/cache bumped
  v4.3.0 → v6.1.0. Master red on every push (go-version tripwire, as
  designed — a 2-part value parses as a YAML float, exact match becomes
  unsatisfiable). Restored in `d543914`; formatting kept.
- 12:54 (`940912b`, daemon commit, caught UNPUSHED this time): same profile
  re-applied. Pre-commit tripwire blocked the next human commit on the
  broken tree. Recovery: sanctioned soft-reset of unpushed daemon commits,
  re-restore, descriptive re-commit.
- **Root cause identified:** `projects-management-automation … buildflow
  --fix` loop (PID seen live) running under the ambient go1.27.0 toolchain
  "aligns" the declarations downward. Cyclic — will strike again.
- **Operator action needed (TODO_LIST #3 adjacent):** point that loop at
  go1.27.2 (or exclude go.mod/.golangci.yml/.go-version from its fix
  profile). Until then every strike is caught by tripwire + pre-commit, but
  each costs a red-CI window or a manual restore.

## b) Incident 2 — bench gate compared nothing (and now does)

- The v0.3.2-era self-relative benchmark gate called
  `gh run download -O`; **no such flag exists**. The swallowed exit 1 made
  every candidate run look artifact-less, so every run silently re-seeded
  and the gate gated nothing — while staying green.
- Exposed by dispatching a run instead of waiting for Sunday's cron
  (38864642165: "no previous bench artifact found" despite the seeded
  artifact being present). Fixed in `1128d12`: extract the artifact into a
  temp dir and take `bench-current.txt`; lookback widened to 30 runs so
  docs-only pushes cannot scroll the baseline out.
- **Live validation (run 38915649152):** `baseline: artifact bench-results
  from run 38864642165` → `bench-compare.txt` with **210 comparisons, all
  OK, zero regressions** (parity on identical code), artifact now ships
  both files (7,906 → 10,105 bytes). Sunday's cron compares against this.

## c) Drift fixed on sight

- CHANGELOG compare links: v0.3.1/v0.3.2 refs were missing entirely and
  [Unreleased] still diffed v0.3.0 (`faab46c`).
- README dual-build floor "1.27.1+" → "1.27.2+" on master (the shipped tag
  keeps the old sentence; next release picks the fix up).
- AGENTS.md gotchas: sweep recurrence + culprit + daemon-amend workflow;
  unknown-CLI-flag lesson (green CI proves only what steps can fail on).

## d) State & owed items after this session

- go-codec master: clean, pushed, every gate green (38864602942, 38864846026,
  38865258605, 38915648800, 38915649152, 38915753690 — all success).
- pkg.go.dev `@v0.3.2`: **live** (indexing lag resolved itself; proxy was
  never in question).
- go-cqrs-lite: fully pushed; a parallel agent's error-taxonomy and
  versions.json/README work observed and left untouched.
- Still operator-gated: ERRAUDIT_PAT (M27/TODO #1), erraudit drafts (M17),
  daemon policy (M11/TODO #3 — now including the buildflow-loop toolchain),
  sibling-push of future trains (TODO #24), M24 builders, M25 LTX.

_Verification sources: fresh CLI runs this session — five tripwires
(pre-commit + `nix run .#tripwires`), dual-mode builds on 1.27.2, actionlint,
local artifact-structure + parity-compare simulation, `gh run view` logs for
every CI claim (38860891898, 38862596496, 38864602942, 38864642165,
38915649152 and the push runs listed above), the bench-results artifact
download, proxy @v/list, and pkg.go.dev fetches at 09:00 (404) and 16:45
(v0.3.2 live). No local benchmarks were run._
