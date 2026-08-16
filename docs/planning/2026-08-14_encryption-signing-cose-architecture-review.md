# Encryption, Signing, and COSE Architecture Review

> 2026-08-14 — comprehensive analysis of how go-codec supports encryption and
> signing, whether COSE should be extracted, whether signing/encryption should
> be extracted from go-cqrs-lite, and the DeterministicCodec marker interface
> proposal.

---

## 1. How go-codec Supports Encryption and Signing Today

go-codec **does not perform cryptography itself.** It provides the byte-shaping
layer that the sibling `signing` and `encryption` modules build on. The
separation is by design: the wire format stays stable and testable, crypto lives
in dedicated modules.

### COSE Layer (RFC 9052)

| Where             | What it provides                                                                                                                                 |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `cose.go:75-109`  | COSE message types: `COSESign1` (signing) and `COSEEncrypt0` (encryption)                                                                        |
| `cose.go:162-277` | `MarshalCOSESign1`/`UnmarshalCOSESign1`, `MarshalCOSEEncrypt0`/`UnmarshalCOSEEncrypt0` — strict element-count validation with stable error codes |
| `cose.go:281-302` | `SigStructure` (RFC 9052 §4.4 — the exact bytes that get signed/verified) and `EncStructure0` (§5.3 — the AAD/KDF input for encrypt/decrypt)     |
| `cose.go:20-73`   | Header labels (alg, kid, IV…) and algorithm IDs (AES-GCM, ChaCha20-Poly1305, EdDSA, HMAC) + `NormalizeCOSEAlgorithm`                             |
| `cose.go:125-150` | `COSEAlgHeader` and `PrepareCOSESetup` — shared boilerplate explicitly factored out of `signing.SignCOSE1` and `encryption.EncryptCOSE0`         |

### Shared Helpers

| Where            | What it provides                                                                                                                                                                                    |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `base64_json.go` | JSON marshalling for `encryption.Ciphertext` / `signing.Signature` wrapper types (`MarshalBase64JSON`, `MarshalBase64JSONWithModule`, `AssignBase64JSON`, `UnmarshalBase64JSON`, `WrapCOSEMarshal`) |
| `cbor.go:13`     | Canonical `CBORCodec` is deterministic — the property that makes signing safe                                                                                                                       |
| `errors.go`      | Stable COSE error sentinels (`ErrInvalidCOSESign1`, `ErrInvalidCOSEEncrypt0`, `ErrCOSEAlgorithmOverflow`, `ErrCOSEInvalidAlgorithm`)                                                                |

### The Flow

1. go-codec shapes and serializes COSE structures with canonical CBOR
2. The external `signing`/`encryption` modules perform the actual crypto
3. `encryption.NewCodec` wraps any `Codec` with encrypt-on-encode / decrypt-on-decode,
   reports encoding `"encrypted"` (codec.go:30)
4. Design keeps crypto out of this module so the wire format stays stable and testable

---

## 2. How COSE Works for JSON and Raw Byte Blobs

### The Separation of Concerns

| Layer                                                                          | Format                                                                         | Deterministic?       |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ | -------------------- |
| COSE envelope (`COSE_Sign1`, `COSE_Encrypt0`, `SigStructure`, `EncStructure0`) | Always CBOR (RFC 9052 requirement — all marshal functions use `CBOREncMode()`) | Yes (canonical mode) |
| Payload bytes inside COSE                                                      | Whatever the `Codec` produces — JSON, CBOR, or raw                             | Depends on codec     |

### Per-Codec Behavior

1. **CBOR payload** — `CBORCodec.Encode()` → canonical CBOR bytes → placed as
   `Payload []byte` in `COSESign1` or `Ciphertext []byte` in `COSEEncrypt0`.
   Natural fit; the payload is already CBOR, the COSE wrapper is CBOR.

2. **JSON payload** — `JSONCodec.Encode()` → JSON bytes → placed as the same
   `[]byte` field. The COSE structure is still CBOR, but the inner payload bytes
   happen to be JSON. Signing/encryption don't inspect or care about the inner
   format — they sign/encrypt opaque bytes.

3. **Raw payload** — `RawCodec.Encode()` → passthrough `[]byte` → same treatment.
   The caller is responsible for whatever those bytes are.

### The Determinism Catch

Signing safety depends on **payload determinism**, not COSE structure
determinism. The COSE envelope (`SigStructure`) is always canonical CBOR, so
that part is always deterministic. But the _payload_ that gets fed into
`SigStructure` is the raw output of `Codec.Encode()`:

- **`CBORCodec`** — deterministic (canonical mode, RFC 7049 sorted keys).
  Signing-safe. This is the recommended default.
- **`JSONCodec`** — **not deterministic** under `encoding/json` v1 (Go map
  iteration order is random). Same input can produce different bytes → different
  signature. Not signing-safe. JSON v2 with `json.Deterministic` fixes this, but
  v1 is the default.
- **`RawCodec`** — caller-controlled. If the caller provides stable bytes, it's
  fine. If not, signatures will break.

### Practical Guidance

- **Signing**: use `CBORCodec` (or `JSONCodec` under `GOEXPERIMENT=jsonv2` with
  deterministic mode). Never sign v1 JSON payloads.
- **Encryption**: any codec works — encryption doesn't require determinism, just
  confidentiality. The ciphertext will differ each time due to nonce/IV, which is
  expected and correct.
- **Detached payloads**: COSE supports nil `Payload`/`Ciphertext` — the bytes are
  signed/encrypted externally and only the signature/ciphertext is stored. This
  works identically regardless of which codec produced the detached content.

---

## 3. Should We Offer a COSE Envelope for Pure JSON? (JWS/JWE)

**No.** Reasoning:

### COSE Already Serves JSON Payloads

The payload inside `COSESign1.Payload` and `COSEEncrypt0.Ciphertext` is `[]byte`
— opaque to the COSE layer. `JSONCodec.Encode()` produces JSON bytes, and those
bytes go straight into the COSE structure. So **you can already sign and encrypt
JSON payloads today.** The COSE envelope is CBOR, but the content it protects is
whatever the codec produced.

### A Pure-JSON Envelope Is a Transport Concern, Not This Module's Job

A JWS/JWE (JOSE) envelope would give you a fully-JSON signed/encrypted structure
with no CBOR anywhere. That's useful when the **entire wire format** must be JSON
— e.g., a browser-facing HTTP API where CBOR bytes are unacceptable. But that's
a **transport-layer concern**, not a payload-serialization concern. This
library's stated purpose is the "payload serialization layer" — it shapes bytes
for the `signing`/`encryption` modules. A JOSE envelope would belong in a
dedicated `jose` module or the `transport/http` package, not here.

### The Determinism Problem Makes JSON Signing Fragile

JWS requires canonical JSON for signature stability (RFC 8785). Go's
`encoding/json` v1 — the project default — has non-deterministic map key
ordering. We'd need to either require `GOEXPERIMENT=jsonv2` (opt-in) or implement
RFC 8785 canonicalization ourselves. Neither is trivial, and shipping a JSON
signing envelope that's non-deterministic by default would be a footgun.

### Surface Area vs. Value

COSE already has 6 types, 4 marshal functions, 2 structure builders, 2
diagnostic helpers, algorithm/header constants, `PrepareCOSESetup`,
`NormalizeCOSEAlgorithm`, plus tests. Adding JOSE would roughly **double** the
COSE surface — `JWSMessage`, `JWEMessage`, `JWSSigningInput`,
`JWEEncryptionInput`, compact vs JSON serialization, header conventions, etc.
That's a lot of code for **zero known consumers**.

### When It WOULD Be Justified

If a concrete consumer emerges — say the `transport/http` layer needs to serve
JWS-signed JSON tokens to browsers — then a **separate `jose` module** (not an
expansion of `go-codec`) is the right move. It can depend on `go-codec` for the
JSON marshalling helpers (`MarshalBase64JSON`, `AssignBase64JSON`,
`WrapCOSEMarshal` pattern) without bloating this library's contract.

**Bottom line:** COSE is codec-agnostic at the payload level. JSON is already
supported. A pure-JSON envelope is a transport concern, belongs in a separate
module, and carries a determinism landmine under v1 JSON. YAGNI applies.

---

## 4. The DeterministicCodec Marker Interface

### Problem

JSON v1 is non-deterministic (random map key order). If someone signs a v1 JSON
payload, the signature can break on re-encoding. This is a silent
signature-corruption bug.

### Solution: Compile-Time Enforcement

A marker interface that turns the footgun into a compile error:

```go
// DeterministicCodec is a Codec whose Encode produces byte-identical output
// for identical inputs. Required for signing and content-addressed storage.
// JSONCodec does NOT implement this under encoding/json v1 (non-deterministic
// map key order); it DOES under GOEXPERIMENT=jsonv2 (Deterministic mode).
type DeterministicCodec interface {
    Codec
    signingSafe() // marker: no body, not callable
}
```

### Implementation Matrix

| Codec                  | Implements `DeterministicCodec`? | Why                                                    |
| ---------------------- | -------------------------------- | ------------------------------------------------------ |
| `CBORCodec`            | Yes, always                      | Canonical mode (RFC 7049 sorted keys)                  |
| `CBORCompactCodec`     | Yes, always                      | CoreDet mode (RFC 8949 bytewise-lexical sort)          |
| `JSONCodec` (v1 build) | **No**                           | `encoding/json` v1 has non-deterministic map key order |
| `JSONCodec` (v2 build) | Yes                              | `json.Deterministic(true)` option                      |
| `RawCodec`             | **No**                           | Caller-controlled, can't guarantee determinism         |

The v1/v2 split is achieved by putting the `signingSafe()` method on `JSONCodec`
only in `json_compat_v2.go` (build-tagged `goexperiment.jsonv2`), not in
`json_compat_v1.go`.

### Enforcement Point

The sibling `signing` module would accept `DeterministicCodec` instead of `Codec`
for its signing path. If someone passes `JSONCodec{}` under v1, it's a **compile
error** — exactly what you want.

### Why This Is the Right Scope

| Concern                  | Answer                                                                                  |
| ------------------------ | --------------------------------------------------------------------------------------- |
| Does it change behavior? | No — pure type-level marker, zero runtime cost                                          |
| Does it force v2?        | No — JSON still works for non-signing paths (transport, debugging)                      |
| Does it handle RawCodec? | No, and it shouldn't — RawCodec determinism is caller-controlled                        |
| Where is enforcement?    | In the `signing` module (separate repo), by accepting `DeterministicCodec`              |
| Is it Go-idiomatic?      | Yes — marker interfaces are the standard pattern for compile-time capability assertions |

### The RawCodec Gap

`RawCodec` is the escape hatch — someone could sign raw bytes that are
non-deterministic. But that's inherent: `RawCodec` is passthrough, the caller
owns the bytes. The marker correctly excludes it, and the signing module can
require an explicit opt-in (e.g., `SignRaw([]byte, ...)`) for detached/raw
signing rather than accepting it through the `Codec` path.

### Status

**Approved.** ~15 lines in go-codec, zero behavioral change. Turns silent
signature corruption into a compile error. Ready to implement.

---

## 5. Should COSE Be Extracted from go-codec into Its Own Module?

**No.** COSE belongs in go-codec.

### Coupling

COSE uses `CBOREncMode()` / `CBORDecMode()` directly (cose.go:115, 176, 237,
289, 301). Extracting would create a circular dep: COSE needs canonical CBOR
mode, go-codec defines it.

### Consumers

Only `signing` and `encryption` use COSE — and they already depend on
`go-codec`. Zero new dependency edges created or removed.

### Composability Payoff

No consumer would import COSE _without_ the rest of go-codec. Both signing and
encryption need `codec.Codec`, `Encoding`, `MarshalBase64JSON`,
`AssignBase64JSON`, `WrapCOSEMarshal` **and** COSE. The seam earns nothing.

### Size

3 source files (`cose.go`, `cose_helpers.go`, `cose_test.go`), ~300 LOC. Not a
god-module.

### Shared Types

`base64_json.go` helpers (`MarshalBase64JSONWithModule`, `AssignBase64JSON`,
`WrapCOSEMarshal`) are already factored to serve COSE consumers. These
cross-reference COSE in their doc comments. Splitting would duplicate or create
awkward cross-deps.

### Unix Principle Test

Fails: COSE doesn't compose like a pipe — it's tightly fused with the CBOR mode
and the base64 JSON helpers. It's not "COSE + codec", it's "codec with COSE
support."

---

## 6. Should Signing and Encryption Be Extracted from go-cqrs-lite?

### Current State

Both are **already separate Go modules** within go-cqrs-lite, with their own
`go.mod` files:

```
github.com/larsartmann/go-cqrs-lite/signing/v4
github.com/larsartmann/go-cqrs-lite/encryption/v4
```

### Cross-Module Coupling: Zero

Signing does **not** import encryption. Encryption does **not** import signing.
Verified across all non-test `.go` files.

### Should They Be 2 Modules or 1? → Keep Them Separate

| Signal                     | Finding                                                                                                                                                                                                                                     |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Cross-imports (production) | **Zero.** Signing does not import encryption. Encryption does not import signing.                                                                                                                                                           |
| Shared types               | **None.** `Signature` vs `Ciphertext` are separate types. `COSESigner` vs `COSEEncrypter` are separate interfaces. They share only `go-codec`'s COSE helpers.                                                                               |
| Shared deps                | `go-codec`, `go-error-family`, `go-cqrs-lite/event/v4`. That's normal shared infrastructure, not coupling.                                                                                                                                  |
| Consumers importing both   | Only `cqrs-lint` (a static analyzer that catalogs modules). Zero runtime consumers need both in the same binary path.                                                                                                                       |
| Co-change frequency        | Signing uses Ed25519/HMAC; encryption uses AES-GCM/XChaCha20. Different crypto primitives, different code paths, different test suites.                                                                                                     |
| Composability payoff       | A consumer who needs signing without encryption (audit logs, event sourcing without encryption) can import signing alone. A consumer who needs encryption without signing (data at rest) can import encryption alone. **The seam is real.** |

Merging would create a module with two unrelated concerns — "crypto operations"
is not one thing, it's two. That's the god-module anti-pattern at the module
level.

### Should They Be Extracted to Be Usable WITHOUT go-cqrs-lite/event/v4?

**No.** The coupling to `event/v4` is **structural and correct**, not incidental.

#### The Dependency Graph

```
signing                          encryption
├── go-codec                     ├── go-codec
├── go-error-family              ├── go-error-family
├── go-cqrs-lite/event/v4 ←──┐   ├── golang.org/x/crypto
│   (Event, Publisher,        │   ├── go-cqrs-lite/event/v4 ←──┐
│    PublishMiddleware,      │   │   (Event, Store, Journal,
│    MetadataKey, NewEvent,  │   │    EventSink, EventSource,
│    PayloadReadOnly, ...)   │   │    PublishMiddleware,
│                            │   │    MetadataKey, NewEvent, ...)
│                            │   ├── go-cqrs-lite/id/v4 ←────┘
│                            │   │   (StreamRef, EventID)
│                            └───│ (event itself depends on id)
└── (nothing else from cqrs)     └── (nothing else from cqrs)
```

#### File-Level Coupling Breakdown

**Signing — 12 non-test files:**

| Pure (no cqrs dep)                         | Event-coupled                                                   |
| ------------------------------------------ | --------------------------------------------------------------- |
| `cose.go` (COSESigner/Verifier interfaces) | `cose_sign1.go` (SignCOSE1/VerifyCOSE1)                         |
| `signature.go` (Signature type)            | `signer.go` (Signer/Verifier interfaces)                        |
| `errors.go`                                | `payload.go` (canonicalPayload)                                 |
| `doc.go`                                   | `middleware.go` (publish middleware)                            |
|                                            | `event.go` (signing event helpers)                              |
|                                            | `ed25519.go`, `hmac.go` (implementations that take event.Event) |
|                                            | `multisig/*` (4 files — all coupled)                            |

**Encryption — 17 non-test files:**

| Pure (no cqrs dep)                                             | Event-coupled                         |
| -------------------------------------------------------------- | ------------------------------------- |
| `cose.go` (COSEEncrypter/Decrypter, EncryptCOSE0/DecryptCOSE0) | `crypto_helpers.go`                   |
| `codec.go` (encryptingCodec)                                   | `store.go` (wraps event.Store)        |
| `aesgcm.go`, `xchacha20.go`                                    | `middleware.go` (publish middleware)  |
| `hkdf.go`, `aead_helpers.go`                                   | `event.go` (encryption event helpers) |
| `ciphertext.go`, `encrypter.go`                                | `envelope.go` (metadata)              |
| `versioned.go`, `static_resolver.go`                           | `algorithm.go` (metadata)             |
| `errors.go`, `doc.go`                                          |                                       |

#### Reasoning

1. **The API surface IS event-shaped.** `SignCOSE1(evt event.Event, ...)`,
   `VerifyCOSE1(evt event.Event, ...)`, encryption middleware wrapping
   `event.Publisher`, `encryption.NewStore(event.Store, ...)`. These aren't
   crypto primitives that happen to be used with events — they are **event
   signing** and **event encryption**. Removing event would gut the API.

2. **The "pure" parts are thin.** The pure crypto layer (signer/verifier
   interfaces, AEAD implementations, Signature/Ciphertext types) is ~12 files
   total. And it overlaps heavily with what go-codec already provides: the COSE
   structures, `MarshalCOSESign1`, `SigStructure`, `EncStructure0`,
   `PrepareCOSESetup`. Extracting the pure parts would create a module that's a
   thin wrapper over go-codec's COSE layer with a few AEAD implementations —
   not enough to justify its own go.mod.

3. **Decoupling via interfaces would add indirection without composability.**
   You'd define `SignedMessage`, `PublishingPipeline`, `EventStore` interfaces
   — but the only implementation would be `go-cqrs-lite/event/v4`. That's not
   composition, it's ceremony. The interface belongs in the event module if
   anywhere, not in a crypto module.

4. **The real extraction target is `event`, not `signing`/`encryption`.** If
   `go-cqrs-lite/event/v4` were extracted to its own repo (it's already a
   separate go.mod), then signing and encryption would transitively depend on
   only `go-codec` + `go-event` + `go-error-family` — a clean, small dependency
   tree. The "usable without go-cqrs-lite" goal is achieved by extracting event,
   not by decoupling signing/encryption from event.

### Actionable Recommendations

| Action                                                           | Why                                                                                        |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Keep signing + encryption as 2 separate modules                  | Zero cross-imports, different crypto primitives, real composability payoff                 |
| Keep them depending on `event/v4`                                | The coupling is structural and correct                                                     |
| Extract `event/v4` to its own repo if you want smaller dep trees | That's the keystone — signing/encryption transitively pull in 5 cqrs modules through event |
| Implement `DeterministicCodec` in go-codec                       | Turns the JSON signing footgun into a compile error                                        |
| Retire the `codec/v4` shim                                       | 26 files still on it; migration is mechanical sed                                          |

---

## 7. The Deprecated `go-cqrs-lite/codec/v4` Shim

### What It Is

A pure alias module — zero logic, 100% re-exports from `go-codec`. The file
`alias.go` is 132 lines of `type X = gocodec.X` and `var Y = gocodec.Y`. It was
the backward-compat layer when go-codec was extracted, and is already marked
deprecated in `doc.go` and `README.md`.

### Migration Status: Half-Finished

| Status                                  | Count         | Examples                                                                                                       |
| --------------------------------------- | ------------- | -------------------------------------------------------------------------------------------------------------- |
| Already migrated to `go-codec` directly | 16 prod files | `benchkit`, `cmd/cqrs-gen`, `system`, `transport/grpc`, `transport/http`, `schema`                             |
| Still using the deprecated shim         | 26 prod files | **`event`** (the keystone), `command`, `decider`, `snapshot`, `kv`, `query`, `stack`, `storage/*`, `watermill` |
| Test files still on shim                | ~30 files     | Same modules' tests                                                                                            |

### The Keystone: `event/codec.go`

The `event` module is the central dependency — nearly everything imports it. It
still does `import "github.com/larsartmann/go-cqrs-lite/codec/v4"` and exposes
`codec.Codec`, `codec.CBORCodec`, `codec.ForEncoding`, etc. through the shim.
Signing and encryption already migrated past it (they import `go-codec`
directly). But until `event` migrates, every downstream consumer (`command`,
`decider`, `snapshot`, `stack`, `storage/*`) is transitively locked to the shim.

### Fix

Mechanical: sed `go-cqrs-lite/codec/v4` → `go-codec` across the 26 production
files (and their tests), starting with `event/codec.go`. Once `event` flips, the
rest cascade. After that, the shim module can be retired with a final
deprecation tag.

Not urgent — it compiles and works. But it's debt that compounds: every new file
written against the shim extends the migration surface.

---

## Summary Decision Matrix

| Question                                         | Decision                                     | Rationale                                                                                                            |
| ------------------------------------------------ | -------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Offer COSE envelope for pure JSON (JWS/JWE)?     | **No**                                       | COSE already handles JSON payloads. Pure-JSON envelope is a transport concern. Determinism landmine under v1. YAGNI. |
| Extract COSE from go-codec?                      | **No**                                       | Circular dep on `CBOREncMode()`. No composability payoff — all consumers need both. 300 LOC, not a god-module.       |
| Extract signing/encryption from go-cqrs-lite?    | **Already extracted** — they have own go.mod | The question is whether to move repos, not split modules.                                                            |
| Merge signing + encryption into 1 module?        | **No**                                       | Zero cross-imports, different primitives, real composability payoff. Merging = god-module.                           |
| Make signing/encryption usable WITHOUT event/v4? | **No**                                       | The API surface IS event-shaped. Decoupling would gut the API and add ceremony without composability.                |
| Extract event/v4 instead?                        | **Yes, if smaller dep trees are the goal**   | That's the keystone. Signing/encryption transitively pull 5 cqrs modules through event.                              |
| Add `DeterministicCodec` marker interface?       | **Yes**                                      | ~15 lines, zero behavioral change, turns silent signature corruption into a compile error.                           |
| Retire `codec/v4` shim?                          | **Yes**                                      | 26 files still on it. Migration is mechanical sed starting with `event/codec.go`.                                    |
