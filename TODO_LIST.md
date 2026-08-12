# TODO List

> Short-term, actionable, bounded work items — each verified open against the
> actual code (not assumed). For long-term vision and unrefined ideas, see
> `ROADMAP.md`. Items are ranked by impact; effort is an estimate.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed; external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

> Every item below was verified against the code on 2026-08-12 and confirmed
> still open. Harvested from the three `docs/status/*` reports and two
> `docs/planning/*` plans; vague/long-term items were routed to `ROADMAP.md`.

## High Impact

| #  | Task                                                                                 | Status    | Impact | Effort | Evidence                                                                                                                                                                                                                           |
| -- | ------------------------------------------------------------------------------------ | --------- | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Add depth cap (`maxDepth`) to `normalizeForJSON` — return error past limit           | 🔴 `TODO` | High   | 15min  | `json_compat_v1.go` — recursive over untrusted CBOR with no cap; stack-overflow DoS in v1 mode. `09-25` §F #1, `03-24` #18, `11-38` #18                                                                                             |
| 2  | Add depth/size guard to `AutoDetect` before trial-decode of untrusted bytes          | 🔴 `TODO` | High   | 15min  | `autodetect.go` — trial-decodes arbitrary input with no size/depth limit. Same DoS class as #1. `09-25` §F #2, `03-24` #17                                                                                                          |
| 3  | Add `gosec` + `gitleaks` steps to CI                                                 | 🔴 `TODO` | High   | 15min  | `.github/workflows/ci.yml` — only test/lint/govulncheck today; security + secret scanning missing. `09-25` §F #9-10, §B                                                                                                             |
| 4  | Add v2-mode lint job to CI (`golangci-lint run --build-tags goexperiment.jsonv2`)    | 🔴 `TODO` | High   | 10min  | `.github/workflows/ci.yml` lint job runs v1 only; `json_compat_v2.go` is compiled in CI but never linted. `09-25` §F #11, §D.4                                                                                                      |

## Medium Impact

| #  | Task                                                                                 | Status    | Impact | Effort | Evidence                                                                                                                                                                                                                           |
| -- | ------------------------------------------------------------------------------------ | --------- | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 5  | Create GitHub Release for `v0.1.0` (`gh release create v0.1.0`)                      | 🔴 `TODO` | Med    | 5min   | Tag `v0.1.0` exists but `gh release view v0.1.0` → "release not found"; release page empty. `09-25` §F #3, §D.3                                                                                                                    |
| 6  | Revert `makezero` to `always: true`, add targeted `//nolint:makezero` in `raw.go`    | 🔴 `TODO` | Med    | 5min   | `.golangci.yml:170-171` globally weakened to `always: false` for one false positive (`raw.go` copy pattern); linter can no longer catch real zero-length-slice bugs. `09-25` §F #4, §D.2                                            |
| 7  | Re-enable `testpackage`; migrate 10 white-box test files to `codec_test` (+ `export_test.go` for `normalizeForJSON`, `canonicalEncMode`) | 🔴 `TODO` | Med    | 30min  | 10 files use `package codec`, 8 use `package codec_test` — inconsistent. `testpackage` linter removed entirely instead of fixing. `09-25` §F #5, §D.1, §G.2                                                                         |
| 8  | Re-enable `paralleltest` in tests; add `t.Parallel()` to the funcs missing it        | 🔴 `TODO` | Med    | 20min  | `.golangci.yml:233` excludes `paralleltest` in `_test.go`; 19 test funcs lack `t.Parallel()`. `09-25` §F #6                                                                                                                         |
| 9  | Add `TestNormalizeForJSON` — table-driven (nil, scalar, map, nested, cycle-like)     | 🔴 `TODO` | Med    | 20min  | `normalizeForJSON` (`json_compat_v1.go`) only exercised transitively today; no dedicated test. Depends on #1 for the depth case. `09-25` §F #22, `03-24` #21                                                                         |
| 10 | Add `FuzzNormalizeForJSON` fuzz target (adversarial nested `map[interface{}]interface{}`) | 🔴 `TODO` | Med    | 15min  | Recursive normalizer over arbitrary CBOR-decoded values is a textbook fuzz target. Depends on #1. `09-25` §F #23, `03-24` #22                                                                                                       |
| 11 | Run all fuzz targets 60s each in v1 and v2 modes                                     | 🔴 `TODO` | Med    | 10min  | `FuzzCBORCodec_*`, `FuzzTranscodeToJSON` compile + seed-pass under `go test` but were never run with `-fuzz=` for any duration. `09-25` §F #13-14, §C                                                                              |
| 12 | Generate coverage report (both modes); add % to FEATURES/README                      | 🔴 `TODO` | Med    | 15min  | No `-coverprofile` run exists; coverage of the compat layer, normalizer, and COSE helpers is unknown. `09-25` §F #15-17, §E.7                                                                                                       |
| 13 | Verify `flake.nix` works: `nix build`, `nix run .#test`, `nix run .#lint`, `nix flake check` | 🔴 `TODO` | Med    | 20min  | `flake.nix` written by adaptation, never executed. `09-25` §F #18-21, `03-24` #7                                                                                                                                                    |
| 14 | Verify `go-cqrs-lite` consumes `go-codec@v0.1.0` (downstream adoption proof)         | 🔴 `TODO` | Med    | 15min  | No consumer has been wired to the published module yet. `09-25` §F #42                                                                                                                                                             |

## Low Impact

| #  | Task                                                                                 | Status    | Impact | Effort | Evidence                                                                                                                                                                                                                           |
| -- | ------------------------------------------------------------------------------------ | --------- | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 15 | Remove dead `testJSONMarshal` from `json_helpers_v1_test.go` + `json_helpers_v2_test.go` | 🔴 `TODO` | Low    | 1min   | Defined in both helper files, never called anywhere (confirmed by grep). `09-25` §C, §F #12                                                                                                                                        |
| 16 | Extract test string constants (`testName`, `testEmail`, …) to dedupe goconst findings | 🔴 `TODO` | Low    | 10min  | `.golangci.yml:232` excludes `goconst` in `_test.go` to silence 20+ repeated-literal findings. `09-25` §F #7                                                                                                                       |
| 17 | Re-enable `tagliatelle` in tests (fix `created_at` → `created-at` or add nolint)     | 🔴 `TODO` | Low    | 5min   | `.golangci.yml:234` excludes `tagliatelle` in `_test.go`. `09-25` §F #8                                                                                                                                                            |
| 18 | `Size` → `SizeResult{JSON, CBOR int}` struct; update callers, tests, `doc.go`        | 🔴 `TODO` | Low    | 30min  | `size.go:12` returns positional `(int, int)` — ambiguous at call sites. `09-25` §F #26-28, `03-24` #15, `11-38` #14                                                                                                                 |
| 19 | Add `TranscodeToCBOR` (symmetric) or document the one-way contract explicitly        | 🔴 `TODO` | Low    | 30min  | `transcode.go` — only `TranscodeToJSON` exists; asymmetry is undocumented. `09-25` §F #29, `03-24` #16, `11-38` #13                                                                                                                 |
| 20 | Add `sync.Pool[*bytes.Buffer]` helper for `BufferEncoder` hot paths                  | 🔴 `TODO` | Low    | 20min  | `BufferEncoder` exists (`codec.go`) but no pool helper for callers. `09-25` §F #30, `03-24` #19, `11-38` #16                                                                                                                        |
| 21 | Benchmark v1 vs v2 JSON paths + `normalizeForJSON` allocation count                  | 🔴 `TODO` | Low    | 30min  | Dual-build perf cost unquantified; normalizer allocates on every CBOR→JSON transcode. `09-25` §F #31-32, `03-24` #46-47                                                                             |
| 22 | Add `CODEOWNERS`, `SECURITY.md`, issue/PR templates                                  | 🔴 `TODO` | Low    | 15min  | None of these exist (`.github/ISSUE_TEMPLATE/`, `CODEOWNERS`, `SECURITY.md` all absent). `09-25` §F #34-37, `03-24` #26-27                                                                                                          |
| 23 | Add `docs/TIMEZONE_HANDLING.md` or accept the inline README section as sufficient    | 🔴 `TODO` | Low    | 10min  | README "Time Handling" section covers the rule inline; no dedicated doc page. Decide: expand or accept. `09-25` §F #38, `03-24` #25                                                                                                 |
| 24 | Flesh out `CONTRIBUTING.md` with snapshot-update flow details                        | 🔴 `TODO` | Low    | 15min  | `CONTRIBUTING.md` covers `UPDATE_SNAPSHOTS=true` but could document the snapshot-update + dual-mode flow more explicitly. `11-38` #24                                                                                               |

---

<!-- Open questions routed from status reports (need a human decision, not a task):
  - Should the v0.1.0 tag move to HEAD (include CI) or cut v0.1.1? See `09-25` §G.1.
    Current state: tag at 3f8ac9d (library code complete); CI commits are post-tag.
  - Is `normalizeForJSON` recursion depth a real threat for your consumers (trusted
    store boundary vs. untrusted input)? Determines whether #1/#2 are P0 or
    defense-in-depth. See `09-25` §G.3, `03-24` g.3. -->
