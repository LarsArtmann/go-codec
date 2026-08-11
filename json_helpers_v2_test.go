//go:build goexperiment.jsonv2

package codec_test

import "encoding/json/v2"

func testJSONMarshal(v any) ([]byte, error)      { return json.Marshal(v) }
func testJSONUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
