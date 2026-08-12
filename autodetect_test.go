package codec_test

import (
	"bytes"
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestAutoDetect_Empty(t *testing.T) {
	t.Parallel()

	if got := codec.AutoDetect(nil); got != codec.EncodingRaw {
		t.Errorf("codec.AutoDetect(nil) = %q, want %q", got, codec.EncodingRaw)
	}

	if got := codec.AutoDetect([]byte{}); got != codec.EncodingRaw {
		t.Errorf("codec.AutoDetect([]byte{}) = %q, want %q", got, codec.EncodingRaw)
	}
}

func TestAutoDetect_JSON(t *testing.T) {
	t.Parallel()

	cases := [][]byte{
		[]byte(`{"name":"Alice"}`),
		[]byte(`[1,2,3]`),
		[]byte(`"hello"`),
		[]byte(`42`),
		[]byte(`true`),
		[]byte(`null`),
		[]byte(`-3.14`),
	}

	for _, data := range cases {
		got := codec.AutoDetect(data)
		if got != codec.EncodingJSON {
			t.Errorf("codec.AutoDetect(%q) = %q, want %q", data, got, codec.EncodingJSON)
		}
	}
}

func TestAutoDetect_CBOR(t *testing.T) {
	t.Parallel()

	// Encode known values with CBORCodec and verify detection.
	type user struct {
		Name  string
		Email string
	}

	cborData, err := codec.CBORCodec{}.Encode(user{Name: testName, Email: "a@b.c"})
	if err != nil {
		t.Fatalf("CBOR Encode: %v", err)
	}

	if got := codec.AutoDetect(cborData); got != codec.EncodingCBOR {
		t.Errorf("codec.AutoDetect(cborMap) = %q, want %q", got, codec.EncodingCBOR)
	}

	// CBOR array
	arrData, err := codec.CBORCodec{}.Encode([3]int{1, 2, 3})
	if err != nil {
		t.Fatalf("CBOR Encode array: %v", err)
	}

	if got := codec.AutoDetect(arrData); got != codec.EncodingCBOR {
		t.Errorf("codec.AutoDetect(cborArray) = %q, want %q", got, codec.EncodingCBOR)
	}
}

func TestAutoDetect_HighBytesAreCBOR(t *testing.T) {
	t.Parallel()

	// Bytes >= 0x80 are CBOR major types 4-7 and never start valid JSON.
	// AutoDetect returns CBOR even if the full payload is invalid CBOR —
	// it identifies the encoding family, not validity (documented behavior).
	cases := [][]byte{
		{0xff, 0xee, 0xdd}, // major type 7
		{0xa0},             // empty map
		{0x9f},             // stream array start
	}

	for _, data := range cases {
		if got := codec.AutoDetect(data); got != codec.EncodingCBOR {
			t.Errorf("codec.AutoDetect(%v) = %q, want %q (first byte >= 0x80)", data, got, codec.EncodingCBOR)
		}
	}
}

func TestAutoDetect_GenuinelyUnknownIsRaw(t *testing.T) {
	t.Parallel()

	// 0x1f is below 0x80, not a JSON structural start, not a valid JSON
	// token start, and not valid standalone CBOR → AutoDetect returns Raw.
	data := []byte{0x1f}
	if got := codec.AutoDetect(data); got != codec.EncodingRaw {
		t.Errorf("codec.AutoDetect(%v) = %q, want %q", data, got, codec.EncodingRaw)
	}
}

func TestSize(t *testing.T) {
	t.Parallel()

	type user struct {
		Name  string
		Email string
	}

	s := codec.Size(user{Name: testName, Email: testEmail})
	jsonSize, cborSize := s.JSON, s.CBOR

	if jsonSize <= 0 {
		t.Errorf("jsonSize = %d, want > 0", jsonSize)
	}

	if cborSize <= 0 {
		t.Errorf("cborSize = %d, want > 0", cborSize)
	}

	// CBOR should be smaller for this simple struct
	if cborSize >= jsonSize {
		t.Errorf("cborSize %d >= jsonSize %d, expected CBOR to be smaller", cborSize, jsonSize)
	}
}

func TestSize_EncodeError(t *testing.T) {
	t.Parallel()

	// chan cannot be encoded by either codec
	s := codec.Size(make(chan int))
	jsonSize, cborSize := s.JSON, s.CBOR

	if jsonSize != -1 {
		t.Errorf("jsonSize = %d, want -1 for unencodable value", jsonSize)
	}

	if cborSize != -1 {
		t.Errorf("cborSize = %d, want -1 for unencodable value", cborSize)
	}
}

func TestAutoDetectDebug_Reasons(t *testing.T) {
	t.Parallel()

	cborData, err := codec.CBORCodec{}.Encode(map[string]string{testField: testName})
	if err != nil {
		t.Fatalf("CBOR Encode: %v", err)
	}

	maxSize := 1 << 20 // 1 MiB, matches maxAutoDetectSize

	cases := []struct {
		name    string
		data    []byte
		wantEnc codec.Encoding
		wantRsn codec.DetectionReason
	}{
		{
			name:    "cbor_major_type_map",
			data:    cborData,
			wantEnc: codec.EncodingCBOR,
			wantRsn: codec.DetectionReasonCBORMajorType,
		},
		{
			name:    "empty_nil",
			data:    nil,
			wantEnc: codec.EncodingRaw,
			wantRsn: codec.DetectionReasonEmpty,
		},
		{
			name:    "empty_slice",
			data:    []byte{},
			wantEnc: codec.EncodingRaw,
			wantRsn: codec.DetectionReasonEmpty,
		},
		{
			name:    "cbor_major_type_high_byte",
			data:    []byte{0xff, 0xee, 0xdd},
			wantEnc: codec.EncodingCBOR,
			wantRsn: codec.DetectionReasonCBORMajorType,
		},
		{
			name:    "json_structure_object",
			data:    []byte(`{"name":"Alice"}`),
			wantEnc: codec.EncodingJSON,
			wantRsn: codec.DetectionReasonJSONStructure,
		},
		{
			name:    "json_structure_array",
			data:    []byte(`[1,2,3]`),
			wantEnc: codec.EncodingJSON,
			wantRsn: codec.DetectionReasonJSONStructure,
		},
		{
			name:    "json_structure_string",
			data:    []byte(`"hello"`),
			wantEnc: codec.EncodingJSON,
			wantRsn: codec.DetectionReasonJSONStructure,
		},
		{
			name:    "json_trial_decode_number",
			data:    []byte(`42`),
			wantEnc: codec.EncodingJSON,
			wantRsn: codec.DetectionReasonJSONTrialDecode,
		},
		{
			name:    "json_trial_decode_true",
			data:    []byte(`true`),
			wantEnc: codec.EncodingJSON,
			wantRsn: codec.DetectionReasonJSONTrialDecode,
		},
		{
			name:    "json_trial_decode_fails_returns_cbor",
			data:    []byte{0x2d}, // '-' ASCII, but valid CBOR negative integer
			wantEnc: codec.EncodingCBOR,
			wantRsn: codec.DetectionReasonCBORMajorType,
		},
		{
			name:    "cbor_trial_decode",
			data:    []byte{0x00}, // CBOR unsigned integer 0, first byte below CBOR major-type range
			wantEnc: codec.EncodingCBOR,
			wantRsn: codec.DetectionReasonCBORTrialDecode,
		},
		{
			name:    "oversized_json_start",
			data:    append([]byte{'1'}, bytes.Repeat([]byte{'0'}, maxSize)...),
			wantEnc: codec.EncodingRaw,
			wantRsn: codec.DetectionReasonOversized,
		},
		{
			name:    "unknown",
			data:    []byte{0x1f},
			wantEnc: codec.EncodingRaw,
			wantRsn: codec.DetectionReasonUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := codec.AutoDetectDebug(tc.data)

			if got.Encoding != tc.wantEnc {
				t.Errorf("Encoding = %q, want %q", got.Encoding, tc.wantEnc)
			}

			if got.Reason != tc.wantRsn {
				t.Errorf("Reason = %q, want %q", got.Reason, tc.wantRsn)
			}

			if got.Detail == "" {
				t.Error("Detail is empty, want non-empty diagnostic string")
			}
		})
	}
}

func TestAutoDetectDebug_WrapsAutoDetect(t *testing.T) {
	t.Parallel()

	for _, data := range [][]byte{
		nil,
		{},
		[]byte(`{"a":1}`),
		{0x80},
		{0x1f},
	} {
		if got := codec.AutoDetect(data); got != codec.AutoDetectDebug(data).Encoding {
			t.Errorf("AutoDetect(%v) != AutoDetectDebug(%v).Encoding", data, data)
		}
	}
}
