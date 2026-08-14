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
| 1  | Decide release strategy and create the GitHub Release (`gh release create`)                      | 🔵 `BLOCKED`  | High   | 5min   | Recommendation: cut `v0.1.1` from HEAD — `v0.1.0` tag at `3f8ac9d` predates the v1-build fix `d871122`. Moving a published tag poisons the module proxy. The two related open questions are resolved: (a) fuzz corpus is artifact-only, no auto-commit; (b) `DeterministicCodec` is documented in README and code for sibling signing integration. Requires user confirmation + remote push/tag. |
