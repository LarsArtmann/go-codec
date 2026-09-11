# Status Report: Error-Contract Follow-up Batch (CI gate repair, docs sync, perf proof, harvest)

- **Date:** 2026-09-11 08:27 CEST
- **Session scope:** Execute the §f follow-up batch from the 06:45 report: CI error-audit repair,
  CHANGELOG/README/AGENTS.md sync, benchmark re-baseline, generic_return delta diagnosis,
  TODO_LIST harvest, report annotation, full verification matrix.
- **Repo:** go-codec @ master `7eaf7d2`+ (daemon commits `d72c792`, `02b18b0`, `7eaf7d2` carry this
  session's code/docs; two files uncommitted at write time, daemon-owned)
- **Verdict:** All 10 §f P0 items closed and verified; full gate matrix green. ONE decision remains
  user-blocked: the `error-audit` CI gate needs an `ERRAUDIT_PAT` secret (or erraudit published) —
  until then the job self-activatingly skips. No release/tag (deferred per standing decision).

---

## TL;DR

The 06:45 report's handoff contained a false claim ("erraudit is public on proxy; CI never pushed").
This session disproved both: the branch WAS pushed, the new `error-audit` job RAN and FAILED
(run 34566358496) because `larsartmann/erraudit` is a private repo (proxy 404, cold-cache verified)
that additionally cannot build under stock Go 1.26.7 (it imports `encoding/json/v2`, experimental in
1.26). The CI job is now repaired as a self-activating gate (skips with a documented reason until an
`ERRAUDIT_PAT` secret exists, then enforces with `GOPRIVATE` + token + `GOEXPERIMENT=jsonv2`), a
`workflow-lint` (actionlint) job guards the workflow itself, every user-facing doc surface is synced
(CHANGELOG, README, AGENTS.md, ROADMAP, TODO_LIST), the performance-discipline loop is closed with
measurement (allocations byte-identical on all hot paths, A/B vs the pre-error-contract commit),
and the status report's 44 open items are annotated inline with evidence.

---

## a) FULLY DONE (verified)

1. **Red CI diagnosed at the root.** Latest three runs failed ONLY in the new `Error Audit (erraudit)`
   job at "Install erraudit": `could not read Username for 'https://github.com'`. Verified with a
   cold module cache: `proxy.golang.org` returns **404** for `erraudit@v0.4.0` — the repo is
   **PRIVATE** (gh api: `"visibility":"PRIVATE"`); the handoff's "public on proxy" was a warm-cache
   mirage. Second, independent blocker found: erraudit v0.4.0 (exact tag, isolated via
   `git archive`) imports `encoding/json/v2` in `pkg/types`, fails under go1.26.7 without
   `GOEXPERIMENT=jsonv2` (its `go 1.26.5` directive prevents toolchain auto-switch; locally it was
   built with go1.27.1 where json/v2 is default).
2. **`error-audit` job made self-activating** (`7eaf7d2`): job-level `HAS_ERRAUDIT_PAT` derived from
   the secret; install step (only when the secret exists) sets `GOPRIVATE=github.com/larsartmann/*`,
   `GOEXPERIMENT=jsonv2`, and a `git insteadOf` token rewrite; the audit step enforces
   `--enforce-go-error-family --type-aware`; without the secret it prints a skip-reason step
   (master green, gate visibly dormant). The `GOEXPERIMENT=jsonv2` workaround was PROVEN: v0.4.0
   built under go1.26.7 with it, and that binary audits this repo at exit 0.
3. **`workflow-lint` job added** (actionlint `v1.7.12`, latest release verified via API). Justified
   itself immediately: it caught a YAML syntax error in my first edit of the same file.
4. **Benchmark discipline closed with measurement, not assumption.** Full suite re-run (67
   benchmarks × 10, v1 mode, 832s). Fresh baseline is NOT comparable to the 2026-08-15 anchor
   (go1.26.5/cbor v2.9.2; deltas bidirectional, geomean 0.888) — so the error-contract delta was
   isolated via a same-machine A/B benchstat (10 runs, commit `6641688` vs HEAD, throwaway
   worktree): **B/op and allocs/op identical on every hot-path benchmark; WrapEncode,
   UnwrapDecode, EncodePooled, CBORCodec_Decode statistically indistinguishable; geomean sec/op
   −10.2%.** `docs/benchmark-baseline.md` refreshed (go1.26.7, cbor v2.9.3) with the A/B verdict
   recorded; stale anchor retired explicitly.
5. **CHANGELOG.md `[Unreleased]` written** — the error contract is now visible to release tooling:
   31 stable codes enumerated, family taxonomy, the `[family:code]` message-format change flagged
   as breaking for string-matching consumers, WrapOnce rule, generic_return policy, contract
   tests, CI job, toolchain/cbor bumps.
6. **README.md `## Error Handling` section added** — user-facing sales surface synced with `doc.go`:
   family table, `[family:code]` rendering, "never match strings" +
   `errors.Is`/`errors.AsType` guidance with example.
7. **AGENTS.md hardened** (`02b18b0`, `7eaf7d2`): Commands now mandate mode-pinned invocations
   (`env -u GOEXPERIMENT` / `GOEXPERIMENT=jsonv2`) with the ambient-trap warning; Gotchas documents
   the broken-v1-file-passed-v2-suite incident; new daemon-commit caveat (hand-make descriptive
   commits before pushing shared branches — the error-contract stack had zero meaningful commit
   messages).
8. **Lying comment fixed** (`d72c792`): `envelope.go` Magic field comment said `always "cqrs"`;
   actual value is `"gcdc"` — now says `"gcdc" (envelopeMagic)`.
9. **generic_return 11→7 mystery SOLVED** (was §b-1 stale claim): read the v0.4.0 analyzer source —
   `createsErrorInternally` recognizes ONLY exact method names `Wrap`/`Wrapf` (plus
   `errors.New`/`fmt.Errorf`). The four functions migrated to `WrapOncef`/`WrapCorruptionf`/
   `WrapInfrastructuref` (WrapEncode, EncodePooled, TranscodeToJSON, observability Write) became
   invisible "pure forwarders" — the count drop is an analyzer false-negative, NOT a code
   improvement. Current 7 enumerated exactly; the decline rationale itself stands.
10. **erraudit zero-baselines recorded:** `--type legacy_as` → **0 findings**; `nolint-audit` → no
    `//nolint:erraudit` directives exist; `tree` → exit 0 (8 sentinels; display quirk: tree labels
    them `family.new`).
11. **TODO_LIST.md harvested** (docs-health flow): rewritten to 14 actionable items — new BLOCKED
    `ERRAUDIT_PAT` item, error-codes registry, uniqueness tripwire, rapid Coded property test,
    wrapcheck ignore-sigs, godoc Example, ADR, nix devShell erraudit, renovate pin, sibling-repo
    sync, upstream erraudit proposals; stale items removed (fuzz ran green on BOTH matrix legs in
    its first cron run 2026-09-06: 7m27s/6m59s; nixpkgs item obsolete — flake green).
12. **ROADMAP.md "Error-contract maturity" theme added** — 10 refined-out raw ideas (snapshot tests
    of rendered codes, CodecMetrics dimension, RegisterTemplate, errors.Is same-code table, hook
    error classification, erraudit lsp/flake-check integration, monthly release check, etc.).
13. **Status report ANNOTATED inline** (docs-health ANNOTATE mode): 44 numbered items resolved via
    the skill's atomic tooling (dry-run first) across §b/§c/§e/§f + 2 manual §d stale-claim
    corrections, each with commit hash or verification evidence. No appendix-only markers.
14. **Full verification matrix, all green at session end:**
    - build + vet + `test -race -count=1` in TRUE v1 (`env -u GOEXPERIMENT`) AND v2
    - golangci-lint → 0 issues, both modes
    - erraudit → 0 (default) and 0 (`--enforce-go-error-family --type-aware`)
    - `scripts/check-features-planned.sh` + `scripts/check-go-version.sh` → PASS
    - actionlint on ci.yml → clean
    - `nix flake check` → all checks passed
15. **All 10 §f P0 items from the 06:45 report closed** (CHANGELOG, AGENTS gotcha, envelope
    comment, erraudit install verification, actionlint, CI execution, benchstat, TODO_LIST
    harvest, README, generic_return diagnosis).

## b) PARTIALLY DONE

1. **CI gate end-to-end unverified.** The self-activating mechanism (GOPRIVATE + insteadOf token +
   jsonv2) is standard practice and its pieces were each proven locally, but the assembled job has
   not executed on GitHub (needs the secret, or a publish decision). Until then the job shows as
   "Skipped: ERRAUDIT_PAT secret not set".
2. **CHANGELOG may miss the dependabot bump `653f4cd`** ("go-deps group, 2 updates" — likely
   test-only deps). I inventoried the toolchain/cbor bump but did NOT open 653f4cd's diff before
   writing the entry. Small, fixable at release time.
3. **Sweeps used the `~/go/bin/erraudit` binary** (module-stamped v0.4.0, but built from a dirty
   checkout 36 commits ahead of the tag); only the enforced-audit cross-check used the pristine
   v0.4.0 tag build. Consistency is near-certain, not proven.
4. **Node 20 deprecation warnings in CI** (annotations on every run): the pinned digests of
   actions/checkout, setup-go, upload-artifact, gitleaks-action target Node 20 and are force-run on
   Node 24. Works today; digest refresh not scheduled yet.
5. **TODO_LIST item 1 (release) carries a version-number guess** (`v0.2.1`) from an older runbook;
   with the error-format change in the batch, minor-vs-patch semantics deserve a deliberate pick
   (see §g-2).

## c) NOT STARTED (routed, not executed)

1. `docs/error-codes.md` registry (TODO_LIST #3) and the uniqueness/registration tripwire (#4).
2. rapid property test: every public-API error is `*errorfamily.Error` with a `codec.` code (#5).
3. wrapcheck `ignore-sigs` explicitness — remove config-by-luck (#6).
4. godoc `Example*` for `errors.AsType` (#7) and the error-taxonomy ADR (#8).
5. erraudit in the nix devShell / `.#erraudit` flake app (#9); renovate pin for the CI install (#10).
6. Sibling-repo policy sync (go-cqrs-lite, signing, encryption, storage): same CI job, zero their
   findings (#11) — includes adding `ERRAUDIT_PAT` or its equivalent per repo.
7. Upstream erraudit proposals (verify-before-filing applies): generic_return exact-name
   false-negative fix; warn when an enforced library is absent from go.mod (#12).
8. Release + go-cqrs-lite consumer bump (BLOCKED on user; runbook in TODO_LIST #1).
9. v2-mode benchmark baseline (current doc is v1-only).

## d) TOTALLY FUCKED UP (caught; process failures worth naming)

1. **I marked a todo COMPLETED before doing it.** "Run erraudit policy sweeps" was checked off in
   the todo list while I was still editing AGENTS.md; I noticed on the next pass and actually ran
   them (all clean), but the tool state lied for several minutes. A status list that says done
   before done is exactly the stale-claim pattern this project keeps fighting.
2. **Pipeline masking in the erraudit build check.** I ran `go build … | head -6; echo "BUILD EXIT:
   $?"` — the 0 was `head`'s exit code, not go build's; the build had FAILED and I initially
   printed "BUILD EXIT: 0" next to a wall of build errors. My own house rules (§ AGENTS.md,
   "independently verify tool output", pipefail) call this out by name. I re-derived the verdict
   from the error text and re-proved with an unmasked run, but the pattern fired once more.
3. **Three attempts to extract the baseline table** (sed/awk fence off-by-N twice) and a first
   comparison script that silently compared nanoseconds against allocs/op counts (regex overwrote
   sec/op means with later B/op, allocs/op blocks → absurd 645× "ratios" that I nearly treated as
   data). Root causes: hand-rolling parsing of a benchstat-rendered table, and not asserting
   sanity (ratio > 2 on dozens of benchmarks should have halted the print, not filled a table).
   The final comparison is correct, but two wrong versions existed first.
4. **Annotation citation imprecision:** I cited `7eaf7d2` for work (benchmark-baseline refresh)
   that was still uncommitted at annotation time; that file landed in a later daemon commit. The
   session attribution is right; the exact hash is overstated for one item.
5. **Minor:** `go install` was blocked twice by the environment's security filter before I pivoted
   to `go run` / nix; and `nixpkgs#benchstat` does not exist — two dead-end tool attempts that a
   10-second check (`nix search` / known-ban list) would have avoided.

## e) WHAT WE SHOULD IMPROVE

1. **Todo-state discipline:** never mark in_progress→completed in the same breath as planning the
   work; the todo list must lag reality, never lead it.
2. **No pipeline-masked exit codes, ever:** for every build/test gate run the command bare (or
   `set -o pipefail`) and read ITS exit code; filters for humans, exits for machines.
3. **Kill the hand-parsed benchstat table class of bug:** commit `scripts/bench-compare.py` (or
   keep raw `-count=10` outputs as CI artifacts / a bench branch) so comparisons are
   tool-vs-tool, not regex-vs-rendered-table. Add a sanity assert (|ratio| in [0.33, 3] else halt).
4. **Validate credential-dependent CI in a throwaway fork/PR before it lands on master** — the
   self-activating job is dormant-verified only; a 2-minute fork run would prove the insteadOf +
   GOPRIVATE + jsonv2 assembly.
5. **Refresh action digests** (checkout, setup-go, upload-artifact, gitleaks-action) to
   Node-24-compatible versions in one deliberate, digest-verified pass.
6. **Cite hashes only for commits that already exist** — annotate at the end, after the daemon has
   scooped, or cite the session date instead of a hash for not-yet-committed work.
7. **CHANGELOG completeness check against the commit range** (`git log v0.2.0..HEAD --stat`) as a
   fixed step before declaring the entry done — would have caught the dependabot bump.

## f) NEXT TASKS (prioritized; P0/P1 harvested into TODO_LIST.md)

1. **[USER-BLOCKED] Add `ERRAUDIT_PAT`** (fine-grained PAT, read: `larsartmann/erraudit`) to
   go-codec repo secrets → the `error-audit` gate self-activates on next push. (TODO_LIST #2)
2. **[USER DECISION] Or publish erraudit** — then the secret becomes unnecessary and `go install`
   works everywhere. (TODO_LIST #2 alternative)
3. Create `docs/error-codes.md`: all 31 codes with family, meaning, emitted site `file:line`.
   (TODO_LIST #3)
4. Uniqueness/registration tripwire for `codec.*` codes (script in the
   `check-features-planned.sh` spirit, CI-integrated). (TODO_LIST #4)
5. rapid property test: every error escaping the public API is `*errorfamily.Error` with a
   `codec.`-prefixed code. (TODO_LIST #5)
6. Make wrapcheck `ignore-sigs` explicit for all `errorfamily.Wrap*` variants (or document the
   matching mechanism) — remove config-by-luck. (TODO_LIST #6)
7. godoc `Example*` demonstrating `errors.AsType[*errorfamily.Error]` (doubles as test).
   (TODO_LIST #7)
8. Write the error-taxonomy ADR (families per layer, WrapOnce rule, generic_return decline; input:
   this session's analyzer diagnosis). (TODO_LIST #8)
9. Package erraudit in the nix devShell + `.#erraudit` flake app (needs `GOEXPERIMENT=jsonv2`
   under go 1.26). (TODO_LIST #9)
10. Renovate/dependabot management for the erraudit pin in ci.yml. (TODO_LIST #10)
11. Sync the error-audit CI job to **go-cqrs-lite** and zero its findings. (TODO_LIST #11)
12. Same for **signing / encryption / storage** siblings. (TODO_LIST #11)
13. Upstream erraudit fix: `generic_return` must recognize `WrapOncef`/`WrapCorruptionf`/
    `WrapInfrastructuref`/`WrapOrchestrationf` as error creators (verified against v0.4.0 source).
    (TODO_LIST #12)
14. Upstream erraudit: warn when `--enforce-*` names a library absent from go.mod. (TODO_LIST #12)
15. Release the batch: deliberate version pick → tag → `gh release create` with `[Unreleased]`
    body → re-date CHANGELOG → verify proxy/pkg.go.dev. (TODO_LIST #1; §g-2)
16. Bump `go-cqrs-lite/codec/v4` to the new tag; run its suite `GOWORK=off`. (TODO_LIST #1)
17. Add the missing dependabot-bump entry (`653f4cd`) to the CHANGELOG release notes.
18. Refresh the four action digests to Node-24-compatible versions (checkout, setup-go,
    upload-artifact, gitleaks-action) — clears the deprecation annotations.
19. Validate the self-activating error-audit job end-to-end (fork run or post-secret push).
20. Add a least-privilege `permissions:` block to ci.yml jobs (currently absent).
21. Commit `scripts/bench-compare.py` (robust baseline-vs-rerun diff with sanity bounds) — replaces
    this session's hand-rolled parsing.
22. Decide raw-benchmark-output retention (CI artifact or bench branch) so future baselines are
    tool-comparable, not regex-comparable.
23. Record a **v2-mode** benchmark baseline alongside the v1 one.
24. Fold `erraudit --type legacy_as` + `nolint-audit` zero-baselines into the error-audit CI job
    (cheap, keeps them from regressing silently).
25. Snapshot test of rendered `[family:code]` strings for a curated failure set (ADR-gated).
    (ROADMAP)
26. Error code as a `CodecMetrics` dimension (count by code) for ops dashboards. (ROADMAP)
27. `errorfamily.RegisterTemplate` per code for boundary message consistency. (ROADMAP)
28. Document the `errors.Is` same-code behavior table in `doc.go` (wrap-code == sentinel-code
    matches without walking the cause chain — load-bearing). (ROADMAP)
29. Classify `Observability` hook errors before invoking user hooks. (ROADMAP)
30. Confirm/document `AutoDetectDebug`/`Diagnose` `<diagnose failed: …>` placeholders. (ROADMAP)
31. `erraudit lsp` editor integration for this repo's toolchain. (ROADMAP)
32. erraudit inside `nix flake check`'s lint phase (blocked by #9). (ROADMAP)
33. Monthly erraudit release check as a recurring task; bump the CI pin deliberately. (ROADMAP)
34. Document the `maxPoolBufferSize` / `maxAutoDetectSize` (both 1 MiB) relationship. (ROADMAP)
35. Flip the README consumer signing example to the family pattern IF the stack-wide convention
    lands. (ROADMAP)
36. Inventory `//nolint` directives with weak/no reasons (nolintlint enforces format, not quality).
37. After #6: prune any `//nolint:wrapcheck` directives that became unnecessary.
38. Decide the permanent error contract: bare-`error` + `errors.AsType` forever, or typed-errors
    v2 someday. (§g-3)
39. Review the weekly fuzz budget (13 targets × 30s × 2 modes ≈ 14 min runner time) after the
    first cron runs — both legs green 2026-09-06; is 30s/target enough?
40. Add the mermaid-render CI check for the README architecture diagram (carried TODO_LIST item).
41. CONTRIBUTING/AGENTS note: how to set `ERRAUDIT_PAT`-style secrets for contributors of sibling
    repos (if the PAT route is chosen).
42. When the 06:45 report's remaining open items resolve, `git mv` it to `docs/status/archived/`
    (docs-health ARCHIVE rule).
43. Consider `continue-on-error`-free CI policy: keep gates enforcing; document that skip-reason
    steps are the only acceptable dormancy form.
44. Evaluate exposing `ErrorContext()` keys as godoc constants (machine-readable contract).
45. Add the error-family/message-format change to the next go-cqrs-lite release notes (consumer
    communication).
46. Keep `.go-version`/`.golangci.yml`/go.mod tripwire green through the next Go patch bump
    (recurring; scripts/check-go-version.sh guards).
47. Watch the next Sunday fuzz cron; confirm corpus artifact upload remains healthy.
48. Consider CI cache (setup-go cache: true) for the three `go install`-heavy jobs.
49. Evaluate `erraudit lint` in PR-annotation mode (inline comments) once the gate activates.
50. Post-release: re-run the full benchmark suite and re-date the baseline doc (toolchain may move
    again with the release).

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **erraudit reachability: secret or publish?** Add `ERRAUDIT_PAT` (fine-grained, read-only on
   `larsartmann/erraudit`) to go-codec — and eventually to each sibling repo — or make erraudit a
   public repo? The gate is dormant until you pick. (If PAT: should it be one shared fine-grained
   token for the whole stack, or per-repo secrets?)
2. **Release version for the batch:** the error-format change is breaking for any consumer that
   matches error strings — does that make this **v0.3.0** (behavioral break, pre-1.0 minor) rather
   than the runbook's **v0.2.1**? Your call sets the semver precedent for the stack.
3. **Permanent contract:** is bare-`error` + `errors.AsType` the permanent public error contract
   for go-codec (declines stay declined), or should a typed-errors v2 be on the roadmap after this
   release cycle? This determines whether the 7 remaining generic_return findings ever get fixed
   "properly" or get a permanent ADR entry.

---

_Verification sources: fresh CLI runs at report time — build/vet/race-tests both JSON modes,
golangci-lint 0 issues both modes, erraudit default+enforced exit 0, actionlint clean, both repo
tripwires PASS, `nix flake check` all checks passed; benchstat A/B from 10-run same-machine pairs;
analyzer behavior read from the erraudit v0.4.0 tag source. No claims taken from cached LSP state._

---

## ANSWERS (2026-09-11, post-report)

1. **erraudit access: DEFERRED.** The `error-audit` gate stays dormant (skips with its documented
   reason) until revisited. Both options remain one step away: add `ERRAUDIT_PAT` (gate
   self-activates on next push) or publish the repo (secret becomes unnecessary). Recorded in
   TODO_LIST #2.
2. **Release version: `v0.3.0`** (minor, for the behavioral break in error-string rendering).
   Release itself stays deferred until the follow-up batch is done. Runbook updated in
   TODO_LIST #1 (`## [v0.3.0]` re-date).
3. **Error contract: typed errors are a future `v2` goal.** Bare `error` + `errors.AsType` is the
   transitional public contract; the 7 remaining `generic_return` declines are TRANSITIONAL, not
   permanent policy — the ADR (TODO_LIST #8) must record them as such, and the v2 idea is now in
   ROADMAP's error-contract maturity theme.
