# Error Code Registry

> Single home of the `codec.*` error-code vocabulary. Every code this package
> can emit is registered here with its behavioral family, meaning, and emit
> site. `scripts/check-error-codes.sh` (runs in CI) fails when source and this
> registry drift apart in either direction.

## Contract

Every error returned by this package's wrapping API surface carries:

- a **stable machine-readable code** (`codec.<operation>`, lowercase snake case),
- a **behavioral family** from `github.com/larsartmann/go-error-family`,
- optionally **structured context** (e.g. `ErrorContext()["encoding"]`).

Error strings render as `[family:code] message: cause`. The string format and
the codes are part of the public contract: **codes may be added, never renamed
or reused** — they appear in logs, metrics, and `errors.Is` matching.

### Matching

```go
// Sentinel identity (survives wrapping):
errors.Is(err, codec.ErrInvalidCOSESign1)

// Code / family / context extraction (Go 1.26+):
if e, ok := errors.AsType[*errorfamily.Error](err); ok {
    log.Printf("code=%s family=%v", e.Code(), e.ErrorFamily())
}
```

Wraps that reuse a sentinel's own code match that sentinel via code+family
identity without walking the cause chain; wraps that use a detail code (the
COSE element-count codes) keep the sentinel as cause, so `errors.Is` still
matches.

### Families

| Family          | Meaning                                                            | Retry / handling posture                    |
| --------------- | ------------------------------------------------------------------ | ------------------------------------------- |
| Rejection       | Caller input fault: wrong type, unknown encoding, bad target       | Fix the caller; do not retry as-is          |
| Corruption      | Undecodable stored/wire bytes (COSE parts, transcode, base64)      | Quarantine or repair the data               |
| Infrastructure  | Should-never-fail plumbing (buffer write, envelope marshal)        | Alert; treat as environment/system failure  |
| Orchestration   | Internal dependency-semantics bug (CBOR mode-init panics)         | File a bug; never expected in production    |

### Scope: classified vs. passthrough

Two shapes of API exist by design (see `docs/adr/0001-error-taxonomy.md`):

- **Classified**: the helpers below wrap every failure they touch; every error
  they return is `*errorfamily.Error` with a `codec.`-prefixed code.
- **Passthrough (thin wrappers)**: `CBORCodec`/`CBORCompactCodec`/`JSONCodec`
  `Encode`/`Decode`/`EncodeToBuffer` and the COSE marshal functions return the
  underlying `fxamacker/cbor` / `encoding/json` errors unclassified, because
  only the caller knows whether the bytes are corrupt (Corruption) or the value
  is unencodable (Rejection). Orchestration boundaries above them classify
  (`WrapEncode`, `EncodePooled`, `TranscodeToJSON`, COSE unmarshal).
  `DecodeEnvelopeOrLegacy` likewise returns the primary decode error unwrapped
  for caller classification.

## Sentinel codes (8)

Declared as stable `error` identities in errors.go and codec.go; wraps of a
sentinel reuse its code unless a detail code is noted.

| Code                              | Family    | Meaning                                        | Sentinel                 | Emitted by                          |
| --------------------------------- | --------- | ---------------------------------------------- | ------------------------ | ----------------------------------- |
| `codec.unknown_encoding`          | Rejection | No built-in codec matches the encoding         | `ErrUnknownEncoding`     | `ForEncoding` (codec.go)            |
| `codec.raw_encode_type`           | Rejection | Value is not `[]byte` / raw JSON               | `ErrEncodeRawType`       | `RawCodec.Encode` (raw.go)          |
| `codec.raw_decode_type`           | Rejection | Decode target is not `*[]byte`                 | `ErrDecodeRawType`       | `RawCodec.Decode` (raw.go)          |
| `codec.invalid_cose_sign1`        | Rejection | COSE_Sign1 array does not have 4 elements      | `ErrInvalidCOSESign1`    | `UnmarshalCOSESign1` (cose.go), surface code is the detail `codec.cose_sign1_element_count` |
| `codec.invalid_cose_encrypt0`     | Rejection | COSE_Encrypt0 array does not have 3 elements   | `ErrInvalidCOSEEncrypt0` | `UnmarshalCOSEEncrypt0` (cose.go), surface code is the detail `codec.cose_encrypt0_element_count` |
| `codec.cose_algorithm_overflow`   | Rejection | Algorithm value exceeds int64                  | `ErrCOSEAlgorithmOverflow` | `NormalizeCOSEAlgorithm` (cose.go) |
| `codec.cose_invalid_algorithm`    | Rejection | Algorithm value is not an integer              | `ErrCOSEInvalidAlgorithm` | `NormalizeCOSEAlgorithm` (cose.go) |
| `codec.normalize_depth_exceeded`  | Rejection | `normalizeForJSON` recursion past the depth cap (100) | `ErrNormalizeDepthExceeded` | `normalizeForJSON` (json_compat_v1.go, v1 build only) |

## Detail codes (24)

Wrap-site codes. Element-count wraps keep their sentinel's Rejection family
(the sentinel identity rules the wrap); structural decode failures of stored
bytes are Corruption.

### COSE structures (cose.go)

| Code                                 | Family         | Meaning                                            | Emitted by                                  |
| ------------------------------------ | -------------- | -------------------------------------------------- | ------------------------------------------- |
| `codec.cose_sign1_unmarshal`         | Corruption     | Bytes are not a CBOR array at all                  | `UnmarshalCOSESign1`                        |
| `codec.cose_sign1_element_count`     | Rejection      | Array length ≠ 4; wraps `ErrInvalidCOSESign1`      | `UnmarshalCOSESign1`                        |
| `codec.cose_sign1_protected`         | Corruption     | Element 0 is not a byte string                     | `UnmarshalCOSESign1`                        |
| `codec.cose_sign1_unprotected`       | Corruption     | Element 1 is not an int-keyed map                  | `UnmarshalCOSESign1`                        |
| `codec.cose_sign1_payload`           | Corruption     | Element 2 is not a byte string / nil               | `UnmarshalCOSESign1`                        |
| `codec.cose_sign1_signature`         | Corruption     | Element 3 is not a byte string                     | `UnmarshalCOSESign1`                        |
| `codec.cose_encrypt0_unmarshal`      | Corruption     | Bytes are not a CBOR array at all                  | `UnmarshalCOSEEncrypt0`                     |
| `codec.cose_encrypt0_element_count`  | Rejection      | Array length ≠ 3; wraps `ErrInvalidCOSEEncrypt0`   | `UnmarshalCOSEEncrypt0`                     |
| `codec.cose_encrypt0_protected`      | Corruption     | Element 0 is not a byte string                     | `UnmarshalCOSEEncrypt0`                     |
| `codec.cose_encrypt0_unprotected`    | Corruption     | Element 1 is not an int-keyed map                  | `UnmarshalCOSEEncrypt0`                     |
| `codec.cose_encrypt0_ciphertext`     | Corruption     | Element 2 is not a byte string / nil               | `UnmarshalCOSEEncrypt0`                     |
| `codec.cose_protected_unmarshal`     | Corruption     | Protected header bytes are not a CBOR map          | `UnmarshalCOSEProtectedHeader`              |
| `codec.cose_marshal_protected`       | Infrastructure | Marshaling the alg-only protected header failed    | `COSEAlgHeader`                             |

### Transcode (transcode.go)

| Code                       | Family     | Meaning                                     | Emitted by         |
| -------------------------- | ---------- | ------------------------------------------- | ------------------ |
| `codec.transcode_decode`   | Corruption | CBOR decode failed before JSON re-encode    | `TranscodeToJSON`  |
| `codec.transcode_encode`   | Corruption | JSON re-encode of the decoded value failed  | `TranscodeToJSON`  |

### Envelope, pool, observability (envelope.go, pool.go, observability.go)

| Code                      | Family         | Meaning                                                            | Emitted by                        |
| ------------------------- | -------------- | ------------------------------------------------------------------ | --------------------------------- |
| `codec.envelope_encode`   | Rejection      | Inner codec encode failed; `WrapOncef` preserves an already-classified inner code | `WrapEncode`       |
| `codec.envelope_marshal`  | Infrastructure | Envelope JSON marshal failed                                       | `WrapEncode`                      |
| `codec.pooled_encode`     | Rejection      | `EncodeToBuffer` failed; `WrapOncef` preserves an inner classified code | `EncodePooled`               |
| `codec.observable_write`  | Infrastructure | Writing encoded bytes to the caller buffer failed                  | `ObservableCodec.EncodeToBuffer`  |

### Base64 (base64_json.go)

| Code                   | Family     | Meaning                                          | Emitted by             |
| ---------------------- | ---------- | ------------------------------------------------ | ---------------------- |
| `codec.base64_decode`  | Corruption | Input is neither URL-safe nor standard base64    | `DecodeBase64String`   |

### CBOR mode initialization (cbor.go, cbor_compact.go)

These fire only from the `sync.OnceValue` mode singletons when hardcoded
option constants stop being valid — a dependency-semantics bug, surfaced as a
panic (Orchestration), never as a returned error.

| Code                                | Family        | Meaning                                  | Emitted by                     |
| ----------------------------------- | ------------- | ---------------------------------------- | ------------------------------ |
| `codec.cbor_encmode_init`           | Orchestration | Canonical `EncMode()` rejected options   | `canonicalEncMode` (cbor.go)   |
| `codec.cbor_decmode_init`           | Orchestration | Canonical `DecMode()` rejected options   | `canonicalDecMode` (cbor.go)   |
| `codec.cbor_compact_encmode_init`   | Orchestration | Compact `EncMode()` rejected options     | `compactEncMode` (cbor_compact.go) |
| `codec.cbor_compact_decmode_init`   | Orchestration | Compact `DecMode()` rejected options     | `compactDecMode` (cbor_compact.go) |

## Parameterized codes (not `codec.*`)

The base64 JSON helpers in base64_json.go mint **caller-supplied** codes from
a module + noun pair — `MarshalBase64JSONWithModule`,
`UnmarshalBase64JSON`, `AssignBase64JSON`, `WrapCOSEMarshal` produce e.g.
`encryption.marshal_ciphertext`, `signing.decode_signature`. Those codes
belong to the calling module's registry, not this one; only
`codec.base64_decode` (from `DecodeBase64String`) is codec-owned.

## Maintenance

1. Adding a code: declare the wrap with `errorfamily.WrapXf(..., "codec.<name>", ...)`,
   pick the family per the table above, and add a row to the matching section.
2. CI (`test` job, v1 leg) runs `scripts/check-error-codes.sh`: the set of
   `codec.*` literals in non-test source must equal the set registered here,
   exactly, in both directions. A renamed or removed code without a registry
   update fails the build.
3. Never rename or reuse an existing code — codes are wire-visible contract.
