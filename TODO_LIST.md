# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use `ROADMAP.md`.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## High Impact

| Task                                                                         | Status    | Impact | Effort | Evidence                                                                                                  |
| ---------------------------------------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------------------------------------------------------------- |
| Bump `go.mod` `go` directive to `1.27.0`                                     | 🔴 `TODO` | High   | 5min   | `go.mod:3` declares `go 1.26.5`; source imports `encoding/json/v2` + `encoding/json/jsontext` (Go ≥1.27). `go build ./...` fails: "build constraints exclude all Go files". |
| Fix `snaps.Clean` compile error in `TestMain`                                | 🔴 `TODO` | High   | 5min   | `snaps_clean_test.go:12` — `_ = snaps.Clean(m)` rejects 2-value return; `snaps.Clean` returns `(bool, error)` in go-snaps v0.5.23. Fix: `_, _ = snaps.Clean(m)`. Blocks `go test`. |

## Medium Impact

| Task                                                                         | Status    | Impact | Effort | Evidence                                                                                                |
| ---------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------- |
| Add direct unit tests for base64 JSON helpers                               | 🔴 `TODO` | Med    | 1h     | `base64_json.go` — `DecodeBase64String`, `MarshalBase64JSON`, `UnmarshalBase64JSON`, `AssignBase64JSON`, `MarshalBase64JSONWithModule`, `WrapCOSEMarshal` have no dedicated test; only indirectly touched by `transcode_test.go:327`. |
| Add direct unit test for `PrepareCOSESetup`                                 | 🔴 `TODO` | Med    | 30min  | `cose.go:144` — generic option-apply helper has no in-repo test; only exercised by sibling `signing`/`encryption` modules. |

## Low Impact

| Task                                                                         | Status    | Impact | Effort | Evidence                                                                                                |
| ---------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------- |
| Flesh out `CONTRIBUTING.md` dev setup                                        | 🔴 `TODO` | Low    | 20min  | `CONTRIBUTING.md` is a skeleton; missing the Go ≥1.27 requirement, snapshot-update flow (`UPDATE_SNAPSHOTS=true go test ./...`), and lint config notes. |
| Clarify README mono-repo cross-links                                         | 🔴 `TODO` | Low    | 30min  | `README.md:235,261,266-270` link to sibling modules via `../` (`../event/README.md`, `../docs/TIMEZONE_HANDLING.md`, `../transport/http/README.md`) which resolve only inside the parent mono-repo; broken when this module is read standalone on pkg.go.dev. |

---

<!-- Guidance for the builder:
  - Source of truth is the CODE. Verify each item before adding, many
    documented TODOs are already done.
  - DONE items should be REMOVED, not kept. Use CHANGELOG.md for history.
  - Cite evidence (file:line) so the next person can verify without re-deriving.
-->
