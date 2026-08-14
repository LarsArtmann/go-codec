package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// BenchmarkNormalizeForJSON measures the overhead of the normalizeForJSON
// path (v1 only) by encoding a map[string]any through JSONCodec. In v2 mode,
// the normalizer is not called — encoding/json/v2 handles map keys natively.
// Run in both modes to compare.
func BenchmarkNormalizeForJSON(b *testing.B) {
	b.ReportAllocs()

	c := codec.JSONCodec{}

	payload := map[string]any{
		testFieldName:  testName,
		testFieldEmail: testEmail,
		testNested: map[string]any{testMapKey: testMapVal},
		"items":    []any{1, "two", true},
	}

	b.ResetTimer()

	for b.Loop() {
		_, err := c.Encode(payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONCodec_MarshalUnmarshal measures the full JSON round-trip.
// Run in both v1 and v2 modes to compare performance.
func BenchmarkJSONCodec_MarshalUnmarshal(b *testing.B) {
	b.ReportAllocs()

	c := codec.JSONCodec{}

	type event struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Version int    `json:"version"`
		Active  bool   `json:"active"`
	}

	payload := event{Name: testName, Email: testEmail, Version: 42, Active: true}

	encoded, err := c.Encode(payload)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		data, err := c.Encode(payload)
		if err != nil {
			b.Fatal(err)
		}

		var decoded event

		if err := c.Decode(data, &decoded); err != nil {
			b.Fatal(err)
		}
	}

	_ = encoded
}
