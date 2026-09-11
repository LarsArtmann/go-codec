# Error-Contract Follow-up Batch 2 — Registry, Tripwires, Property Test, ADR, Sibling Sync

- **Session scope:** Execute TODO_LIST items #3–#11 (+#14) from the error-contract
  follow-up queue: error-code registry, uniqueness tripwire, rapid Coded property
  test, wrapcheck mechanism, godoc example, error-taxonomy ADR, nix erraudit
  wiring, pin management, sibling-repo sync, mermaid CI check.
- **Verdict:** All 10 queued items executed and verified. Full gate matrix green
  (build/test -race/lint both JSON modes, 4 tripwires, erraudit enforced gate,
  actionlint, `nix flake check`). Two honest findings recorded: the CHANGELOG
  under-counted codes (31 → actual 32, two sentinel names typo'd), and
  go-cqrs-lite is NOT at zero erraudit findings (253 across 22 modules) — routed
  to its own TODO_LIST as the gate-activation precondition.
- **Deferred on user decisions (unchanged):** release (v0.3.0 chosen, not tagged),
  ERRAUDIT_PAT secret, upstream erraudit issue filing (#12 — needs explicit
  instruction; verify-before-filing applies).

## a) Deliverables

1. **`docs/error-codes.md`** — the registry: all 32 `codec.*` codes with family,
   meaning, emit site; matching guidance (`errors.Is` sentinels /
   `errors.AsType`); families table with handling posture; the
   classified-vs-passthrough API split; parameterized (`<module>.*`) codes
   documented as caller-owned, not codec-owned.
2. **`scripts/check-error-codes.sh`** (CI, `test` job v1 leg) — bidirectional
   drift tripwire: set of `"codec.*"` literals in non-test source must equal the
   backticked tokens in the registry. Negative-tested both directions (phantom
   code, missing code). It caught its first real defect during development: the
   registry doc's backticked filename mention `codec.go` false-positived until
   the extraction was tightened to backticked tokens.
3. **`errors_property_test.go`** — rapid property: every error escaping the
   classified API surface is `*errorfamily.Error` with a `codec.`-prefixed code
   and a known family. 13 API calls × generated failing inputs
   (unencodable values, wrong targets, random bytes/strings, uint64 overflow);
   passes in both JSON modes. Scope note in file header: thin codec wrappers are
   deliberately out (passthrough by design, ADR §5).
4. **wrapcheck (#6)** — resolved by evidence, not config-by-luck: an
   empty-`ignoreSigs` probe showed wrapcheck flags ONLY bare `return err` of
   external origin; the `errorfamily.Wrap*` wraps pass structurally. All 16
   errorfamily signatures now explicit in `.golangci.yml` (defense-in-depth) +
   mechanism documented in a config comment. Probe also proved test files are
   the only wrapcheck findings (excluded by existing rules).
5. **`ExampleForEncoding_classified`** (#7) — godoc example (doubles as test)
   printing code/family/context via `errors.AsType`; pinned to actual rendering
   (`family: rejection` — lowercase `String()`).
6. **`docs/adr/0001-error-taxonomy.md`** (#8) — classification library choice,
   families per failure layer, sentinel-identity preservation, WrapOnce rule,
   classified-vs-passthrough split, wrapcheck mechanics, and the LOCKED
   direction: typed public errors are a future v2 goal; the 7 `generic_return`
   declines are transitional. Records the 2026-09-11 analyzer diagnosis (the
   11→7 drop was partly a `createsErrorInternally` false-negative).
7. **`flake.nix`** (#9) — `.#erraudit` app + devShell `erraudit`: runs the
   pinned erraudit from source at invocation time (`GOPRIVATE` + `GOEXPERIMENT=jsonv2`
   in the wrapper). NOT a flake input — erraudit is private, an input would
   force credentials for every nix command. Verified end-to-end:
   `nix run .#erraudit -- lint ./... --enforce-go-error-family --type-aware` → exit 0.
8. **`scripts/check-erraudit-version.sh`** (#10, CI) — ci.yml ↔ flake.nix pin
   consistency (v0.4.0). Renovate is NOT installed on this repo (no PRs, no
   branches, no sibling configs) and dependabot's gomod ecosystem cannot manage
   `go install` pins — so bumps stay deliberate (ROADMAP monthly check), guarded
   by this tripwire. Negative-tested (v0.9.9 drift → FAIL).
9. **CI mermaid job** (#14) — renders every mermaid block in README.md via
   `@mermaid-js/mermaid-cli@11.17.0` (verified latest against the npm registry)
   + setup-node v4.4.0 (digest-pinned). Diagram verified renderable locally
   (nixpkgs mermaid-cli 11.17.0 → SVG). shellcheck-clean after a SC2012 fix.
10. **Sibling sync (#11) — go-cqrs-lite** (signing/encryption/storage are its
    sub-modules, not separate repos): added the same self-activating
    `error-audit` job with PER-MODULE enforcement (35 modules). Repo baseline:
    **253 findings across 22 modules** (storage 53, graph 46, event 25,
    encryption 14, command 13, decider 12, kv 11, …) — recorded in the job
    comment AND go-cqrs-lite `TODO_LIST.md` (CI section) as the activation
    precondition; the gate cannot redden master because it skips until a
    secret exists. Their daemon owns the commits.

## b) Corrections found and fixed

- `CHANGELOG.md [Unreleased]`: "31 stable error codes" → **32** (verified by
  counting unique literals); sentinel names `codec.encode_raw_type`/
  `codec.decode_raw_type` typo'd → corrected to `codec.raw_encode_type`/
  `codec.raw_decode_type` (source + `errors_contract_test.go` are the truth).
- Taxonomy prose in the same entry: `normalize depth cap` (Rejection) was
  listed under Corruption; `base64 failures` (Corruption) under
  Infrastructure. Per-code families now live only in the registry.
- `erraudit --type nolint-audit` does not exist (flag list verified); the
  zero-baseline is `--type legacy_as` + `--type legacy_is` + a
  no-`//nolint:legacyerrors` grep (no such suppressions exist in source).

## c) Errors encountered (and root causes)

1. **`git checkout -- flake.nix` wiped uncommitted work** during a tripwire
   negative test — HEAD predated the #9 changes. Re-applied from memory of the
   exact diff. Rule re-learned the hard way: NEVER restore/check-out with a
   dirty tree; negative-test by sed + sed-back (which the ci.yml drift test
   then did correctly).
2. **First drift-test design was flawed** (`v0.4.0-PHONY` suffix still matched
   the prefix regex) — re-tested with a genuine version mismatch.
3. **go-cqrs-lite `./...` illusion** — a root-module erraudit run exits 0
   because the monorepo has 34 sub-modules the root module never sees.
   Per-module enumeration (their `discover-modules` pattern) is the only
   honest audit shape there.

## d) Honesty notes

- The erraudit-version tripwire counts matches, not semantics: it would not
  catch a pin bump that only changed one file if BOTH files were edited to
  DIFFERENT-but-valid versions... it does catch that (mismatch). It cannot
  catch "both files bumped to a stale version" — that is the ROADMAP monthly
  deliberate check, by design.
- go-cqrs-lite ci.yml edit is NOT pushed (no push standing constraint);
  their Actions state is unchanged until the user pushes.

## e) Verification matrix (all green, 2026-09-11)

build v1+v2; `go test ./... -race` v1+v2; golangci-lint 0 issues v1+v2;
error-codes PASS (32); erraudit-version PASS (v0.4.0); features-planned PASS;
go-version PASS (1.26.7); actionlint clean; erraudit enforced gate exit 0
(both `~/go/bin/erraudit` and `nix run .#erraudit`); `nix flake check` all
passed; go-cqrs-lite ci.yml actionlint clean.

## f) Left open

- #1 release v0.3.0 (user-gated; runbook in TODO_LIST).
- #2 ERRAUDIT_PAT (user-gated).
- #3 upstream erraudit filings (user-gated; drafts not yet written).
- #4 daemon policy (user-gated).
- go-cqrs-lite 253-finding zeroing campaign (routed to its TODO_LIST).
