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

// BenchmarkUnwrapDecode_FallbackRawCBOR measures the backward-compat path:
// raw (unenveloped) CBOR data must fall through to the fallback codec without
// paying a doomed JSON envelope parse. This is the hot path for blind stores
// reading pre-envelope data.
func BenchmarkUnwrapDecode_FallbackRawCBOR(b *testing.B) {
	c := codec.CBORCodec{}

	type payload struct {
		Name  string
		Email string
	}

	raw, err := c.Encode(payload{Name: testName, Email: testEmail})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, _ = codec.UnwrapDecode(raw, c)
	}
}
