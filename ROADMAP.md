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

CBOR and JSON already expose `BufferEncoder` for zero-alloc hot paths. There is
headroom to push further on throughput and GC pressure for high-volume event
streams.

Completed:

- Buffer-pool-backed encode helper (`EncodePooled` in `pool.go`) — callback-based
  API managing `sync.Pool[*bytes.Buffer]` lifecycle automatically
- CBOR reflection caching investigated and documented — fxamacker/cbor caches
  type metadata in a process-wide `sync.Map` (cold ~117µs → warm ~340ns, 344x
  faster). Code generation is NOT needed; the cache handles it
- `toarray` / `keyasint` size/speed tradeoffs benchmarked and documented across
  small/medium/large payload shapes — `BenchmarkTagTradeoffs_Encode/Decode`
- Streaming JSON encoder/decoder for parity with `NewCBOREncoder` /
  `NewCBORDecoder` — `NewJSONEncoder` / `NewJSONDecoder` (NDJSON convention)

Remaining raw ideas:

- Decode-side buffer pool (currently encode-only; decode takes `[]byte` directly,
  so the caller owns the read buffer)
- Explore whether hot payload types benefit from pre-warming the CBOR type cache
  at startup (trading a few ms of init time for consistent first-request latency)

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

`Size` compares JSON vs CBOR sizes today; transcoding and autodetection already
exist. There is room to make codec behavior more measurable.

Raw ideas:

- Encode/decode counters and last-result metrics hooks for stores that want
  per-codec telemetry
- A "why was this detected as X?" debug mode for `AutoDetect` to aid triage of
  mixed-encoding streams

### 5. Ecosystem & public presence

The module is published but has no in-repo consumer wired yet and no dedicated
public site. Directions for making it discoverable and proving adoption.

Raw ideas:

- Project website / docs site via the `website-launch` skill (Astro + Starlight
  + Firebase Hosting pattern used by sibling modules)
- A worked end-to-end example consuming `go-codec` from `go-cqrs-lite` to prove
  the `Codec` contract in a real store
- pkg.go.dev polish: ensure every exported symbol carries a godoc example

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Cryptography.** Signing and encryption live in sibling modules; this package
  only shapes COSE structures and deterministic bytes. Crypto belongs behind the
  `Codec` seam, not inside it.
- **Event storage / persistence.** Stores, snapshots, and projections are sibling
  modules. `go-codec` stays a pure serialization layer.
- **A custom JSON implementation.** We use the Go standard library
  (`encoding/json` v1 by default, `encoding/json/v2` opt-in via
  `GOEXPERIMENT=jsonv2`) for correctness and interop; we do not reimplement
  JSON.
- **A security boundary in `AutoDetect`.** Format sniffing is for diagnostics
  and tooling only — it will never gate validation.

---

<!-- Guidance for the builder:
  - NO bounded actionable tasks here. If it has a clear scope and effort
    estimate, it belongs in TODO_LIST.md.
  - NO status indicators on individual items. This is vision, not inventory.
  - Ideas should be raw and unrefined by design.
  - Revisit quarterly to prune stale directions.
-->
