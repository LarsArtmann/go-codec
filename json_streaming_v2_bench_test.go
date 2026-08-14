//go:build goexperiment.jsonv2

package codec_test

import (
	"bytes"
	"encoding/json/v2"
	"testing"

	"github.com/larsartmann/go-codec"
)

// BenchmarkStreamingJSONV2_DecoderComparison measures the cost of two v2 JSON
// streaming approaches for NDJSON: the jsontext.Decoder-based path used by
// codec.NewJSONDecoder (one reader, stateful decoding) versus json.UnmarshalRead
// on each pre-split line (one reader per value). The latter is the alternative
// the codec explicitly avoids because it loses over-read bytes between values.
func BenchmarkStreamingJSONV2_DecoderComparison(b *testing.B) {
	batch := makeOrderBatch()

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)

	for _, evt := range batch {
		if err := enc.Encode(evt); err != nil {
			b.Fatal(err)
		}
	}

	data := buf.Bytes()
	lines := splitNonEmptyLines(data)

	b.ReportAllocs()
	b.ResetTimer()

	b.Run("jsontext.Decoder", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			dec := codec.NewJSONDecoder(bytes.NewReader(data))

			var count int

			for {
				var evt realisticOrder

				if err := dec.Decode(&evt); err != nil {
					break
				}

				count++
			}

			if count != len(batch) {
				b.Fatalf("decoded %d values, want %d", count, len(batch))
			}
		}
	})

	b.Run("json.UnmarshalRead_per_line", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var count int

			for _, line := range lines {
				var evt realisticOrder

				err := json.UnmarshalRead(
					bytes.NewReader(line),
					&evt,
					json.MatchCaseInsensitiveNames(true),
				)
				if err != nil {
					b.Fatal(err)
				}

				count++
			}

			if count != len(batch) {
				b.Fatalf("decoded %d values, want %d", count, len(batch))
			}
		}
	})
}

func splitNonEmptyLines(data []byte) [][]byte {
	var lines [][]byte

	for line := range bytes.SplitSeq(data, []byte{'\n'}) {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	return lines
}
