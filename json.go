package codec

import "bytes"

// JSONCodec implements Codec using encoding/json.
type JSONCodec struct{}

var _ Codec = JSONCodec{}

func (JSONCodec) Encoding() Encoding { return EncodingJSON }

// Encode marshals a value to JSON bytes.
func (JSONCodec) Encode(v any) ([]byte, error) {
	return jsonMarshal(v)
}

// Decode unmarshals JSON bytes into a value.
func (JSONCodec) Decode(data []byte, v any) error {
	return jsonUnmarshal(data, v)
}

// EncodeToBuffer writes JSON encoding of v directly into buf,
// avoiding the allocation that Encode returns. Implements BufferEncoder.
func (JSONCodec) EncodeToBuffer(v any, buf *bytes.Buffer) error {
	return jsonMarshalBuf(v, buf)
}
