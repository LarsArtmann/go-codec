package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// BenchmarkWrapEncode measures the cost of the self-describing envelope:
// one map allocation with the encoding tag plus the inner codec's Encode.
func BenchmarkWrapEncode(b *testing.B) {
	c := codec.CBORCodec{}

	type payload struct {
		Name  string
		Email string
	}

	p := payload{Name: testName, Email: testEmail}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := codec.WrapEncode(p, c); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnwrapDecode measures envelope unwrapping (encoding lookup plus
// fallback decode of the raw payload bytes).
func BenchmarkUnwrapDecode(b *testing.B) {
	c := codec.CBORCodec{}

	type payload struct {
		Name  string
		Email string
	}

	wrapped, err := codec.WrapEncode(payload{Name: testName, Email: testEmail}, c)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, _ = codec.UnwrapDecode(wrapped, c)
	}
}
