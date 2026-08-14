//go:build goexperiment.jsonv2

package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// TestDeterministicCodec_JSONCodecV2Satisfies locks the documented contract
// for the GOEXPERIMENT=jsonv2 build: encoding/json/v2 runs with
// json.Deterministic, so JSONCodec.Encode is byte-deterministic and DOES
// satisfy DeterministicCodec.
func TestDeterministicCodec_JSONCodecV2Satisfies(t *testing.T) {
	t.Parallel()

	if _, ok := any(codec.JSONCodec{}).(codec.DeterministicCodec); !ok {
		t.Error("JSONCodec must satisfy DeterministicCodec in the v2 build " +
			"(encoding/json/v2 with json.Deterministic)")
	}
}
