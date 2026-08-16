# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to `TODO_LIST.md`.

## Themes

### 1. Format coverage

The library currently covers CBOR (two determinism profiles), JSON, and raw
passthrough. Natural directions for consumers who need other wire formats in the
same `Codec` abstraction.

Raw ideas:

- MessagePack codec (compact, JSON-like data model) for interop with systems
  that standardize on msgpack
- Protocol Buffers / FlatBuffers adapter for schema-first stacks
- A first-class `Codec` registration/dispatch table so users can plug custom
  encodings (e.g. `"encrypted"`, proprietary) into `ForEncoding` instead of
  building their own switch. `ForEncoding` already dispatches the three built-in
  encodings (JSON, CBOR, Raw); the open part is user-registration of custom codecs

### 2. Performance & allocation discipline

CBOR and JSON already expose `BufferEncoder` and pool-backed helpers
(`GetBuffer`/`PutBuffer`/`EncodePooled`) for zero-alloc hot paths. There is
headroom to push further on throughput and GC pressure for high-volume event
streams.

Raw ideas:

- Decode-side buffer pool (currently encode-only; decode takes `[]byte` directly,
  so the caller owns the read buffer)
- Explore whether hot payload types benefit from pre-warming the CBOR type cache
  at startup (trading a few ms of init time for consistent first-request latency)
- Benchmark regression detection in CI (baseline comparison, `benchstat`, a
  `nix run .#bench` app) — benchmarks exist, but nothing guards against
  regressions
- Lazy normalization: only convert `map[interface{}]interface{}` keys when the
  v1 marshaler actually chokes, instead of always normalizing

### 3. Schema evolution & drift detection

`CBORCompactCodec` already rejects unknown fields as a drift guard. There is a
richer space around helping payloads evolve safely across versions.

Raw ideas:

- Tooling to diff two payload types' CBOR/JSON wire shapes (detect reorder,
  renumber, rename before they break stored data)
- A "migration" helper that pairs `UnwrapDecode` with a codec swap to move old
  stores to a new encoding incrementally
- Guidance / helpers for the `toarray` field-order commitment (lint that blocks
  reordering fields in an array-encoded struct)

### 4. Observability

`ObservableCodec` telemetry and explainable `AutoDetectDebug` detection are
shipped. Remaining ideas are richer metrics, all deferred until a consumer asks
for them.

Raw ideas:

- `LastEncodeTime` / `LastDecodeTime` timestamps, payload-size histograms, and
  per-encoding aggregated metrics helpers
- Atomics-based `CodecMetrics` (drop `sync.RWMutex`) if benchmarks justify the
  refactor
- Make `maxAutoDetectSize` configurable (safe default) if consumers need a
  different trial-decode ceiling

### 5. Ecosystem & public presence

The module is published and consumed by `go-cqrs-lite` via the module proxy, but
no sibling has wired the newer APIs yet and there is no dedicated public site.
Directions for making it discoverable and proving adoption.

Raw ideas:

- Project website / docs site via the `website-launch` skill (Astro + Starlight
  - Firebase Hosting pattern used by sibling modules)
- A worked end-to-end example consuming `go-codec` from `go-cqrs-lite` to prove
  the `Codec` contract in a real store
- pkg.go.dev polish: ensure every exported symbol carries a godoc example
- A small `codec-cli` diagnostic tool (CBOR→JSON dump, `Diagnose` wrapper) if
  triage workflows want a binary
- Cross-repo integration (lives in `go-cqrs-lite`, driven from here): make the
  sibling `signing` module accept `DeterministicCodec` (compile-time gate
  against non-deterministic codecs — the marker interface exists here, the
  consumer change is there), have siblings reuse `CBOREncMode()`/`CBORDecMode()`
  instead of rebuilding identical modes, wire `ObservableCodec` into the event
  store, use `AutoDetectDebug` for mixed-stream diagnostics, and retire the
  deprecated `codec/v4` shim (mechanical migration starting at `event/codec.go`
  — see `docs/planning/2026-08-14_encryption-signing-cose-architecture-review.md`
  §6-7)

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Cryptography.** Signing and encryption live in sibling modules; this package
  only shapes COSE structures and deterministic bytes. Crypto belongs behind the
  `Codec` seam, not inside it.
- **Event storage / persistence.** Stores, snapshots, and projections are sibling
  modules. `go-codec` stays a pure serialization layer.
- **A custom JSON implementation.** We use the Go standard library
  (`encoding/json` v1 by default, `encoding/json` v2 opt-in via
  `GOEXPERIMENT=jsonv2`) for correctness and interop; we do not reimplement
  JSON.
- **A JOSE/JWS/JWE envelope.** COSE already protects JSON payloads opaquely; a
  pure-JSON envelope is a transport concern for a separate `jose` module, and
  carries a determinism landmine under v1 JSON (see the 2026-08-14 architecture
  review, §3).
- **A security boundary in `AutoDetect`.** Format sniffing is for diagnostics
  and tooling only — it will never gate validation.
