//go:build !goexperiment.jsonv2

package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// TestDeterministicCodec_JSONCodecV1DoesNotSatisfy locks the documented
// contract for the DEFAULT build: encoding/json v1 iterates map keys in
// non-deterministic order, so JSONCodec.Encode is not byte-deterministic and
// must NOT satisfy DeterministicCodec. Signing a v1 JSON payload is the
// silent signature-corruption bug the marker interface exists to prevent.
//
// This is a runtime assertion (not a compile-time `var _` line) on purpose:
// the compile-time form cannot express a negative claim.
func TestDeterministicCodec_JSONCodecV1DoesNotSatisfy(t *testing.T) {
	t.Parallel()

	if _, ok := any(codec.JSONCodec{}).(codec.DeterministicCodec); ok {
		t.Error("JSONCodec must NOT satisfy DeterministicCodec in the v1 build " +
			"(encoding/json v1 map key order is non-deterministic)")
	}
}
