//go:build goexperiment.jsonv2

package codec_test

// gopls reports stdversion warnings on encoding/json/v2 symbols because the
// module targets go 1.26.5. These are expected: this file is compiled only
// under GOEXPERIMENT=jsonv2, where the v2 import path is available. The
// warnings are inherent to the dual-build pattern and not actionable.
import "encoding/json/v2"

func testJSONUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
