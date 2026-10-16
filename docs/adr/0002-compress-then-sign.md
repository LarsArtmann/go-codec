# ADR-0002: Compress-Then-Sign — Ordering Contract for Future Compression

**Status:** Proposed (candidate — record now, decide at feature time)
**Date:** 2026-10-16
**Supersedes:** none
**Related:** ADR-0001 (error taxonomy), `cose.go` (COSE_Sign1/COSE_Encrypt0
structures), `docs/research/2026-10-10_ltx-deep-dive.html` (origin of the
invariant framing), the sibling `signing` and `encryption` modules

## Context

go-codec shapes the bytes that the sibling `signing` and `encryption` modules
protect. No compression feature exists in the codec today, but the ROADMAP
keeps the idea alive, and the ltx deep-dive (2026-10-10) surfaced how a
storage format's invariants interact with integrity protection. The failure
mode to avoid is deciding the compress/sign ordering *after* a compression
feature ships — at which point one ordering is already baked into stored
bytes, and the safe ordering may be the incompatible one.

## The Invariant

**If compression is ever added to the encode path, the pipeline order is
fixed: compress first, sign (or encrypt) the compressed bytes — never the
reverse.**

Reasons, in force order:

1. **Verification must not require decompression.** If bytes are signed
   before compression, a verifier must decompress before it can check the
   signature — executing decompressor code on unauthenticated input. That is
   the classic decompression-bomb / parser-on-untrusted-data class; signing
   the compressed form lets the signature check run before any decompression.
2. **Signature validity must be a function of the stored bytes.** Whichever
   representation the store persists is the representation that must be
   signed. Compress-after-sign breaks determinism guarantees too: two encoders
   that compress differently produce different bytes for the same signed
   payload, so signature-validity checks stop implying content equality.
3. **`DeterministicCodec` composes only with byte-deterministic transforms.**
   The signing module accepts `DeterministicCodec`, whose `Encode` must be
   byte-deterministic (that is what makes signatures reproducible). A
   compression stage inserted *after* signing is outside the codec contract
   entirely; one inserted *before* it must itself be deterministic
   (level/dictionary pinned) to preserve the property.

## Decision (to bind when a compression feature lands)

- The compressed representation is the codec's `Encode` output; signing and
  encryption wrap it. COSE_Sign1/COSE_Encrypt0 already model this: their
  payload is opaque bytes, and the pipeline builds structures over the final
  representation.
- Any compression codec must itself satisfy `DeterministicCodec` semantics
  (deterministic output, documented level and dictionary commitments) before
  it can sit in a signing pipeline.
- Wire-format commitment: once any compressed bytes are stored or signed, the
  compression scheme, level, and dictionary become part of the format
  contract, exactly like the `toarray`/`keyasint` cbor tags. Decoders must
  identify the scheme before trusting decoded output.

## Consequences

- Signing safety for a future compression feature is decided at design time:
  the ordering is already written down and referenced from the ROADMAP idea.
- Cost: a future compression PR must carry the deterministic-compression
  proof; the cheap "compress after signing" shortcut is pre-rejected.
