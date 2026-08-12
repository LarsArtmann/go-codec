//go:build goexperiment.jsonv2

package codec

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io"
)

// This file provides JSON helpers backed by encoding/json/v2, activated when
// GOEXPERIMENT=jsonv2 is set (Go 1.25+) or on Go 1.27+ where the import path
// exists natively. The companion json_compat_v1.go provides the default path.

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v) //nolint:wrapcheck // thin wrapper
}

func jsonMarshalDet(v any) ([]byte, error) {
	return json.Marshal(v, json.Deterministic(true)) //nolint:wrapcheck // thin wrapper
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v, json.MatchCaseInsensitiveNames(true)) //nolint:wrapcheck
}

func jsonMarshalBuf(v any, buf *bytes.Buffer) error {
	return json.MarshalWrite(buf, v) //nolint:wrapcheck // thin wrapper
}

type rawJSONValue = jsontext.Value

// JSONEncoder streams JSON values to an [io.Writer]. Each call to
// [JSONEncoder.Encode] writes one JSON value followed by a newline
// (NDJSON / JSON Lines convention), enabling a reader to consume values
// incrementally from the stream.
//
// In the v2 build (GOEXPERIMENT=jsonv2), this uses [json.MarshalWrite] per
// value. [jsontext.Encoder] is intentionally NOT used because it inserts
// separator tokens between top-level values, corrupting NDJSON output.
type JSONEncoder struct {
	w io.Writer
}

func newJSONEncoder(w io.Writer) *JSONEncoder {
	return &JSONEncoder{w: w}
}

// Encode writes the JSON encoding of v followed by a newline to the stream.
func (e *JSONEncoder) Encode(v any) error {
	if err := json.MarshalWrite(e.w, v); err != nil {
		return err //nolint:wrapcheck // thin wrapper
	}

	_, err := e.w.Write([]byte{'\n'})

	return err //nolint:wrapcheck // thin wrapper
}

// JSONDecoder reads JSON values from an [io.Reader]. Each call to
// [JSONDecoder.Decode] reads one JSON value from the stream, skipping
// whitespace and newlines between values.
//
// In the v2 build (GOEXPERIMENT=jsonv2), this wraps [jsontext.Decoder] with
// [json.UnmarshalDecode] for stateful streaming. [json.UnmarshalRead] is NOT
// used because it creates a new internal buffer per call, losing over-read
// bytes between values.
type JSONDecoder struct {
	dec *jsontext.Decoder
}

func newJSONDecoder(r io.Reader) *JSONDecoder {
	return &JSONDecoder{dec: jsontext.NewDecoder(r)}
}

// Decode reads the next JSON value from the stream into v.
func (d *JSONDecoder) Decode(v any) error {
	return json.UnmarshalDecode(d.dec, v, json.MatchCaseInsensitiveNames(true)) //nolint:wrapcheck
}
