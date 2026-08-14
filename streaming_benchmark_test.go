package codec_test

import (
	"bytes"
	"testing"

	"github.com/larsartmann/go-codec"
)

// streamingBatchSize is the number of events encoded/decoded per benchmark
// iteration. Streaming APIs are designed for batches, not single values, so the
// benchmark measures the per-batch overhead.
const streamingBatchSize = 100

func makeOrderBatch() []realisticOrder {
	order := sampleOrder()

	batch := make([]realisticOrder, 0, streamingBatchSize)

	for range streamingBatchSize {
		batch = append(batch, order)
	}

	return batch
}

func BenchmarkStreamingJSON_Encode(b *testing.B) {
	batch := makeOrderBatch()

	var buf bytes.Buffer

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		buf.Reset()

		enc := codec.NewJSONEncoder(&buf)

		for _, evt := range batch {
			if err := enc.Encode(evt); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkStreamingJSON_Decode(b *testing.B) {
	batch := makeOrderBatch()

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)

	for _, evt := range batch {
		if err := enc.Encode(evt); err != nil {
			b.Fatal(err)
		}
	}

	data := buf.Bytes()

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
}

func BenchmarkStreamingCBOR_Encode(b *testing.B) {
	batch := makeOrderBatch()

	var buf bytes.Buffer

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		buf.Reset()

		enc := codec.NewCBOREncoder(&buf)

		for _, evt := range batch {
			if err := enc.Encode(evt); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkStreamingCBOR_Decode(b *testing.B) {
	batch := makeOrderBatch()

	var buf bytes.Buffer

	enc := codec.NewCBOREncoder(&buf)

	for _, evt := range batch {
		if err := enc.Encode(evt); err != nil {
			b.Fatal(err)
		}
	}

	data := buf.Bytes()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		dec := codec.NewCBORDecoder(bytes.NewReader(data))

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
}
