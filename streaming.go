package codec

import (
	"io"

	"github.com/fxamacker/cbor/v2"
)

// NewCBOREncoder creates a streaming CBOR encoder that writes to w.
// Use for encoding large event batches without materializing the full
// byte slice in memory. The encoder uses the same canonical encoding mode
// as CBORCodec.
func NewCBOREncoder(w io.Writer) *cbor.Encoder {
	return canonicalEncMode().NewEncoder(w)
}

// NewCBORDecoder creates a streaming CBOR decoder that reads from r.
// Use for decoding large event batches from a stream without loading all
// bytes into memory at once. The decoder uses the same decoding mode as
// CBORCodec.
func NewCBORDecoder(r io.Reader) *cbor.Decoder {
	return canonicalDecMode().NewDecoder(r)
}

// NewJSONEncoder creates a streaming JSON encoder that writes to w.
// Each call to [JSONEncoder.Encode] writes one JSON value followed by a
// newline (NDJSON / JSON Lines convention), enabling incremental consumption
// on the reader side. Use for encoding large event batches without
// materializing the full byte slice in memory.
//
// Parity with [NewCBOREncoder] for consumers that standardize on JSON.
func NewJSONEncoder(w io.Writer) *JSONEncoder {
	return newJSONEncoder(w)
}

// NewJSONDecoder creates a streaming JSON decoder that reads from r.
// Each call to [JSONDecoder.Decode] reads the next JSON value from the
// stream, skipping whitespace and newlines between values. Use for decoding
// large event batches from a stream without loading all bytes into memory.
//
// Parity with [NewCBORDecoder] for consumers that standardize on JSON.
func NewJSONDecoder(r io.Reader) *JSONDecoder {
	return newJSONDecoder(r)
}
