# TODO List

> Short-term, actionable, bounded work items. For long-term vision and unrefined
> ideas, see `ROADMAP.md`.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed; external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Active Items

| #  | Task                                                                        | Status    | Impact | Effort | Notes                                                                          |
| -- | --------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------ |
| 1  | Verify `go-cqrs-lite` consumes `go-codec@v0.1.0` (downstream adoption proof) | 🔴 `TODO` | Med    | 15min  | No consumer has been wired to the published module yet.                        |
| 2  | Create GitHub Release for `v0.1.0` (`gh release create v0.1.0`)             | 🔵 `BLOCKED` | Med | 5min  | Tag exists; awaiting user decision on tag strategy (keep at `3f8ac9d` or cut `v0.1.1`). |
| 3  | Investigate and fix stale gopls/golangci-lint diagnostics                   | 🔵 `BLOCKED` | Critical | M | CLI passes; LSP cache/config mismatch. See status report 2026-08-12.         |
| 4  | Re-run coverage and update `FEATURES.md` figures                            | 🔴 `TODO` | High   | S      | New tests added; coverage numbers likely stale.                                |
| 5  | Add README telemetry section and `ExampleObserveCodec` / `ExampleAutoDetectDebug` | 🔴 `TODO` | High   | S      | End-user discoverability for the new observability APIs.                         |
| 6  | Add concurrent stress test for `ObservableCodec` + `CodecMetrics`           | 🔴 `TODO` | High   | S      | Lock in the goroutine-safety claim.                                              |
| 7  | Document `MetricsHook` panic policy                                       | 🔴 `TODO` | High   | S      | Decide and encode whether panics are recovered.                                |
| 8  | Document `AutoDetectDebug.Detail` as human-readable / unstable              | 🔴 `TODO` | Medium | S      | Code should branch on `Reason`, not parse `Detail`.                              |
| 9  | Add property/fuzz tests for `AutoDetect` ↔ `AutoDetectDebug` consistency    | 🔴 `TODO` | Medium | S      | Rapid/fuzz random payloads to lock in delegation.                                |

## Completed in this session (logged in CHANGELOG [Unreleased])

All 22 items from the prior TODO list have been completed:
- Security: depth cap in normalizeForJSON (#1), AutoDetect size guard (#2)
- CI: gosec/gitleaks (#3), v2 lint matrix (#4)
- Lint: makezero reverted (#6), dead code removed (#15), goconst/tagliatelle re-enabled (#16-17), testpackage migration (#7), paralleltest (#8)
- Tests: normalizer tests + fuzz (#9-10), fuzz verification (#11), coverage report (#12), nix verification (#13)
- API: SizeResult struct (#18), TranscodeToJSON one-way doc (#19), buffer pool (#20)
- Benchmarks: v1 vs v2 (#21)
- Repo hygiene: CODEOWNERS, SECURITY.md, templates (#22), TIMEZONE accepted inline (#23), CONTRIBUTING polish (#24)

---

<!-- Open questions (need a human decision, not a task):
  - Should the v0.1.0 tag move to HEAD (include CI + post-tag work) or cut v0.1.1?
  - Is `normalizeForJSON` recursion depth a real threat for your consumers (trusted
    store boundary vs. untrusted input)? The depth cap is now in place regardless. -->
