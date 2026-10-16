# AGENTS.md — go-codec

> Context for AI agents working in this repository. Enduring facts only —
> things hard to discover from reading the code. No changelogs, no task lists.

## What This Is

`go-codec` (`github.com/larsartmann/go-codec`, package `codec`) is the payload
serialization layer for an event-sourcing/CQRS stack. It defines the `Codec`
contract used by event stores, snapshots, event construction, signing, and
encryption across sibling modules. It is published as an independent Go module
designed for consumption by `go-cqrs-lite` and other projects.

It does NOT produce a binary. It is a library only: four codecs (CBOR canonical,
CBOR compact, JSON, Raw) plus COSE structure helpers, transcoding, envelope
wrapping, autodetection, streaming, observability, and size/diagnostic utilities.

**Adoption proof:** `go-cqrs-lite/codec/v4` requires the published `go-codec`
tag from the module proxy (not just a go.work overlay) — the sibling repo
builds against the proxy artifact with `GOWORK=off`.

## Directory Structure

Flat single-package layout — every `.go` file is package `codec` at the repo
root. There are no sub-packages.

```
go-codec/
├── *.go               # all implementation, one package (codec)
├── *_test.go          # tests, fuzz, property, snapshot, benchmarks, examples
├── json_compat_v*.go  # dual-build JSON compat layer (v1 default, v2 opt-in)
├── testdata/          # fuzz corpus + go-snaps golden snapshots
├── flake.nix          # build, test, lint, devShell automation
├── go.mod / go.sum    # module metadata
└── README.md          # user-facing docs
```

## Commands

Plain Go toolchain (Go 1.27.2 floor; see ADR-0003). Both `encoding/json` v1
(default) and v2 (`GOEXPERIMENT=jsonv2`) are supported. **Mode must be
explicit:** an ambient `GOEXPERIMENT=jsonv2` silently flips plain invocations
to v2 and excludes every v1-tagged file — a broken v1 file once "passed" the
suite this way. Never verify with bare `go ...`:

```bash
env -u GOEXPERIMENT go build ./...      # build (v1 JSON, TRUE default)
env -u GOEXPERIMENT go test ./... -race # test with race detector (v1)
GOEXPERIMENT=jsonv2 go build ./...      # build (v2 JSON)
GOEXPERIMENT=jsonv2 go test ./... -race # test with race detector (v2)
env -u GOEXPERIMENT golangci-lint run ./...                 # lint (v1)
golangci-lint run --build-tags goexperiment.jsonv2 ./...    # lint (v2)
```

Five tripwires (run in CI v1 leg, pre-commit hook, or one command):

```bash
bash scripts/check-go-version.sh          # go.mod/.go-version/.golangci.yml agree (full patch)
bash scripts/check-error-codes.sh         # errors.go ↔ docs/error-codes.md in sync
bash scripts/check-features-planned.sh    # PLANNED symbols must not exist yet
bash scripts/check-erraudit-version.sh    # erraudit pin agrees ci.yml ↔ flake.nix
bash scripts/check-golangci-pin.sh        # golangci ci.yml pin >= known-good floor
nix run .#tripwires                       # all five, CI order
```

With flake.nix (preferred in LarsArtmann projects): `nix build`,
`nix flake check` (hermetic: format + build + dual-mode tests),
`nix run .#test` / `.#test-race` / `.#lint` (both JSON modes),
`nix run .#write-go-version -- VERSION` (toolchain bump, ADR-0003),
`nix run .#bench-gate` (quiet-machine benchmark gate).

## Architecture

- **`Codec` interface** (`codec.go`) — `Encoding() / Encode / Decode`. The
  optional `BufferEncoder` interface adds zero-alloc `EncodeToBuffer`. `Encoding`
  is a string tag (`"json"`, `"cbor"`, `"raw"`) stamped on payloads so blind
  stores stay self-describing.
- **Dual JSON build.** `json_compat_v1.go` (`//go:build !goexperiment.jsonv2`)
  and `json_compat_v2.go` (`//go:build goexperiment.jsonv2`) provide the same
  five helpers (`jsonMarshal`, `jsonMarshalDet`, `jsonUnmarshal`, `jsonMarshalBuf`,
  `rawJSONValue`) backed by `encoding/json` v1 and v2. All source files call
  helpers, never `json.*` directly.
- **CBOR modes are process-wide singletons.** The canonical/compact enc+dec
  modes are computed once via `sync.OnceValue`. Sibling modules MUST reuse the
  exported `CBOREncMode()`/`CBORDecMode()` — byte-identical output depends on
  the exact options.
- **Two CBOR codecs are NOT interoperable.** `CBORCodec` = CanonicalEncOptions
  (RFC 7049, length-first key sort); `CBORCompactCodec` = CoreDetEncOptions
  (RFC 8949, bytewise-lexical sort) + unknown-field rejection on decode. Different
  sort → different bytes → they cannot read each other's data.
- **`DeterministicCodec`** (`codec.go`) — `Codec` + unexported `signingSafe()`
  marker for codecs whose `Encode` is byte-deterministic (signing-safe).
  `CBORCodec`/`CBORCompactCodec` always satisfy it; `JSONCodec` only in the v2
  build; `RawCodec` never. The sibling `signing` module should accept this
  interface, not plain `Codec`, so non-deterministic codecs fail at compile time.
- **COSE layer** (`cose.go`/`cose_helpers.go`) — RFC 9052 structure
  marshal/unmarshal for `COSE_Sign1` and `COSE_Encrypt0`, plus `SigStructure` /
  `EncStructure0` builders. This package does NOT perform crypto; it shapes bytes
  for the `signing` and `encryption` modules.
- **Errors** use `github.com/larsartmann/go-error-family`: stable codes
  (`codec.raw_encode_type`, …) + behavioral family (Rejection/Corruption/
  Infrastructure/Orchestration). Sentinels are declared as the `error`
  interface; wraps carry the SAME code (or a detail code) so `errors.Is`
  matches via code+family identity AND cause chain. Locked by
  `errors_contract_test.go` + `errors_property_test.go`; codes registered in
  `docs/error-codes.md`; rationale (WrapOnce, classified-vs-passthrough,
  `generic_return` declines, typed-errors-v2 direction) in
  `docs/adr/0001-error-taxonomy.md`.
- **Streaming** (`streaming.go`) — `NewCBOREncoder`/`NewCBORDecoder` return
  `*cbor.Encoder`/`*cbor.Decoder`; `NewJSONEncoder`/`NewJSONDecoder` return
  wrapper types from the dual-build compat files. JSON streaming is NDJSON:
  each `Encode` writes one value + `\n`. v2 gotcha: `jsontext.Encoder` inserts
  separator tokens that corrupt NDJSON (hence per-value `json.MarshalWrite` +
  manual `\n`), and the decoder wraps `jsontext.NewDecoder`, NOT
  `json.UnmarshalRead` (which over-reads and breaks sequential `Decode`). See
  `json_compat_v2.go`.
- **Observability & detection** (`observability.go`, `autodetect.go`) —
  `ObserveCodec` wraps any `Codec` with goroutine-safe `CodecMetrics`
  (counters, byte totals, last errors; `Snapshot()`/`Reset()`) plus an optional
  push-style `MetricsHook` (hook panics propagate; metrics are recorded before
  the hook runs). `AutoDetectDebug`: `Reason` is the stable machine-readable
  contract, `Detail` is unstable prose — never parse it.
- **Buffer pool** (`pool.go`) — `GetBuffer`/`PutBuffer` manage a `sync.Pool`
  of `*bytes.Buffer`; `EncodePooled` handles the full lifecycle; an oversized
  buffer is rejected on put.
- **Performance** — fxamacker/cbor caches type metadata process-wide (cold
  ~117us → warm ~340ns); code generation is NOT needed. Reference benchmarks
  and the 10-run benchstat baseline live in `docs/benchmark-baseline.md`;
  `scripts/bench-compare.py` sanity-gates two raw outputs (wide 0.33..3
  bounds; benchstat for statistics).

## Conventions

- **Indentation:** tabs (see `.editorconfig`). Markdown keeps trailing
  whitespace (`trim_trailing_whitespace = false` for `*.md`).
- **Error wrapping:** codec methods are thin wrappers over `cbor`/`json` and
  use `//nolint:wrapcheck` rather than re-wrapping — callers classify with
  their own context. Orchestrating helpers (`TranscodeToJSON`, `WrapEncode`,
  COSE unmarshal) DO wrap with `errorfamily.WrapXf(err, Family,
  "codec.<operation>", msg)` — never `fmt.Errorf`; `WrapOncef` where the
  inner error may already be classified (envelope, pooled encode) so codes
  never stack. erraudit runs in CI with `--enforce-go-error-family
  --type-aware`; its `generic_return` check stays OFF deliberately — the bare
  `error` interface is required by the `Codec`/`BufferEncoder` contracts;
  type-safe matching uses `errors.AsType[*errorfamily.Error]`.
- **Testing stack:** stdlib `testing`; `onsi/gomega`; `pgregory.net/rapid`
  property tests; `gkampitakis/go-snaps` golden snapshots (`testdata/golden/`);
  native fuzz targets; godoc `Example*` functions (run as tests). `TestMain`
  calls `snaps.Clean`. Mostly black-box (`codec_test`); white-box via
  `export_test.go`. All tests call `t.Parallel()`.
- **Naming:** codecs are value types (`CBORCodec{}`), constructed zero-valued.

## Gotchas

- **`flake.nix` vendorHash:** after ANY `go.mod`/`go.sum` dependency change,
  `nix build` reports a hash mismatch — paste the `got: sha256-…` value into
  `vendorHash`.
- **Dual-build: never import `json.*` directly.** All JSON calls go through
  the compat helpers; `json_contract_test.go` catches violations. goimports
  has corrupted v1 files into importing `encoding/json/v2` before (breaking
  the default build, making gopls diagnostics look "stale"); if the default
  build fails with "build constraints exclude all Go files", check the v1
  compat file imports first. Combined with an ambient `GOEXPERIMENT=jsonv2`,
  a syntactically broken v1 file can produce a green suite (a broken
  `normalize_test.go` once did). Pin the mode (see Commands).
- **`CBORCodec` ≠ `CBORCompactCodec` bytes.** Never assume data written by one
  round-trips through the other (different key sort + compact rejects unknown
  fields). Document per-store which codec owns the data.
- **`time.Time` must be `.UTC()` before encoding.** CBOR uses `TimeUnixDynamic`
  (float64 epoch); decoded times reconstruct in `time.Local`, not the original
  zone. Wall-clock times must not use `time.Time` (store components + IANA zone).
- **`toarray` / `keyasint` cbor tags are wire-format commitments.** Field order
  and integer keys become part of the bytes; reordering fields or renumbering
  keys breaks existing stored data and invalidates signatures.
- **`AutoDetect` is a heuristic, not a security boundary.** It inspects the
  leading byte and falls back to trial decode (skipped for payloads >1 MiB via
  `maxAutoDetectSize`). Never use it to skip encoding validation — always pair
  with the matched codec's `Decode`.
- **`TranscodeToJSON` is schema-free.** CBOR maps → JSON objects with
  non-deterministic key order (Go map iteration). `toarray` structs stay arrays
  (field names are lost). Not suitable for byte-deterministic uses.
- **`RawCodec` copies on Decode.** The target `*[]byte` receives a fresh copy so
  callers can't mutate the input buffer. Encode accepts only `[]byte` or
  `rawJSONValue`.
- **Go-version contract (ADR-0003):** one full-patch string across go.mod
  (authoritative consumer floor), `.go-version`, `.golangci.yml run.go`; bump
  atomically via `nix run .#write-go-version -- VERSION` (or hand-edit all
  three + CHANGELOG). The tripwire strips YAML quotes and rejects 2-part
  values — both encoded real incidents (a quoted `go: "1.27"` could never pass
  the raw-token compare and sat red from v0.3.1 until 2026-10-16). golangci
  also "needs" quotes for 2-part values because YAML parses them as floats;
  full 3-part unquoted satisfies both sides.
- **nix toolchain pinning (three related traps):**
  1. nixpkgs lags Go patch releases — hermetic checks go RED until nixpkgs
     packages the patch (`nix flake update nixpkgs` once available); since the
     `goPkg = pkgs.go_1_27` overrides, only nixpkgs' DEFAULT go lag can bite
     (auxiliary tooling), not package/build/format checks.
  2. `buildGoModule`'s `go` is an outer-param in current nixpkgs:
     `pkgs.buildGoModule { go = …; }` is SILENTLY ignored (build then runs the
     default go under `GOTOOLCHAIN=local` and fails once go.mod outpaces it).
     Always `.override { go = …; }` (see `goModule` in flake.nix).
  3. treefmt's goimports shells out to `go` for import resolution; in the
     hermetic format check an old ambient go with `GOTOOLCHAIN=auto` fails
     offline trying to download the pinned toolchain. `checks.format` pins
     `goPkg` + `GOTOOLCHAIN = "local"`.
- **Auto-commit daemon flattens history into `chore: auto-commit N changed
  file(s)` commits.** Commit human-meaningful changes YOURSELF immediately
  after each edit batch — the daemon scoops uncommitted work within seconds.
  Unpushed daemon noise may be soft-reset and re-committed descriptively;
  pushed history stays. The error-contract overhaul shipped with no meaningful
  commit message anywhere — CHANGELOG and status reports were the only
  narrative.
- **Automated formatter/lint sweeps ride the daemon and can break the version
  contract.** One sweep rewrote go.mod/.golangci.yml to 2-part versions
  (`go 1.27`), added an unreviewed `lll` linter, and bumped a pinned action;
  the daemon committed AND pushed it, leaving master red on every push (the
  go-version tripwire caught it in CI, as designed). After any foreign sweep:
  diff the semantic files (go.mod, .golangci.yml, ci.yml) before assuming
  formatting-only, and run `nix run .#tripwires` locally before pushing.
- **Benchmarks need a quiet machine — this box is multi-tenant.** Sibling
  tenants routinely push loadavg to 30-60; running `go test -bench` under load
  inflates ns/op uniformly 2-3x while B/op and allocs/op stay IDENTICAL (the
  tell: deterministic counters unchanged, timing scattered ±20-50%). Check
  `/proc/loadavg` as step 0 of any benchmarking session; discard runs that
  overlapped 1-min load >10; keep quiet-window occupancy minimal (siblings
  gate on the same silence). Clean canary: `RawCodec_Encode` ≈11-13n on go1.27
  at ±<10% (reject >25n). Baseline-grade runs: `scripts/bench-gate.sh OUT_DIR`.
  The baseline table lives in `docs/benchmark-baseline.md`; accepted raw
  outputs (+ `.ok`/`.loadpeak`, benchstat copy) in `~/bench-artifacts/go-codec/`
  — /tmp is not safe for them (the cleaner ate two suites' evidence overnight).

## Dependencies

- `github.com/fxamacker/cbor/v2` — CBOR codec; canonical & core-deterministic modes.
- `github.com/larsartmann/go-error-family` — categorized errors with stable codes.
- `encoding/json` (v1, default) or `encoding/json/v2` (opt-in via `GOEXPERIMENT=jsonv2`).
- Test-only: `onsi/gomega`, `pgregory.net/rapid`, `gkampitakis/go-snaps`.

## High-Value References

| File                                  | Why it matters                                                                               |
| ------------------------------------- | -------------------------------------------------------------------------------------------- |
| `doc.go`                              | Package-level overview, codec-choice guidance, tag usage                                     |
| `README.md`                           | User-facing usage + when-to-use matrix; sibling cross-refs                                   |
| `FEATURES.md`                         | Honest feature inventory with status                                                         |
| `codec.go`                            | `Codec`/`BufferEncoder`/`DeterministicCodec` contracts                                       |
| `errors.go`                           | Stable error sentinels and codes                                                             |
| `docs/error-codes.md`                 | The error-code registry (family, meaning, emit site per code) — update whenever codes change |
| `docs/adr/0001-error-taxonomy.md`     | Error-contract ADR: families, WrapOnce rule, typed-errors-v2 direction                       |
| `docs/adr/0002-compress-then-sign.md` | Compress-then-sign ordering contract for any future compression (Proposed candidate)         |
| `docs/adr/0003-go-version-policy.md`  | Go-version contract: one full-patch string, three declarations, mechanical enforcement       |
| `scripts/check-go-version.sh`         | Go-version single-source tripwire (three declarations must agree); runs in CI                |
| `scripts/check-error-codes.sh`        | Error-code drift tripwire (source literals equal the docs registry, both directions)         |
| `scripts/check-features-planned.sh`   | FEATURES.md drift tripwire (PLANNED symbols must not resolve via `go doc`); runs in CI       |
| `scripts/check-erraudit-version.sh`   | erraudit pin tripwire (ci.yml ↔ flake.nix agree)                                             |
| `scripts/check-golangci-pin.sh`       | golangci CI-pin floor guard (pin >= known-good floor)                                        |
| `scripts/write-go-version.sh`         | One-argument toolchain bump: writes all three declarations, runs the tripwire (ADR-0003)     |
| `docs/benchmark-baseline.md`          | 10-run benchstat reference baseline (dual-mode) — diff here before perf-sensitive changes    |
