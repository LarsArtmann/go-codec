//go:build goexperiment.jsonv2

package codec

import (
	"bytes"
	"encoding/json/v2"
	"encoding/json/jsontext"
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
