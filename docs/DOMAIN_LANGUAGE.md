# Domain Language

> Ubiquitous vocabulary for the `go-codec` package. Terms used consistently
> across code, docs, and sibling modules. Definitions are domain-level, not API
> signatures.

## Glossary

| Term                          | Definition                                                                                                                                              | Where used                                                          |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| Codec                         | The contract that serializes a value to bytes and back for a declared `Encoding`. The abstraction every store, snapshot, and event relies on.           | `codec.go` — `Codec` interface                                      |
| Encoding                      | A short string tag identifying a wire format (`"json"`, `"cbor"`, `"raw"`). Stamped on payloads so blind stores stay self-describing.                   | `codec.go` — `Encoding` type                                        |
| Canonical CBOR                | Deterministic CBOR per RFC 7049: sorted map keys (length-first), shortest floats. Same input → same bytes, safe for signing and content addressing.     | `cbor.go` — `CBORCodec`                                             |
| Core Deterministic CBOR       | Stricter determinism per RFC 8949: bytewise-lexical key sort. Used by the compact codec; **not** byte-compatible with Canonical CBOR.                   | `cbor_compact.go` — `CBORCompactCodec`                              |
| Determinism                   | The guarantee that identical input always produces identical output bytes — the property that makes an encoding signing-safe.                           | `cbor.go`, `cbor_compact.go`, tests                                 |
| Compact codec                 | The opt-in strict codec: Core Deterministic encoding plus unknown-field rejection on decode (schema-drift guard).                                       | `cbor_compact.go` — `CBORCompactCodec`                              |
| BufferEncoder                 | Optional zero-allocation encoding seam: write directly into a caller-owned `bytes.Buffer` instead of returning a fresh `[]byte`.                        | `codec.go` — `BufferEncoder` interface                              |
| Raw (passthrough)             | An encoding that treats payloads as already-encoded `[]byte`: encode returns them unchanged, decode copies into a `*[]byte`.                            | `raw.go` — `RawCodec`                                               |
| Envelope                      | A self-describing JSON wrapper (magic `"gcdc"` + `Encoding` + inner data) that lets a blind store recover the codec used to write a value.              | `envelope.go` — `WrapEncode` / `UnwrapDecode`                       |
| Transcode                     | Schema-free re-encoding from a stamped encoding to JSON; CBOR is decoded generically and re-marshalled, JSON/raw pass through unchanged.                | `transcode.go` — `TranscodeToJSON`                                  |
| AutoDetect                    | Best-effort heuristic that infers an `Encoding` from raw bytes (leading byte + trial decode). For diagnostics/tooling only, not a security boundary.    | `autodetect.go` — `AutoDetect`                                      |
| Diagnostic notation (EDN)     | Human-readable text form of CBOR bytes, for inspecting corrupt events without decoding into a Go struct.                                                | `cbor.go` — `Diagnose`                                              |
| COSE diagnostics              | Human-readable diagnostic notation of COSE messages (`COSE_Sign1`, `COSE_Encrypt0`) for debugging.                                                      | `cose_helpers.go` — `COSESign1Diagnostic`, `COSEEncrypt0Diagnostic` |
| COSE                          | CBOR Object Signing and Encryption (RFC 9052). This package only **shapes** COSE structures; cryptography lives in the `signing`/`encryption` siblings. | `cose.go`, `cose_helpers.go`                                        |
| COSE_Sign1                    | A single-signer signed COSE message: protected header, unprotected header, payload, signature.                                                          | `cose.go` — `COSESign1`                                             |
| COSE_Encrypt0                 | A single-recipient encrypted COSE message: protected header, unprotected header, ciphertext.                                                            | `cose.go` — `COSEEncrypt0`                                          |
| Protected header              | COSE header parameters included in the cryptographic check (serialized as a CBOR map wrapped in a bstr).                                                | `cose.go`                                                           |
| Unprotected header            | COSE header parameters not cryptographically protected (e.g. key id, IV).                                                                               | `cose.go`                                                           |
| Sig_structure / Enc_structure | The canonical byte strings built for signing (`Signature1`) and encryption (`Encrypt0`) per RFC 9052 — the exact bytes that get signed/AAD'd.           | `cose.go` — `SigStructure`, `EncStructure0`                         |

## Wire-format commitments (cbor struct tags)

These tags turn struct layout choices into part of the on-disk wire format.
Changing them breaks existing stored data and invalidates signatures.

| Tag                    | Meaning                                                            | Commitment                                                                            |
| ---------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------- |
| `toarray`              | Encode a struct as a positional CBOR array instead of a keyed map. | Field **order** is frozen; add new fields only at the end. ~30-40% smaller.           |
| `keyasint`             | Encode struct fields under integer keys instead of string keys.    | Integer **key numbers** are frozen; renumbering breaks existing data.                 |
| `omitzero`/`omitempty` | Skip zero-valued fields.                                           | Changes which fields appear; safe to adopt but alters the wire shape for zero values. |

## Dual-build infrastructure

| Term               | Definition                                                                                                                                                                                                                              | Where used                               |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| Dual-build         | The library supports both `encoding/json` (v1, default) and `encoding/json/v2` (opt-in via `GOEXPERIMENT=jsonv2`).                                                                                                                      | `json_compat_v1.go`, `json_compat_v2.go` |
| Compat helpers     | Five unexported helpers — four functions (`jsonMarshal`, `jsonMarshalDet`, `jsonUnmarshal`, `jsonMarshalBuf`) plus the `rawJSONValue` type alias — that abstract the JSON stdlib. All source files call these, never `json.*` directly. | `json_compat_v*.go`                      |
| `rawJSONValue`     | Type alias for the JSON raw-byte type (`json.RawMessage` in v1, `jsontext.Value` in v2). Used by `RawCodec.Encode`.                                                                                                                     | `json_compat_v*.go`                      |
| `normalizeForJSON` | v1-only recursive normalizer that converts `map[interface{}]interface{}` (CBOR decode artifact) to `map[string]any`, since v1's `json.Marshal` rejects it.                                                                              | `json_compat_v1.go`                      |

## Bounded contexts

- **This package vs. `signing` / `encryption`:** `go-codec` produces
  deterministic bytes and shapes COSE structures; the actual cryptographic
  operations happen in the sibling `signing` and `encryption` modules, which
  consume `CBOREncMode`, `SigStructure`, `EncStructure0`, and the COSE
  marshal/unmarshal helpers defined here.
- **This package vs. `event` / stores:** `go-codec` is encoding-only. Event
  construction, payload decoding (`event.DecodePayload[T]`), and persistence
  belong to sibling modules that depend on the `Codec` contract.
