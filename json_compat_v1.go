//go:build !goexperiment.jsonv2

package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// This file provides JSON helpers backed by encoding/json (v1). The companion
// json_compat_v2.go (build-tagged goexperiment.jsonv2) provides the same
// helpers backed by encoding/json/v2. Consumers get v1 by default; v2 is
// opt-in via GOEXPERIMENT=jsonv2 (Go 1.25+) or the native import on Go 1.27+.

// jsonMarshal marshals v to JSON bytes.
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(normalizeForJSON(v)) //nolint:wrapcheck // thin wrapper
}

// jsonMarshalDet marshals v to deterministic JSON bytes. In v1, struct fields
// are emitted in declaration order (deterministic for fixed structs) and scalar
// values have a single representation. Map keys are NOT ordered in v1, but all
// current callers marshal structs or scalars, so this is equivalent to v2's
// Deterministic mode for those types.
func jsonMarshalDet(v any) ([]byte, error) {
	return json.Marshal(normalizeForJSON(v)) //nolint:wrapcheck // thin wrapper
}

// jsonUnmarshal unmarshals data into v. In v1, struct field matching is
// case-insensitive by default (matching v2's MatchCaseInsensitiveNames option).
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v) //nolint:wrapcheck // thin wrapper
}

// jsonMarshalBuf writes the JSON encoding of v into buf. Unlike
// json.NewEncoder(buf).Encode, it does NOT append a trailing newline.
func jsonMarshalBuf(v any, buf *bytes.Buffer) error {
	b, err := json.Marshal(normalizeForJSON(v))
	if err != nil {
		return err
	}

	_, err = buf.Write(b)
	return err
}

// rawJSONValue is a JSON byte slice that passes through marshalling unchanged.
// In v1 it aliases json.RawMessage.
type rawJSONValue = json.RawMessage

// normalizeForJSON recursively converts values that encoding/json v1 cannot
// marshal into equivalent forms it can. The primary case is
// map[interface{}]interface{}, which fxamacker/cbor produces when decoding CBOR
// maps into any. encoding/json/v2 handles this natively; v1 does not, so we
// convert map keys to strings here. Non-string keys are formatted via fmt.
func normalizeForJSON(v any) any {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]any, len(val))
		for k, v := range val {
			out[fmt.Sprint(k)] = normalizeForJSON(v)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v := range val {
			out[k] = normalizeForJSON(v)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = normalizeForJSON(v)
		}
		return out
	default:
		return v
	}
}
