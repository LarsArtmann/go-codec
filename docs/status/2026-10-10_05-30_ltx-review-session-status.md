# Status Report — LTX Review Session (go-codec)

**Date:** 2026-10-10 05:30 CEST
**Session scope:** Deep review of `github.com/superfly/ltx` for go-codec (single task, one session)
**Prior state:** clean tree at `d3688b7`; auto-commit daemon active
**Note:** This report is Markdown because the requester explicitly demanded a `.md`
path. The status-report skill's canonical format is HTML; this is a logged override,
not a new default.

---

## Session summary

One task: review `superfly/ltx` deeply "for this project." I loaded the
library-deep-dive + verify-external-claims skills, verified ltx is NOT a go-codec
dependency (go.mod + zero repo references), cloned the repo at `f299ac5` (v0.5.3),
read the core source, ran its test suite, verified adoption/metadata/issues via the
GitHub API, and produced a Bauhaus HTML research report at
`docs/research/2026-10-10_ltx-deep-dive.html` (54 KB, tag-balanced, already
auto-committed as `1b752f8`/`b9ab8b1`).

**Verdict delivered:** no as a go-codec dependency (layer mismatch); yes as a
design reference (5 transferable invariants); the real adoption angle is
LiteStream+LTX as backup/replication for go-cqrs-lite's SQLite engines.

---

## a) FULLY DONE

1. **Dependency-status verification** — ltx appears nowhere in go.mod nor anywhere
   in the repo (rg over all files: zero hits). Evidence: go.mod printed, grep exit 1.
2. **Clone refreshed and pinned** — `/tmp/ltx-review` at `f299ac5b951c`
   (2026-09-10), confirmed "Already up to date."
3. **Core source read** — encoder.go (full, 560 ln), decoder.go (to ~405/574),
   checksum.go (full), file_spec.go (full), ltx.go (to 460/642), compactor.go
   (structure + head), encode_db.go (full), apply.go (head), format spec (full).
4. **Test suite executed** — `go test ./...` on ltx: both packages pass (6.9s).
   Plain run only; no `-race` (see d).
5. **Repo metadata verified via GitHub API** — 349 stars, 15 forks, Apache-2.0,
   created 2022-06-17, pushed 2026-09-10; releases v0.5.3 ← v0.3.16 enumerated;
   open issues #98, #101, #103, #88, #93, #77 fetched with dates.
6. **Downstream adoption verified from primary sources** —
   benbjohnson/litestream go.mod requires `superfly/ltx v0.5.2` (14,471 stars,
   pushed 2026-10-09); wjordan/syzy pins v0.5.1 (Sourcegraph). No doc-claim
   trusted unverified.
7. **Maintenance-cadence analysis** — commit histogram by month; 15-month dormancy
   (2023-09 → 2025-01) identified from git log, not memory.
8. **Deliverable written** — `docs/research/2026-10-10_ltx-deep-dive.html`:
   verdict, format anatomy, 6 strengths, 6 findings, go-codec relevance section
   (5 transferable invariants + stack angle), version currency, opportunity table
   with priorities, verification appendix. Tag balance checked (70/70 divs, 9/9
   sections). Auto-committed by the daemon (no manual commit, per harness rule).
9. **Post-hoc honesty fix** — report hero + appendix corrected this session after
   I caught an overclaim (see d): "reads the whole library" → exact coverage
   stated. Fix applied before this status report was written.

## b) PARTIALLY DONE

1. **Source coverage of ltx** — ~70% of core, less elsewhere.
   - Missing: decoder.go 405-574 (contains `streamPageIndex` and
     `pageIndexValidator` implementations — I cited their _behavior_ from Close()
     usage but did not read the implementations); ltx.go 460-642 (`LockPgno` and
     remaining helpers); compactor.go merge logic (~75% unread); cmd/ltx main,
     verify, dump, list, checksum (~430 ln unread); internal/hexdump.go; test
     bodies unread (counts only).
   - Effort to finish: S (30-45 min of reading).
   - Impact of the gap: low for the verdict, medium for the two findings that lean
     on decoder/compactor behavior.
2. **Test analysis** — counted test/bench/fuzz functions (3 benchmarks, zero fuzz
   targets — a real observation vs go-codec's fuzz discipline that did NOT make it
   into the report); suite run once, plain mode. No coverage measurement.
3. **go-cqrs-lite angle** — identified and hedged ("verify WAL single-writer
   compatibility before relying on it") but not validated; cross-repo check never
   attempted (out of this repo's scope, correctly deferred).
4. **Report polish** — sidebar nav has 8 links for 9 sections (#scorecard missing
   from nav); no render check in a browser (structure validated mechanically only).

## c) NOT STARTED

1. **HARVEST of the report's opportunity table into TODO_LIST.md / ROADMAP.md** —
   the status-report skill says section (f) feeds docs-health HARVEST; not run
   (user instructed: report only, no unrelated work). Risk: findings entombed in a
   timestamped file.
2. **Compression-vs-signing ADR note** (report's #1 opportunity, priority 20) —
   not written into docs/adr/.
3. **Cross-referencing prior status/review reports** — docs/status and docs/reviews
   series not consulted for continuity (skill step; skipped under the user's
   "report based on this session only" constraint).
4. **`go test -race` on ltx** — relevant because `ChecksumPages` spawns up to 24
   goroutines writing disjoint slices of a shared `checksums` backing array; I
   reasoned it safe but never ran the detector nor read edge behavior
   (nWorkers > nPages creates empty workers).
5. **Watching ltx upstream** — #77 (HPKE), #98, #101, #103 flagged in the report
   but no tracking mechanism created anywhere.

## d) TOTALLY FUCKED UP

1. **I wrote a false claim into the deliverable.** The report hero said "This
   review reads the whole library at commit f299ac5." I read ~70% of core and
   structurally skimmed the rest. This from a session that had
   verify-external-claims loaded — I policed ltx's claims and then overclaimed my
   own work. Severity: trust damage, zero code impact. Root cause: hero copy
   written from intent, not from the reading log. Mitigation: fixed this session
   (hero + appendix now state exact coverage); the failure itself is recorded here
   so the pattern is visible.
2. **Sourcegraph query was a token bonfire.** I searched `"superfly/ltx" go.mod`
   and got 30 results of go.sum noise (AWS SDK lines) pumped into context.
   Correct tool existed: `gh api` on candidate repos' go.mod (which I ended up
   using anyway). Wasted ~8k tokens and a round trip.
3. **Wrong-org chase on litestream.** Queried `superfly/litestream` (404 — I
   assumed the org from the ltx namespace), plus one failed base64 decode of the
   go.mod content endpoint. Recovered in two calls (search → benbjohnson/litestream)
   but the assumption cost a round trip it shouldn't have.
4. **Skill letter vs user intent judgment made silently.** library-deep-dive says
   "if the library is not found in dependency files, stop and report" — I pivoted
   to a prospective review without saying so at the time. The pivot was right; not
   surfacing the deviation in the moment was not. (It is disclosed in the report's
   appendix now.)

## e) WHAT WE SHOULD IMPROVE

1. **Write hero/summary claims LAST, from the evidence log** — the d)1 failure is
   structural: narrative written before the audit trail is complete. Rule: no
   scope/superlative claim in a deliverable unless a reading log line backs it.
2. **Dependency-trace playbook: gh api first, Sourcegraph second** — for "who
   imports X," `gh api repos/<owner>/<repo>/contents/go.mod` is precise and cheap;
   Sourcegraph belongs only when candidates are unknown.
3. **Finish cited files** — if a finding cites behavior of code I haven't read
   (pageIndexValidator), either read it or hedge in place. Cite-what-you-read.
4. **Race-run foreign test suites** — when a review claims "tests: PASS," running
   `go test -race` costs ~2x time and upgrades the claim materially, especially
   when concurrency is visible in the API surface.
5. **Report nav completeness** — mechanical checklist: every section id appears in
   the sidebar nav. (Trivial, but it shipped wrong once.)
6. **One-shot skill dispatch note** — when deviating from a skill's letter
   (not-a-dependency pivot), state the deviation in the response, not only inside
   the artifact.
7. **Fuzz-gap observation should have made the report** — ltx has zero fuzz
   targets while parsing untrusted bytes (a format decoder!); go-codec fuzzes.
   That is a real quality differentiator I noticed and dropped.

## f) Next tasks (ranked, session-derived)

| #  | Task                                                                                                        | Impact | Effort | Category      |
| -- | ----------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | docs-health HARVEST: route this report's opportunities into TODO_LIST/ROADMAP                               | High   | S      | Documentation |
| 2  | Write compression-vs-signing invariant note (ADR or COSE docs section in go-codec)                          | High   | S      | Documentation |
| 3  | Fix: add fuzz-target gap observation to the ltx report's findings                                           | Medium | S      | Documentation |
| 4  | Read ltx decoder.go 405-574 + compactor.go fully; amend findings if behavior differs from citations         | Medium | S      | Quality       |
| 5  | Decide go-cqrs-lite position: evaluate LiteStream+LTX as sqlite-engine backup (cross-repo task; needs Lars) | High   | M      | Feature       |
| 6  | Add "never-zero distinguished bit" pattern to go-codec envelope design notes (future-work note)             | Medium | S      | Documentation |
| 7  | Add ltx upstream watch items (#77 HPKE, #98/#101/#103 fixes) to ROADMAP                                     | Low    | S      | Documentation |
| 8  | Re-run ltx suite with `-race`; note result in report appendix                                               | Low    | S      | Quality       |
| 9  | Add #scorecard to the report's sidebar nav                                                                  | Low    | S      | Cleanup       |
| 10 | Add docs/research/index.md listing deep-dives (if more follow)                                              | Low    | S      | Documentation |
| 11 | Record the "hero claims last" rule into my workflow (personal lesson, logged here as the record)            | Medium | S      | Process       |
| 12 | If go-codec ever adds envelope checksums: steal ChecksumFlag bit pattern (ROADMAP fuel)                     | Medium | M      | Feature       |
| 13 | Compare COSE_Encrypt0 vs HPKE framing when ltx #77 lands (ROADMAP watch)                                    | Low    | M      | Documentation |
| 14 | Cross-reference prior docs/status reports for open loops in a follow-up session                             | Low    | S      | Process       |

Not listed: anything requiring research into go-codec itself — this session touched
only `docs/` outputs, and the tree was otherwise clean.

## g) Questions I cannot answer myself

1. **HARVEST now or entomb?** Should the report's opportunity table (f/1) actually
   land in TODO_LIST.md/ROADMAP.md now, or is this review a one-shot that stays in
   docs/research? I can run docs-health HARVEST, but whether the ltx-derived items
   deserve backlog slots is a priority call only you can make.
2. **Is the go-cqrs-lite ↔ LiteStream+LTX evaluation wanted at all?** The report's
   strongest forward-looking claim is that SQLite-backed engines get
   backup/replication via LiteStream. Whether that direction interests you (vs.
   the existing storage module's own story) is a cross-repo product decision I
   can't derive from this repo.
3. **ADR appetite:** do you want the compression-vs-signing invariant as a formal
   `docs/adr/0002-*.md` now (S effort), or is a note inside the COSE docs enough?
   Formalizing now costs little but presumes the envelope layer evolves; your
   call on ceremony level.

---

_Generated 2026-10-10 05:30 CEST · Session-scoped: covers only this session's
ltx-review run and defects noticed in it · Auto-commit daemon owns the commit_
