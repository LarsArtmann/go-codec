package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// TestDeterministicCodec_CBORAndCompactAlwaysSatisfy locks the documented
// contract that both CBOR codecs are signing-safe in every build: their
// encoding modes sort map keys deterministically, so identical inputs always
// produce identical bytes. See codec.DeterministicCodec.
func TestDeterministicCodec_CBORAndCompactAlwaysSatisfy(t *testing.T) {
	t.Parallel()

	if _, ok := any(codec.CBORCodec{}).(codec.DeterministicCodec); !ok {
		t.Error("CBORCodec must always satisfy DeterministicCodec (canonical CBOR is byte-deterministic)")
	}

	if _, ok := any(codec.CBORCompactCodec{}).(codec.DeterministicCodec); !ok {
		t.Error("CBORCompactCodec must always satisfy DeterministicCodec (core deterministic CBOR)")
	}
}

// TestDeterministicCodec_RawNeverSatisfies locks the documented contract that
// RawCodec is never signing-safe: it performs no encoding, so determinism is
// caller-controlled and cannot be guaranteed by the type system.
func TestDeterministicCodec_RawNeverSatisfies(t *testing.T) {
	t.Parallel()

	if _, ok := any(codec.RawCodec{}).(codec.DeterministicCodec); ok {
		t.Error("RawCodec must never satisfy DeterministicCodec (passthrough cannot guarantee byte determinism)")
	}
}
