# Status Report — TODO sweep resume: toolchain hermetic fixes, erraudit drafts, benchmark re-baseline

> Point-in-time snapshot, 2026-10-05 00:32. Continuation of the 2026-10-04
> TODO sweep (prior report: `2026-10-04_08-21_….md`). Benchmarks were still
> running at write time; see b). Auto-commit daemon active throughout
> (latest scoops: `dd353be`, `1e75b72`, `9e151f0`).

## a) FULLY DONE

1. **Item 2 — erraudit issue drafts (verification + drafting).** Evidence
   re-verified against erraudit local HEAD `7773253` (v0.6.1-6): claim (a)
   `createsErrorInternally` (`internal/ast/predicates.go:95-127`) recognizes
   only `errors.New`/`errors.Join`/`fmt.Errorf` (`:244-248`) plus exact-name
   `Wrap`/`Wrapf` (`:116`, constants `:22-23`); the live repro was re-run this
   session — `controlCreator` flagged (WARNING `generic_return`, exit 2),
   `wrapOncefCreator` silently clean, identical under `--type-aware`. Claim
   (b) enforce flags are boolean pass-throughs (`appconfig.go:35-36`,
   `audit_args.go:59-63`) with no go.mod dependency check. Both drafts written
   to `docs/drafts/erraudit-issue-*.md` in own-repo solicited register (no
   provenance banner), passed `check-draft.py --kind body-issue --ai-drafted`
   with 0 FAIL / 0 WARN. Caught and fixed a hallucinated module path
   (`github.com/getsops/samber-oops` → `github.com/samber/oops`) before it
   shipped in the draft. **Not filed upstream — gated on explicit instruction.**
2. **Item 12 — hermetic tail COMPLETE. `nix flake check` fully green**
   (format + build + dual-mode in-sandbox tests):
   - Root-caused why the go-modules FOD ran go 1.26.8: current nixpkgs'
     `buildGoModule` takes `go` as an OUTER `.override` param; a call-argument
     `go = …` is silently swallowed by `...@args`. Verified against
     `pkgs/build-support/go/module.nix` at the locked rev. Fixed via
     `(pkgs.buildGoModule.override { go = goPkg; })`.
   - `vendorHash` updated to `sha256-auINCYX4nm63FITg+IWYBFfSsj7uut6KeWai1VLVR9Y=`
     from the real mismatch output.
   - Hermetic format check fixed: treefmt's `goimports` shells out to `go`;
     ambient-older-go + `GOTOOLCHAIN=auto` → offline toolchain-download
     failure. `checks.format` now pins `goPkg` + `GOTOOLCHAIN = "local"`;
     duplicate auto-added `checks.treefmt` disabled (`flakeCheck = false`).
   - Verified in isolation (`nix build .#checks.x86_64-linux.format`) and via
     the full `nix flake check` (all checks passed).
3. **Item 4/7 groundwork.** benchstat provisioned into dedicated
   `/tmp/go-tools` GOBIN (locked nixpkgs `c59305ba` has NO benchstat — eval
   verified failing). `scripts/bench-compare.py` created: raw-vs-raw parser,
   per-benchmark mean of ns/op + B/op + allocs/op across counts, sanity bounds
   ratio ∈ [0.33, 3] else exit 1, GOMAXPROCS-suffix name normalization,
   zero/zero → 1.0, drift warnings for one-sided names, exit 2 on parse
   errors. Unit-tested on synthetic fixtures (all three exit codes, both
   warning classes, zero-handling).
4. **Benchmark runs (item 4).** v1 mode COMPLETE: 706 lines, `ok
   github.com/larsartmann/go-codec 878.609s` (10-count, go1.27.1, quiet
   machine — sequenced after all CPU-heavy verification). Early signal:
   `JSONCodec_Encode` ~952 ns / 88 B / 4 allocs vs the 2026-09-11 baseline's
   263 ns / 192 B / 6 allocs — fewer allocations but much slower; plausibly
   the go1.27 `encoding/json` v1 engine change; the 10-count benchstat mean is
   the authority, not single lines.
5. **Docs sync (most of it).**
   - README: Go floor line 1.26.7+ → 1.27.1+; the "Go 1.27+ natively" v2
     claim VERIFIED empirically (a scratch module imports `encoding/json/v2`
     and builds on 1.27.1 with `GOEXPERIMENT` unset and empty).
   - AGENTS.md: Go 1.27+, `nix run .#tripwires` command, bench-compare.py in
     the Performance bullet, plus two new Gotchas (buildGoModule outer-param
     trap; treefmt goimports/GOTOOLCHAIN fix) and an updated nixpkgs-lags note.
   - TODO_LIST: items 5/6/9/11/12 deleted (done); item 2 → BLOCKED (drafts
     prepared, filing gated); items 4/7 → IN_PROGRESS with status; item 3 note
     extended with this session's wipe incident; NEW item 13 (v0.3.1 tag has
     no CHANGELOG section).
   - CHANGELOG Unreleased: Added (release workflow, tripwires app,
     bench-compare.py, dependabot+cache, permissions, drafts) / Fixed
     (buildGoModule go pin, hermetic format check) / Changed (Go 1.27.1
     toolchain bump) sections written.
   - SECURITY.md reviewed — current, no changes needed.
6. **actionlint** over ci.yml + release.yml (pinned actionlint from locked
   nixpkgs): clean.
7. **github-voice fan-out symlink repaired** (`~/.config/crush/skills/
   github-voice` → `/home/lars/projects/SKILLS/github-voice`); skill verified
   readable through the fan-out path.

## b) PARTIALLY DONE

1. **Items 4+7 — benchmark re-baseline.** v1 raw output done; v2 run at ~60%
   (411/~706 lines) at report time, same 10-count protocol. Remaining:
   benchstat summaries for both modes, `bench-compare.py` verification against
   real output (only synthetic fixtures so far), `docs/benchmark-baseline.md`
   rewrite (v1+v2 sections, go1.27.1, 2026-10-05, artifact-only retention
   recommendation per fuzz-corpus precedent), then delete TODO 4/7.
2. **Final verification sweep.** actionlint ✅; still to run after benchmarks
   (quiet-machine sequencing): golangci-lint both JSON modes, `go test -race`
   both modes, `nix fmt` drift check, lsp diagnostics on touched files.

## c) NOT STARTED

1. Items 1 (ERRAUDIT_PAT), 3 (daemon policy), 8 (v0.1/v0.2 release backfill),
   10 (contributor secrets note) — user decisions, deliberately untouched.
2. New TODO 13 (v0.3.1 CHANGELOG reconciliation) — routed, not started.
3. Filing the erraudit drafts — waits on user instruction.

## d) TOTALLY FUCKED UP!

1. **Mid-session working-tree wipe (between sessions).** This session started
   with `AD docs/drafts/*` + `MM flake.nix` git state: staged drafts were
   DELETED from the worktree and my flake.nix fixes reverted; the daemon
   commit carrying `go = goPkg` (`d75c8a1`) was left DANGLING (discarded from
   master). Recovered everything (drafts rewritten verbatim from session
   context; flake fixes re-applied; hash reused from the real mismatch).
   Root cause NOT determined (daemon heuristic and git-town sync are the
   suspects). Without the conversation summary this work would have been
   unrecoverable — same failure class as the 2026-10-03 lost-draft incident.
2. **First flake fix was wrong twice.** (a) `go = goPkg` as a call argument —
   silently ignored by buildGoModule, and my rebuild initially "confirmed" the
   wrong theory because of the pipe-status trap (below). (b) Even the
   `override.__functionArgs` introspection output was misread on first pass.
   Only reading the nixpkgs `module.nix` source produced the correct fix.
3. **`rg -rn` trap hit AGAIN** (the `-r` is REPLACE, not recursive) — despite
   the handoff and prior reports explicitly warning about it. Second
   documented occurrence in this repo's sessions.
4. **Pipe-exit-code trap repeated.** `nix build .# 2>&1 | tail && echo OK`
   printed OK on a FAILED build; `PIPESTATUS` is unsupported in this shell
   (mvdan/sh). Fixed pattern: `set -o pipefail; … ; rc=$?`. Cost one wasted
   build cycle and one false "BUILD OK" report.
5. **bench-compare test data bug.** First "regression" fixture used 1495/513
   ns = ratio 2.91 — INSIDE the 3.0 bound — so the exit-1 path was not
   actually exercised by that test; caught on inspection and re-tested with
   ratio 5.855. Synthetic tests must be computed against the threshold, not
   eyeballed.

## e) WHAT WE SHOULD IMPROVE!

1. **Protect in-flight work from the wipe** until item 3 (daemon policy) is
   settled: hand-commit meaningful changes immediately rather than leaving
   them to the daemon's next scoop; the daemon/git-town interplay destroyed
   staged work once now.
2. **Read nixpkgs source before version-sensitive API guesses.** The
   buildGoModule `go` outer-param change cost two failed build cycles; the
   module source was one fetch away. Same discipline as "read before you
   write" applied to nixpkgs APIs.
3. **Shell hygiene in this environment:** never rely on PIPESTATUS; always
   `set -o pipefail` + explicit rc capture when piping (now also a documented
   personal-level lesson).
4. **Pin benchstat** (`@latest` install is floating) — either a versioned
   GOBIN install or upstream nixpkgs packaging request.
5. **FEATURES.md perf figures** cite the baseline doc; after the baseline
   rewrite, verify those indicative numbers still hold or annotate them
   (go1.27's v1-json change makes some stale).
6. **Fan-out checker missing:** global AGENTS documents
   `~/.config/crush/scripts/check-skill-fanout.sh`, which does not exist on
   this machine — crush-config repo issue (the symlink itself is now fixed;
   the checker gap remains).
7. **v0.3.1 CHANGELOG drift** (new TODO 13): release.yml would emit
   generated notes for that tag; reconcile.

## f) NEXT TASKS (prioritized)

1. Wait for v2 benchmark completion; benchstat both raw files.
2. Verify `scripts/bench-compare.py` against the REAL v1 output (self-compare
   - v1-vs-v2 cross-check for the sanity-gate behavior).
3. Rewrite `docs/benchmark-baseline.md`: v1+v2 sections, go1.27.1 environment,
   2026-10-05 date, benchstat tables, artifact-only retention decision.
4. Delete TODO items 4/7; append their CHANGELOG entries.
5. Run golangci-lint (v1 + v2 build tags) — post-benchmark.
6. Run `go test -race -count=1` both modes — post-benchmark.
7. `nix fmt` drift check (expect 0 changed).
8. lsp diagnostics on files touched this session.
9. Cross-check FEATURES.md indicative perf numbers against the new baseline.
10. One human-meaningful hand commit for this session's work (daemon noise
    otherwise buries it).
11. Answer-driven: file erraudit drafts (needs user go-ahead).
12. Answer-driven: item 1 ERRAUDIT_PAT vs publish erraudit.
13. Answer-driven: item 3 daemon policy (now with the wipe incident as
    evidence).
14. Answer-driven: item 8 v0.1/v0.2 release backfill.
15. TODO 13: reconcile v0.3.1 with the CHANGELOG.
16. Investigate the wipe root cause (daemon logs / git-town state) if the
    user confirms it was not manual.
17. go-cqrs-lite sibling sweep: adopt Go 1.27.1, re-verify against published
    go-codec tag with GOWORK=off (ecosystem follow-up of the toolchain bump).
18. Consider pre-commit tripwire hooks (recurring improvement item — two
    silent go.mod bumps plus this session's evidence).
19. Consider a scheduled CI benchmark job (weekly, bench-compare sanity gate
    against committed baseline summary).
20. Pin/bundle benchstat reproducibly (versioned GOBIN or nixpkgs PR).
21. Add `bench-compare.py` usage line to README's development section (tool
    is committed; discoverability).
22. Re-run `nix flake check` once more at session end (post-docs edits) as
    the final gate.

## g) QUESTIONS (cannot resolve myself)

1. **The wipe:** do you know what reverted the working tree between the last
   session and this one (git-town sync? daemon reset? manual action)? It
   destroyed in-flight staged work; the answer decides whether I should
   hand-commit immediately after every meaningful edit until item 3 is
   settled.
2. **Erraudit drafts:** file both issues upstream now, or do you want to
   review the two `docs/drafts/erraudit-issue-*.md` bodies first?
3. **Baseline propagation:** should the new benchmark numbers flow into
   FEATURES.md's indicative `~` figures (some will shift on go1.27's v1-json
   change), or stay baseline-doc-only?
