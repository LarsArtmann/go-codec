# ADR-0001: Error Taxonomy and Contract for go-codec

**Status:** Accepted
**Date:** 2026-09-11
**Supersedes:** none
**Related:** go-cqrs-lite `docs/adr/0002-error-taxonomy.md` (the event-layer
taxonomy this one composes with), `docs/error-codes.md` (the code registry)

## Context

The codec package is the payload-serialization layer for an event-sourcing
stack: event stores, snapshots, signing, and encryption all consume it. Before
this decision, codec errors were free-form (`fmt.Errorf`): consumers could not
distinguish "caller passed the wrong type" (fix the caller) from "stored bytes
are corrupt" (quarantine the data) from "the environment failed" (alert). The
error contract had to become classifiable without leaking codec internals to
callers, and stable enough to appear in logs, metrics, and alerts.

## Decision

### 1. Classification library, not a local enum

Classify via `github.com/larsartmann/go-error-family` (the standalone library
of the stack-wide taxonomy), not a package-local `Family` enum. Codes live in
the codec namespace (`codec.<operation>`) and are registered in
`docs/error-codes.md`; the behavioral families are the library's six
(Rejection, Conflict, Transient, Corruption, Infrastructure, Orchestration).
This keeps go-codec, go-cqrs-lite, signing, encryption, and storage on one
classification vocabulary and one `errors.AsType[*errorfamily.Error]` shape.

### 2. Families per failure layer

| Failure layer in the codec                                  | Family        |
| ----------------------------------------------------------- | ------------- |
| Caller input: unknown encoding, non-`[]byte` raw values, bad decode targets, COSE element counts | Rejection |
| Undecodable stored/wire bytes: COSE structure and per-part decodes, transcode, base64 | Corruption |
| Should-never-fail plumbing: envelope marshal, observable buffer write | Infrastructure |
| Internal dependency-semantics bugs (CBOR mode-init panics)  | Orchestration |

Codec never emits Conflict/Transient: it performs no concurrency control and
has no retryable remote dependencies. Callers add their own context on top
(e.g. a store maps Corruption → quarantine, Rejection → 4xx).

### 3. Sentinel identity is preserved through wrapping

Sentinels are declared as the `error` interface (so `errors.Is` call sites
match the sentinel guard) and always appear first in the chain. Wraps that
reuse a sentinel's own code match via code+family identity; wraps that use a
detail code (the COSE element-count codes) keep the sentinel as cause. Both
forms satisfy `errors.Is(err, Sentinel)`.

### 4. The WrapOnce rule

A boundary that receives an **already-classified** inner error wraps with
`WrapOncef` (no-op when the inner error already carries a code), so codes
never stack: `WrapEncode` over a raw-type rejection surfaces
`codec.raw_encode_type`, not `codec.envelope_encode` stacked on top of it. A
boundary that receives **raw library errors** (cbor, encoding/json) classifies
once with `WrapXf` + the detail code for the failing part
(`WrapCorruptionf`, `WrapInfrastructuref`, ...).

### 5. Classified vs. passthrough API split

- **Classified (wrapping surface):** `ForEncoding`, `NormalizeCOSEAlgorithm`,
  `UnmarshalCOSESign1`/`UnmarshalCOSEEncrypt0`/`UnmarshalCOSEProtectedHeader`,
  `TranscodeToJSON`, `WrapEncode`, `EncodePooled`, `DecodeBase64String`,
  `ObservableCodec` buffer write, mode-init panics. Every error they return is
  `*errorfamily.Error` with a `codec.`-prefixed code — enforced by
  `TestProperty_ClassifiedErrorsCarryStableCodes`.
- **Passthrough (thin wrappers):** `CBORCodec`/`CBORCompactCodec`/`JSONCodec`
  `Encode`/`Decode`/`EncodeToBuffer` and the COSE marshal functions return raw
  library errors (`//nolint:wrapcheck`), because only the caller can know
  whether the bytes are corrupt (Corruption) or the value is unencodable
  (Rejection). Orchestration boundaries above them classify.
  `DecodeEnvelopeOrLegacy` returns the primary decode error unwrapped for the
  same reason (documented in its godoc).

### 6. The bare-`error` contract is transitional

The public `Codec`/`BufferEncoder` signatures intentionally keep returning the
bare `error` interface (Go idiom, interface-contract stability); type-safe
matching is provided by `errors.AsType[*errorfamily.Error]`.

**Direction locked 2026-09-11: typed public errors are a future v2 goal.** A
v2 contract would return a concrete classified error type from the classified
surface (making unclassified errors unrepresentable at compile time). Until
then, the following `generic_return` declines are **transitional, not
permanent policy**: base64_json.go:12, codec.go:38, cose.go:45, cose.go:185,
cose.go:251, raw.go:16, raw.go:34 — one per public classified/passthrough
function that returns `error`.

The erraudit `generic_return` linter stays OFF in this repo's enforcement for
the same reason: the `Codec`/`BufferEncoder` contracts require the bare
`error` interface, and Go idiom favors it at interface boundaries.

**Analyzer diagnosis (2026-09-11, recorded so the number is never
misread):** erraudit v0.4.0's `generic_return` check counted 11 findings
before the migration and 7 after, but the drop 11→7 is partly an artifact:
`createsErrorInternally` exact-matches the method names `Wrap`/`Wrapf` and is
invisible to `WrapOncef`/`WrapCorruptionf`/`WrapInfrastructuref`, so four of
the eleven "resolved" functions (WrapEncode, EncodePooled, TranscodeToJSON,
observability Write) were false negatives, not fixes. The decline rationale
itself stands.

### 7. Tooling mechanics (recorded to kill config-by-luck)

- **wrapcheck:** verified 2026-09-11 with an empty-`ignoreSigs` probe —
  wrapcheck flags only bare `return err` of external-package origin; call-
  wrapped returns (`errorfamily.Wrap*`) pass structurally. The
  `errorfamily.Wrap*` entries in `.golangci.yml` `ignore-sigs` are explicit
  defense-in-depth, not the reason the wraps pass.
- **erraudit CI gate:** `erraudit lint ./... --enforce-go-error-family
  --type-aware`, plus zero-baselines for `--type legacy_as`/`--type
  legacy_is` and a no-`//nolint:legacyerrors` suppression rule, so the legacy
  patterns cannot regrow silently. The job self-activates when an
  `ERRAUDIT_PAT` secret exists and otherwise skips with a documented reason
  (erraudit is a private module; publishing it would drop the secret
  requirement).

## Consequences

**Positive:**

- One classification vocabulary across the stack; retry/quarantine/alert
  decisions become switchable on family.
- Codes are greppable, registered, and tripwired
  (`scripts/check-error-codes.sh` in CI) — renames and phantom codes fail
  builds.
- Sentinel identity survives wrapping; `errors.Is` keeps working for callers
  who never import error-family.

**Negative:**

- Error strings changed format (`codec: X: cause` → `[family:code] X:
  cause`) — a behavioral break shipping in v0.3.0; consumers must not match
  strings (verified: go-cqrs-lite matches none).
- Two API shapes (classified vs passthrough) require reading the split above;
  the property test locks only the classified surface.

**Neutral:**

- Codes, not messages, are the API. Message wording may change freely.
- Typed-error v2 is deferred with a locked direction; no API break is taken
  now.
