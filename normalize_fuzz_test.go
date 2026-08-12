//nolint:testpackage // tests internal helpers
//go:build !goexperiment.jsonv2

package codec

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// FuzzNormalizeForJSON verifies that normalizeForJSON never panics and always
// terminates within the depth cap, regardless of input structure. The fuzzer
// generates random CBOR bytes, decodes them into any, then normalizes.
func FuzzNormalizeForJSON(f *testing.F) {
	f.Add([]byte("\xa1\x64name\x65Alice"))                    // simple map
	f.Add([]byte("\x81\x81\x81\x81\x81\x00"))                 // nested arrays
	f.Add([]byte("\xa1aanested\xa1aanested\xa1aanested\x00")) // nested maps

	f.Fuzz(func(t *testing.T, data []byte) {
		var v any
		if err := cbor.Unmarshal(data, &v); err != nil {
			return // skip non-CBOR input
		}

		_, err := normalizeForJSON(v)
		// The function must either succeed or return a depth-cap error.
		// It must never panic or loop indefinitely.
		_ = err
	})
}
