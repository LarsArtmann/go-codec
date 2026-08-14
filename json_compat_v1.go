//go:build !goexperiment.jsonv2

package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// This file provides JSON helpers backed by encoding/json (v1). The companion
// json_compat_v2.go (build-tagged goexperiment.jsonv2) provides the same
// helpers backed by encoding/json/v2. Consumers get v1 by default; v2 is
// opt-in via GOEXPERIMENT=jsonv2 (Go 1.25+) or the native import on Go 1.27+.

// jsonMarshal marshals v to JSON bytes.
func jsonMarshal(v any) ([]byte, error) {
	normalized, err := normalizeForJSON(v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(normalized) //nolint:wrapcheck // thin wrapper
}

// jsonMarshalDet marshals v to deterministic JSON bytes. In v1, struct fields
// are emitted in declaration order (deterministic for fixed structs) and scalar
// values have a single representation. Map keys are NOT ordered in v1, but all
// current callers marshal structs or scalars, so this is equivalent to v2's
// Deterministic mode for those types.
func jsonMarshalDet(v any) ([]byte, error) {
	normalized, err := normalizeForJSON(v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(normalized) //nolint:wrapcheck // thin wrapper
}

// jsonUnmarshal unmarshals data into v. In v1, struct field matching is
// case-insensitive by default (matching v2's MatchCaseInsensitiveNames option).
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v) //nolint:wrapcheck // thin wrapper
}

// jsonMarshalBuf writes the JSON encoding of v into buf. Unlike
// json.NewEncoder(buf).Encode, it does NOT append a trailing newline.
func jsonMarshalBuf(v any, buf *bytes.Buffer) error {
	normalized, err := normalizeForJSON(v)
	if err != nil {
		return err
	}

	b, err := json.Marshal(normalized)
	if err != nil {
		return err //nolint:wrapcheck // thin wrapper
	}

	_, err = buf.Write(b)
	return err //nolint:wrapcheck // thin wrapper
}

// rawJSONValue is a JSON byte slice that passes through marshalling unchanged.
// In v1 it aliases json.RawMessage.
type rawJSONValue = json.RawMessage

// maxNormalizeDepth bounds the recursion depth of normalizeForJSON to prevent
// stack exhaustion from adversarial deeply-nested CBOR structures.
const maxNormalizeDepth = 100

// normalizeForJSON recursively converts values that encoding/json v1 cannot
// marshal into equivalent forms it can. The primary case is
// map[interface{}]interface{}, which fxamacker/cbor produces when decoding CBOR
// maps into any. encoding/json/v2 handles this natively; v1 does not, so we
// convert map keys to strings here. Non-string keys are formatted via fmt.
//
// A depth cap (maxNormalizeDepth) prevents stack-overflow DoS from
// pathologically nested input.
func normalizeForJSON(v any) (any, error) {
	return normalizeForJSONDepth(v, 0)
}

func normalizeForJSONDepth(v any, depth int) (any, error) {
	if depth > maxNormalizeDepth {
		return nil, fmt.Errorf("%w: %d", ErrNormalizeDepthExceeded, maxNormalizeDepth)
	}

	switch val := v.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]any, len(val))
		for k, item := range val {
			nv, err := normalizeForJSONDepth(item, depth+1)
			if err != nil {
				return nil, err
			}
			out[fmt.Sprint(k)] = nv
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			nv, err := normalizeForJSONDepth(item, depth+1)
			if err != nil {
				return nil, err
			}
			out[k] = nv
		}
		return out, nil
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			nv, err := normalizeForJSONDepth(item, depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = nv
		}
		return out, nil
	default:
		return v, nil
	}
}

// JSONEncoder streams JSON values to an [io.Writer]. Each call to [JSONEncoder.Encode]
// writes one JSON value followed by a newline (NDJSON / JSON Lines convention),
// enabling a reader to consume values incrementally from the stream.
//
// In the v1 build (default), this wraps [json.Encoder].
type JSONEncoder struct {
	enc *json.Encoder
}

func newJSONEncoder(w io.Writer) *JSONEncoder {
	return &JSONEncoder{enc: json.NewEncoder(w)}
}

// Encode writes the JSON encoding of v followed by a newline to the stream.
func (e *JSONEncoder) Encode(v any) error {
	normalized, err := normalizeForJSON(v)
	if err != nil {
		return err
	}

	return e.enc.Encode(normalized) //nolint:wrapcheck // thin wrapper
}

// JSONDecoder reads JSON values from an [io.Reader]. Each call to
// [JSONDecoder.Decode] reads one JSON value from the stream, skipping
// whitespace and newlines between values.
//
// In the v1 build (default), this wraps [json.Decoder].
type JSONDecoder struct {
	dec *json.Decoder
}

func newJSONDecoder(r io.Reader) *JSONDecoder {
	return &JSONDecoder{dec: json.NewDecoder(r)}
}

// Decode reads the next JSON value from the stream into v.
func (d *JSONDecoder) Decode(v any) error {
	return d.dec.Decode(v) //nolint:wrapcheck // thin wrapper
}
