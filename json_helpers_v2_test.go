//go:build goexperiment.jsonv2

package codec_test

import "encoding/json/v2"

func testJSONUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
