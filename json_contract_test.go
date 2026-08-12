package codec

import (
	"os"
	"strings"
	"testing"
)

// The dual-JSON architecture relies on two build-tagged file pairs. A recurring
// hazard is that goimports, when re-adding the json import, picks
// encoding/json/v2 for the v1 file — silently breaking the default build mode.
// These tests lock down the contract so any drift fails loudly in both CI modes.

const (
	jsonV1ImplFile      = "json_compat_v1.go"
	jsonV2ImplFile      = "json_compat_v2.go"
	jsonV1HelperTest    = "json_helpers_v1_test.go"
	jsonV2HelperTest    = "json_helpers_v2_test.go"
	jsonV1ImportPath    = "\"encoding/json\""
	jsonV2ImportPath    = "\"encoding/json/v2\""
	jsonV1BuildConstrnt = "//go:build !goexperiment.jsonv2"
	jsonV2BuildConstrnt = "//go:build goexperiment.jsonv2"
)

func readSource(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// TestDualJSONContract_Imports locks the import split: the v1 files MUST import
// encoding/json (never v2) and the v2 files MUST import encoding/json/v2.
func TestDualJSONContract_Imports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		filename    string
		mustContain string
		mustOmit    string
	}{
		{"v1 impl uses json v1 import", jsonV1ImplFile, jsonV1ImportPath, jsonV2ImportPath},
		{"v1 helper uses json v1 import", jsonV1HelperTest, jsonV1ImportPath, jsonV2ImportPath},
		{"v2 impl uses json v2 import", jsonV2ImplFile, jsonV2ImportPath, jsonV1ImportPath},
		{"v2 helper uses json v2 import", jsonV2HelperTest, jsonV2ImportPath, jsonV1ImportPath},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content, err := readSource(tc.filename)
			if err != nil {
				t.Fatalf("read %s: %v", tc.filename, err)
			}

			if !strings.Contains(content, tc.mustContain) {
				t.Errorf("%s must contain %s", tc.filename, tc.mustContain)
			}

			if strings.Contains(content, tc.mustOmit) {
				t.Errorf("%s must NOT contain %s", tc.filename, tc.mustOmit)
			}
		})
	}
}

// TestDualJSONContract_BuildTags locks the build constraints.
func TestDualJSONContract_BuildTags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		filename string
		want     string
	}{
		{jsonV1ImplFile, jsonV1BuildConstrnt},
		{jsonV2ImplFile, jsonV2BuildConstrnt},
		{jsonV1HelperTest, jsonV1BuildConstrnt},
		{jsonV2HelperTest, jsonV2BuildConstrnt},
	}

	for _, tc := range cases {
		t.Run(tc.filename, func(t *testing.T) {
			t.Parallel()

			content, err := readSource(tc.filename)
			if err != nil {
				t.Fatalf("read %s: %v", tc.filename, err)
			}

			if !strings.Contains(content, tc.want) {
				t.Errorf("%s must contain build constraint %s", tc.filename, tc.want)
			}
		})
	}
}
