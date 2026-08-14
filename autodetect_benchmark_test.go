package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// BenchmarkAutoDetect quantifies the cost of the leading-byte heuristic on
// the hot path. Small payloads take the fast path (JSON/CBOR recognized from
// the first byte); an unrecognized byte falls through to trial decode.
func BenchmarkAutoDetect(b *testing.B) {
	cborCodec := codec.CBORCodec{}

	type payload struct {
		Name  string
		Email string
	}

	cborBytes, err := cborCodec.Encode(payload{Name: testName, Email: testEmail})
	if err != nil {
		b.Fatal(err)
	}

	jsonBytes := []byte(`{"name":"Alice","email":"alice@example.com"}`)
	unknownBytes := []byte{0x1f, 0x1f, 0x1f, 0x1f}

	b.Run("json", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			_ = codec.AutoDetect(jsonBytes)
		}
	})

	b.Run("cbor", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			_ = codec.AutoDetect(cborBytes)
		}
	})

	b.Run("unknown", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			_ = codec.AutoDetect(unknownBytes)
		}
	})
}

// BenchmarkAutoDetectDebug quantifies the diagnostic variant: same detection
// plus a formatted Detail string, intended for logging paths — not hot paths.
func BenchmarkAutoDetectDebug(b *testing.B) {
	unknownBytes := []byte{0x1f, 0x1f, 0x1f, 0x1f}

	b.ReportAllocs()

	for b.Loop() {
		_ = codec.AutoDetectDebug(unknownBytes)
	}
}
