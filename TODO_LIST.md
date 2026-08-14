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
| 1  | Create GitHub Release for `v0.1.0` (`gh release create v0.1.0`)             | 🔵 `BLOCKED` | Med | 5min  | Tag exists; awaiting user decision on tag strategy (keep at `3f8ac9d` or cut `v0.1.1`). Note: HEAD `v0.1.0` tag predates the v1-build fix — cutting `v0.1.1` is now the safer option. |

## Completed in this session (logged in CHANGELOG [Unreleased])

Items 3-9 from the prior TODO list are done:

- **Item 3 (stale diagnostics) — root cause found and fixed:** the "stale"
  gopls/golangci-lint diagnostics were not stale — `json_compat_v1.go` and
  `json_helpers_v1_test.go` at HEAD imported `encoding/json/v2` while carrying
  `!goexperiment.jsonv2` build tags, so the default v1 build was genuinely
  broken. Imports restored; both build modes green; CLI and LSP now agree.
- **Item 4 (coverage):** re-measured — 85.3% (v1) / 85.4% (v2), updated in
  `FEATURES.md`.
- **Item 5 (README/examples):** telemetry + `AutoDetectDebug` sections added to
  README; `ExampleObserveCodec` / `ExampleAutoDetectDebug` godoc examples.
- **Item 6 (stress test):** `TestObservableCodec_ConcurrentStress` — 16,000
  concurrent ops, shared metrics + hook, race-clean.
- **Item 7 (panic policy):** documented on `MetricsHook` — panics propagate
  (not recovered), metrics recorded before the hook; locked by
  `TestObservableCodec_HookPanicPropagates`.
- **Item 8 (Detail contract):** `AutoDetectResult.Detail` documented as
  unstable human-readable prose; `Reason` is the stable contract (godoc +
  README + example).
- **Item 9 (property/fuzz):** `TestProperty_AutoDetectDelegatesToDebug` (rapid)
  + `FuzzAutoDetectDebug_Consistency` lock the `AutoDetect` ↔ `AutoDetectDebug`
  delegation for arbitrary payloads.
- **Item 1 (downstream adoption):** verified `go-cqrs-lite/codec/v4` go.mod
  requires `go-codec v0.1.0` from the module proxy (GOWORK=off build green) —
  adoption proof complete; remaining work is the cross-repo PR to wire
  `ObservableCodec` into the event store (see ROADMAP integration items).

Prior session's 22 items remain completed (see CHANGELOG [Unreleased]).

---

<!-- Open questions (need a human decision, not a task):
  - Should the v0.1.0 tag move to HEAD (include CI + post-tag work) or cut v0.1.1?
    The v1-build fix makes v0.1.1 the recommended path. -->
