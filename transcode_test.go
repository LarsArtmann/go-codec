package codec_test

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestTranscodeToJSON_CBOR_Map(t *testing.T) {
	t.Parallel()

	in := map[string]any{testFieldName: testUserName, testCount: 42.0}

	cborData, err := (codec.CBORCodec{}).Encode(in)
	if err != nil {
		t.Fatalf("encode CBOR: %v", err)
	}

	out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	var got map[string]any
	if err := codec.JSONUnmarshal(out, &got); err != nil {
		t.Fatalf("decode JSON: %v\njson: %s", err, out)
	}

	if got[testFieldName] != testUserName {
		t.Errorf("name = %v, want alice", got[testFieldName])
	}

	if got[testCount] != float64(42) {
		t.Errorf("count = %v, want 42", got[testCount])
	}
}

func TestTranscodeToJSON_CBOR_ToArrayStruct_StaysArray(t *testing.T) {
	t.Parallel()

	// toarray-encoded structs decode to CBOR arrays, not maps. TranscodeToJSON
	// is schema-free, so the array is preserved on the JSON side. This documents
	// the boundary: generic transcoding cannot reconstruct field names.
	type user struct {
		_     struct{} `cbor:",toarray"`
		Name  string
		Email string
	}

	cborData, err := (codec.CBORCodec{}).Encode(user{Name: testUserName, Email: "a@b.com"})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	var got []any
	if err := codec.JSONUnmarshal(out, &got); err != nil {
		t.Fatalf("decode JSON: %v\njson: %s", err, out)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2; json: %s", len(got), out)
	}

	if got[0] != testUserName || got[1] != "a@b.com" {
		t.Errorf("got = %v, want [alice a@b.com]", got)
	}
}

func TestTranscodeToJSON_JSON_Passthrough(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"already":"json"}`)

	out, err := codec.TranscodeToJSON(payload, codec.EncodingJSON)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	if string(out) != string(payload) {
		t.Errorf("JSON should pass through unchanged; got %q, want %q", out, payload)
	}
}

func TestTranscodeToJSON_Raw_Passthrough(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"maybe":"json-or-not"}`)

	out, err := codec.TranscodeToJSON(payload, codec.EncodingRaw)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	if string(out) != string(payload) {
		t.Errorf("Raw should pass through unchanged; got %q, want %q", out, payload)
	}
}

func TestTranscodeToJSON_InvalidCBOR_Error(t *testing.T) {
	t.Parallel()

	// 0xa1 = CBOR map of 1 pair, but the trailing key/value are garbage.
	badCBOR := []byte{0xa1, 0xff, 0xff}

	_, err := codec.TranscodeToJSON(badCBOR, codec.EncodingCBOR)
	if err == nil {
		t.Fatal("expected error for invalid CBOR, got nil")
	}

	if !strings.Contains(err.Error(), "transcode") {
		t.Errorf("error should mention transcode context; got %q", err)
	}
}

func TestTranscodeToJSON_NestedAndScalars(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		testNested: map[string]any{"deep": []any{1, "two", true}},
		"flag":     true,
		"none":     nil,
	}

	cborData, err := (codec.CBORCodec{}).Encode(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	var got map[string]any
	if err := codec.JSONUnmarshal(out, &got); err != nil {
		t.Fatalf("decode JSON: %v\njson: %s", err, out)
	}

	nested, ok := got[testNested].(map[string]any)
	if !ok {
		t.Fatalf("nested not a map: %T", got[testNested])
	}

	deep, ok := nested["deep"].([]any)
	if !ok || len(deep) != 3 {
		t.Fatalf("deep = %v", nested["deep"])
	}

	if got["flag"] != true {
		t.Errorf("flag = %v, want true", got["flag"])
	}

	if got["none"] != nil {
		t.Errorf("none = %v, want nil", got["none"])
	}
}

func TestTranscodeToJSON_CBORCompactCodec(t *testing.T) {
	t.Parallel()

	// CBORCompactCodec reports EncodingCBOR too (ADR-CODEC). Its wire format
	// must be decodable by the canonical decoder TranscodeToJSON uses, so the
	// two CBOR variants share one transcode path.
	in := map[string]any{testFieldName: "compact", testCount: 7.0}

	cborData, err := (codec.CBORCompactCodec{}).Encode(in)
	if err != nil {
		t.Fatalf("encode compact CBOR: %v", err)
	}

	out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	var got map[string]any
	if err := codec.JSONUnmarshal(out, &got); err != nil {
		t.Fatalf("decode JSON: %v\njson: %s", err, out)
	}

	if got[testFieldName] != "compact" {
		t.Errorf("name = %v, want compact", got[testFieldName])
	}

	if got[testCount] != float64(7) {
		t.Errorf("count = %v, want 7", got[testCount])
	}
}

// TestTranscodeToJSON_LargeNumbers exercises values at and beyond the int64
// and uint64 range. CBOR encodes uint64 values exceeding int64 max as a
// bignum (tag 2); values beyond uint64 require *big.Int. The schema-free
// generic decode path must still yield valid JSON without panicking. This
// documents the boundary for the generic decode (#20).
func TestTranscodeToJSON_LargeNumbers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		val  any
	}{
		{"int64_max", uint64(1<<63 - 1)},
		{"uint64_max", ^uint64(0)},
		{"bignum_over_uint64", new(big.Int).Lsh(big.NewInt(1), 70)}, // 2^70, a real CBOR bignum
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cborData, err := (codec.CBORCodec{}).Encode(map[string]any{"n": tc.val})
			if err != nil {
				t.Fatalf("encode: %v", err)
			}

			out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
			if err != nil {
				t.Fatalf("transcode: %v", err)
			}

			// The output must be valid JSON. Numeric representation varies by
			// generic-decode path (float64 vs decimal number); we only assert
			// validity + key presence, since float64 loses precision above 2^53.
			var got map[string]any
			if err := codec.JSONUnmarshal(out, &got); err != nil {
				t.Fatalf("invalid JSON: %v\nraw: %s", err, out)
			}

			if _, ok := got["n"]; !ok {
				t.Errorf("key n missing from %s", out)
			}
		})
	}
}

// TestTranscodeToJSON_EmptyContainers verifies that empty CBOR maps and arrays
// transcode to their JSON equivalents ("{}" and "[]").
func TestTranscodeToJSON_EmptyContainers(t *testing.T) {
	t.Parallel()

	t.Run("empty_map", func(t *testing.T) {
		t.Parallel()

		cborData, err := (codec.CBORCodec{}).Encode(map[string]any{})
		if err != nil {
			t.Fatalf("encode: %v", err)
		}

		out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
		if err != nil {
			t.Fatalf("transcode: %v", err)
		}

		if string(out) != "{}" {
			t.Errorf("empty map = %q, want {}", out)
		}
	})

	t.Run("empty_array", func(t *testing.T) {
		t.Parallel()

		cborData, err := (codec.CBORCodec{}).Encode([]any{})
		if err != nil {
			t.Fatalf("encode: %v", err)
		}

		out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
		if err != nil {
			t.Fatalf("transcode: %v", err)
		}

		if string(out) != "[]" {
			t.Errorf("empty array = %q, want []", out)
		}
	})
}

// TestTranscodeToJSON_MapKeysRoundTrip documents the key-ordering reality of
// the generic transcode path (#23): CBOR canonical encoding sorts map keys on
// the wire, but the generic decode produces a Go map[string]any whose keys are
// then re-encoded by jsonMarshal. Under encoding/json the output key order
// is NOT guaranteed to be sorted or stable across runs (Go map iteration is
// randomized). So this test asserts key presence + values, not byte order.
// Callers needing deterministic key order must use event.DecodePayloadAuto[T]
// with a concrete struct type, or sort the JSON keys themselves.
func TestTranscodeToJSON_MapKeysRoundTrip(t *testing.T) {
	t.Parallel()

	in := map[string]any{"zebra": 1, "apple": 2, "mango": 3}

	cborData, err := (codec.CBORCodec{}).Encode(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	var got map[string]any
	if err := codec.JSONUnmarshal(out, &got); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, out)
	}

	want := map[string]any{"zebra": float64(1), "apple": float64(2), "mango": float64(3)}
	for k, wantV := range want {
		gotV, ok := got[k]
		if !ok {
			t.Errorf("key %q missing from %s", k, out)

			continue
		}

		if gotV != wantV {
			t.Errorf("key %q = %v, want %v", k, gotV, wantV)
		}
	}
}

// TestTranscodeToJSON_ByteSliceAsBase64 verifies that CBOR byte strings (major
// type 2) transcode to JSON as base64-encoded strings, matching encoding/json's
// default []byte handling (#22).
func TestTranscodeToJSON_ByteSliceAsBase64(t *testing.T) {
	t.Parallel()

	in := map[string]any{"data": []byte("hello world")}

	cborData, err := (codec.CBORCodec{}).Encode(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("transcode: %v", err)
	}

	var got map[string]any
	if err := codec.JSONUnmarshal(out, &got); err != nil {
		t.Fatalf("decode JSON: %v\njson: %s", err, out)
	}

	// encoding/json marshals []byte as base64 standard encoding.
	data, ok := got["data"].(string)
	if !ok {
		t.Fatalf("data = %T, want string (base64); json: %s", got["data"], out)
	}

	if data != "aGVsbG8gd29ybGQ=" {
		t.Errorf("base64 = %q, want aGVsbG8gd29ybGQ=", data)
	}
}

// TestTranscodeToJSON_FloatSpecials verifies the behavior when CBOR contains
// NaN, +Inf, or -Inf float values (#23). JSON does not support these values,
// so TranscodeToJSON should return an error rather than producing invalid JSON.
func TestTranscodeToJSON_FloatSpecials(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		val  float64
	}{
		{"nan", math.NaN()},
		{"pos_inf", math.Inf(1)},
		{"neg_inf", math.Inf(-1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cborData, err := (codec.CBORCodec{}).Encode(map[string]any{"v": tc.val})
			if err != nil {
				t.Fatalf("encode: %v", err)
			}

			out, err := codec.TranscodeToJSON(cborData, codec.EncodingCBOR)
			if err != nil {
				// Error is acceptable: JSON cannot represent NaN/Inf.
				return
			}

			// If it didn't error, the output must still be valid JSON
			// (some json encoders emit null for these).
			var probe any
			if perr := codec.JSONUnmarshal(out, &probe); perr != nil {
				t.Errorf("produced invalid JSON for %s: %v\nraw: %s", tc.name, perr, out)
			}
		})
	}
}

// TestTranscodeToJSON_DuplicateMapKeys verifies that duplicate CBOR map keys
// are rejected by the decoder (#24). The canonical decoder uses
// DupMapKeyEnforcedAPF, so decoding CBOR with duplicate keys yields an error.
func TestTranscodeToJSON_DuplicateMapKeys(t *testing.T) {
	t.Parallel()

	// Raw CBOR: map of 2 pairs with duplicate key "a".
	// 0xa2 = map(2), 0x61 0x61 = str("a"), 0x01 = int(1),
	// 0x61 0x61 = str("a"), 0x02 = int(2).
	dupKeys := []byte{0xa2, 0x61, 0x61, 0x01, 0x61, 0x61, 0x02}

	_, err := codec.TranscodeToJSON(dupKeys, codec.EncodingCBOR)
	if err == nil {
		t.Fatal("expected error for duplicate map keys, got nil")
	}

	if !strings.Contains(err.Error(), "transcode") {
		t.Errorf("error should mention transcode context; got %q", err)
	}
}

// TestTranscodeToJSON_CBORTag0 verifies that CBOR tag 0 (standard date/time
// string) decodes to a time.Time in the generic path, which json.Marshal
// renders as an RFC3339 string. The output is a JSON string, not an object.
func TestTranscodeToJSON_CBORTag0(t *testing.T) {
	t.Parallel()

	// Raw CBOR: tag 0 (0xc0) wrapping an RFC3339 string.
	// 0xc0 = tag 0, 0x74 = str(20), "2026-07-27T00:00:00Z"
	tagged := append([]byte{0xc0, 0x74},
		[]byte("2026-07-27T00:00:00Z")...)

	out, err := codec.TranscodeToJSON(tagged, codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("transcode tag 0: %v", err)
	}

	// Tag 0 decodes to time.Time → JSON string "2026-07-27T00:00:00Z".
	var got string
	if err := codec.JSONUnmarshal(out, &got); err != nil {
		t.Fatalf("decode JSON: %v\nraw: %s", err, out)
	}

	if got != "2026-07-27T00:00:00Z" {
		t.Errorf("tag 0 = %q, want 2026-07-27T00:00:00Z", got)
	}
}
