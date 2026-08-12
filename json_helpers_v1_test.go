//go:build !goexperiment.jsonv2

package codec_test

import "encoding/json/v2"

// Test helpers backed by encoding/json (v1). The companion
// json_helpers_v2_test.go (build-tagged goexperiment.jsonv2) provides the
// same helpers backed by encoding/json/v2.

func testJSONMarshal(v any) ([]byte, error)      { return json.Marshal(v) }
func testJSONUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
