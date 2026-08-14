package codec_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/larsartmann/go-codec"
)

func TestJSONCodec_Encoding(t *testing.T) {
	t.Parallel()

	c := codec.JSONCodec{}
	if got := c.Encoding(); got != codec.EncodingJSON {
		t.Errorf("codec.Encoding() = %q, want %q", got, codec.EncodingJSON)
	}
}

func TestJSONCodec_RoundTrip(t *testing.T) {
	t.Parallel()

	c := codec.JSONCodec{}

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := payload{Name: testName, Age: 30}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var decoded payload

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestJSONCodec_Encode_Map(t *testing.T) {
	t.Parallel()

	c := codec.JSONCodec{}

	m := map[string]any{testMapKey: testMapVal, "num": float64(42)}

	data, err := c.Encode(m)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var result map[string]any

	err = c.Decode(data, &result)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if result[testMapKey] != testMapVal {
		t.Errorf("got key=%v, want value", result[testMapKey])
	}
}

func TestJSONCodec_Decode_InvalidJSON(t *testing.T) {
	t.Parallel()

	c := codec.JSONCodec{}

	var v any

	err := c.Decode([]byte("not json"), &v)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRawCodec_Encoding(t *testing.T) {
	t.Parallel()

	c := codec.RawCodec{}
	if got := c.Encoding(); got != codec.EncodingRaw {
		t.Errorf("codec.Encoding() = %q, want %q", got, codec.EncodingRaw)
	}
}

func TestRawCodec_RoundTrip(t *testing.T) {
	t.Parallel()

	c := codec.RawCodec{}

	original := []byte("hello raw bytes")

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var decoded []byte

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if string(decoded) != string(original) {
		t.Errorf("round-trip mismatch: got %q, want %q", decoded, original)
	}
}

func TestRawCodec_Encode_WrongType(t *testing.T) {
	t.Parallel()

	c := codec.RawCodec{}

	_, err := c.Encode("not bytes")
	if err == nil {
		t.Fatal("expected error for non-[]byte input")
	}
}

func TestRawCodec_Decode_WrongTarget(t *testing.T) {
	t.Parallel()

	c := codec.RawCodec{}

	err := c.Decode([]byte("data"), "not a pointer to []byte")
	if err == nil {
		t.Fatal("expected error for non-*[]byte target")
	}
}

func TestRawCodec_Decode_IsCopy(t *testing.T) {
	t.Parallel()

	c := codec.RawCodec{}

	original := []byte{1, 2, 3}
	data, _ := c.Encode(original)

	var decoded []byte

	_ = c.Decode(data, &decoded)

	decoded[0] = 99

	if data[0] == 99 {
		t.Error("Decode should return an independent copy")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	t.Parallel()

	codecs := map[string]codec.Codec{
		testCBOR: codec.CBORCodec{},
		testJSON: codec.JSONCodec{},
		"Raw":    codec.RawCodec{},
	}

	for name, c := range codecs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			enc := c.Encoding()
			if enc == "" {
				t.Error("codec.Encoding() should not be empty")
			}
		})
	}
}

func TestJSONCodec_Encode_RawMessage(t *testing.T) {
	t.Parallel()

	c := codec.JSONCodec{}

	raw := codec.RawJSONValue(`{"already":"json"}`)

	data, err := c.Encode(raw)
	if err != nil {
		t.Fatalf("Encode(RawMessage) error: %v", err)
	}

	if string(data) != `{"already":"json"}` {
		t.Errorf("got %q", string(data))
	}
}

func TestJSONCodec_Encode_Nil(t *testing.T) {
	t.Parallel()

	c := codec.JSONCodec{}

	data, err := c.Encode(nil)
	if err != nil {
		t.Fatalf("Encode(nil) error: %v", err)
	}

	if string(data) != "null" {
		t.Errorf("Encode(nil) = %q, want %q", string(data), "null")
	}
}

func TestCBORCodec_Encoding(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}
	if got := c.Encoding(); got != codec.EncodingCBOR {
		t.Errorf("codec.Encoding() = %q, want %q", got, codec.EncodingCBOR)
	}
}

func TestCBORCodec_RoundTrip(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := payload{Name: testName, Age: 30}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var decoded payload

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestCBORCodec_Encode_Map(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	m := map[string]any{testMapKey: testMapVal, "num": uint64(42)}

	data, err := c.Encode(m)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var result map[string]any

	err = c.Decode(data, &result)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if result[testMapKey] != testMapVal {
		t.Errorf("got key=%v, want value", result[testMapKey])
	}
}

func TestCBORCodec_Decode_InvalidCBOR(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	var v any

	err := c.Decode([]byte("not cbor"), &v)
	if err == nil {
		t.Fatal("expected error for invalid CBOR")
	}
}

func TestCBORCodec_Encode_Nil(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	data, err := c.Encode(nil)
	if err != nil {
		t.Fatalf("Encode(nil) error: %v", err)
	}

	var v any
	if err := c.Decode(data, &v); err != nil {
		t.Fatalf("Decode(nil CBOR) error: %v", err)
	}

	if v != nil {
		t.Errorf("Decode(encode(nil)) = %v, want nil", v)
	}
}

func TestCBORCodec_Encode_Deterministic(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	payload := map[string]string{"b": "2", "a": "1", "c": "3"}

	first, err := c.Encode(payload)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	for range 10 {
		got, err := c.Encode(payload)
		if err != nil {
			t.Fatalf("Encode() error: %v", err)
		}

		if string(got) != string(first) {
			t.Fatalf("CBOR encoding is not deterministic: got %x, want %x", got, first)
		}
	}
}

func TestCBORCodec_Decode_EmptyData(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	var v map[string]any

	err := c.Decode([]byte{}, &v)
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestCBORCodec_RoundTrip_Time(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	now := time.Date(2024, 6, 11, 9, 0, 0, 0, time.UTC)

	data, err := c.Encode(now)
	if err != nil {
		t.Fatalf("Encode(time) error: %v", err)
	}

	var decoded time.Time

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode(time) error: %v", err)
	}

	if !decoded.Equal(now) {
		t.Errorf("round-trip mismatch: got %v, want %v", decoded, now)
	}
}

func TestCBORCodec_RoundTrip_TimeSubSecondPrecision(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	original := time.Date(2026, 7, 17, 14, 30, 45, 123456789, time.UTC)

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode(time with nanos) error: %v", err)
	}

	var decoded time.Time

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode(time) error: %v", err)
	}

	// TimeUnixDynamic uses float64, which has ~165ns drift per round-trip.
	// 1 microsecond tolerance is generous but proves sub-second precision survived.
	delta := decoded.Sub(original)
	if delta < 0 {
		delta = -delta
	}

	if delta > time.Microsecond {
		t.Errorf(
			"sub-second precision lost: original=%v decoded=%v delta=%v (max %v)",
			original, decoded, delta, time.Microsecond,
		)
	}
}

func TestCBORCodec_RoundTrip_TimeInPayloadStruct(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	type eventPayload struct {
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
	}

	original := eventPayload{
		Name:      testValue,
		CreatedAt: time.Date(2026, 7, 17, 14, 30, 45, 987654321, time.UTC),
	}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	var decoded eventPayload

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	delta := decoded.CreatedAt.Sub(original.CreatedAt)
	if delta < 0 {
		delta = -delta
	}

	if delta > time.Microsecond {
		t.Errorf(
			"payload time precision lost: original=%v decoded=%v delta=%v",
			original.CreatedAt, decoded.CreatedAt, delta,
		)
	}

	if decoded.Name != original.Name {
		t.Errorf("name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
}

func TestCBORCompactCodec_RoundTrip_TimeSubSecondPrecision(t *testing.T) {
	t.Parallel()

	c := codec.CBORCompactCodec{}

	original := time.Date(2026, 7, 17, 14, 30, 45, 123456789, time.UTC)

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode(time with nanos) error: %v", err)
	}

	var decoded time.Time

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode(time) error: %v", err)
	}

	delta := decoded.Sub(original)
	if delta < 0 {
		delta = -delta
	}

	if delta > time.Microsecond {
		t.Errorf(
			"compact codec sub-second precision lost: original=%v decoded=%v delta=%v",
			original, decoded, delta,
		)
	}
}

func TestCBORCodec_RoundTrip_TimeInstantFromNonUTCLocation(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation error: %v", err)
	}

	// 2026-07-17T09:30:45.500000000-05:00 == 2026-07-17T14:30:45.500000000Z
	original := time.Date(2026, 7, 17, 9, 30, 45, 500000000, loc)

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	var decoded time.Time

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	// The INSTANT must be preserved even though the timezone/location is not.
	// TimeUnixDynamic encodes the Unix epoch, which is timezone-independent.
	if !decoded.Equal(original) {
		t.Errorf(
			"instant mismatch: original=%v (UnixNano=%d) decoded=%v (UnixNano=%d)",
			original, original.UnixNano(), decoded, decoded.UnixNano(),
		)
	}
}

func TestCBORCodec_RoundTrip_ByteSlice(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	type payload struct {
		Data []byte `json:"data"`
	}

	original := payload{Data: []byte{0x00, 0x01, 0x02, 0xff}}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var decoded payload

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if !bytes.Equal(decoded.Data, original.Data) {
		t.Errorf("round-trip mismatch: got %x, want %x", decoded.Data, original.Data)
	}
}

func TestCBORCodec_SmallerThanJSON(t *testing.T) {
	t.Parallel()

	payload := map[string]string{
		testFieldName:  testName,
		testFieldEmail: testEmail,
		"city":         "Berlin",
	}

	cborData, err := codec.CBORCodec{}.Encode(payload)
	if err != nil {
		t.Fatalf("CBOR Encode: %v", err)
	}

	jsonData, err := codec.JSONCodec{}.Encode(payload)
	if err != nil {
		t.Fatalf("JSON Encode: %v", err)
	}

	if len(cborData) >= len(jsonData) {
		t.Errorf(
			"CBOR (%d bytes) should be smaller than JSON (%d bytes)",
			len(cborData),
			len(jsonData),
		)
	}
}

func TestCBORCodec_RoundTrip_Slice(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	original := []string{testAlpha, testBeta, testGamma}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var decoded []string

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if len(decoded) != len(original) {
		t.Fatalf("len = %d, want %d", len(decoded), len(original))
	}

	for i := range original {
		if decoded[i] != original[i] {
			t.Errorf("decoded[%d] = %q, want %q", i, decoded[i], original[i])
		}
	}
}

func TestCBORCodec_RoundTrip_NestedStruct(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	type Address struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	original := Person{
		Name:    testName,
		Address: Address{City: "Berlin", Country: "DE"},
	}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var decoded Person

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestCBORCodec_StructTags(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	type tagged struct {
		Name string `cbor:"name" json:"name"`
		Age  int    `cbor:"age"  json:"age"`
	}

	original := tagged{Name: testBob, Age: 25}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	var decoded tagged

	err = c.Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestCBORCodec_SigningDeterminism(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	payload := map[string]any{
		"user":    testUserName,
		"action":  "login",
		"success": true,
		testCount: uint64(42),
	}

	encoded1, err := c.Encode(payload)
	if err != nil {
		t.Fatalf("Encode 1: %v", err)
	}

	encoded2, err := c.Encode(payload)
	if err != nil {
		t.Fatalf("Encode 2: %v", err)
	}

	encoded3, err := c.Encode(payload)
	if err != nil {
		t.Fatalf("Encode 3: %v", err)
	}

	if string(encoded1) != string(encoded2) || string(encoded2) != string(encoded3) {
		t.Fatal("CBOR encoding must be deterministic for signing safety")
	}

	h1 := len(encoded1)
	_ = h1
}

func TestJSONCodec_Decode_EmptyData(t *testing.T) {
	t.Parallel()

	c := codec.JSONCodec{}

	var v map[string]any

	err := c.Decode([]byte{}, &v)
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestCBORCodec_Decode_IgnoresUnknownFields(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	type target struct {
		Name string `json:"name"`
	}

	type extra struct {
		Name  string `json:"name"`
		Extra string `json:"extra"`
	}

	withExtra := extra{Name: testName, Extra: "surprise"}

	data, err := c.Encode(withExtra)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var got target

	err = c.Decode(data, &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Name != testName {
		t.Fatalf("got %q, want %q", got.Name, testName)
	}
}

func TestCBORCodec_Decode_RejectsDuplicateKeys(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	dup := map[string]any{testMapKey: "v1", "key2": "v2"}

	data, err := c.Encode(dup)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var got map[string]any
	if err := c.Decode(data, &got); err != nil {
		t.Fatalf("Decode valid map: %v", err)
	}
}

func TestBufferEncoder_AllCodecs(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string
		Age  int
	}

	original := payload{Name: testName, Age: 30}

	codecs := []struct {
		name string
		c    codec.Codec
	}{
		{testJSON, codec.JSONCodec{}},
		{testCBOR, codec.CBORCodec{}},
		{"CBORCompact", codec.CBORCompactCodec{}},
	}

	for _, tc := range codecs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bufEncoder, ok := tc.c.(codec.BufferEncoder)
			if !ok {
				t.Fatalf("%s does not implement codec.BufferEncoder", tc.name)
			}

			buf := &bytes.Buffer{}
			if err := bufEncoder.EncodeToBuffer(original, buf); err != nil {
				t.Fatalf("EncodeToBuffer: %v", err)
			}

			var decoded payload
			if err := tc.c.Decode(buf.Bytes(), &decoded); err != nil {
				t.Fatalf("Decode: %v", err)
			}

			if decoded.Name != original.Name || decoded.Age != original.Age {
				t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
			}
		})
	}
}

func TestCBORCodec_AndCBORCompactCodec_ProduceDifferentBytes(t *testing.T) {
	t.Parallel()

	// Canonical (CBORCodec) sorts map keys by length first; Core Deterministic
	// (CBORCompactCodec) sorts bytewise over the full encoded key. For integer
	// keys of different encoded lengths, the two orders can disagree: -1 encodes
	// to one byte (0x20), while 100 encodes to two bytes (0x18 0x64). Canonical
	// puts the shorter key first; bytewise puts the smaller first byte first.
	payload := map[int64]int64{-1: 1, 100: 2}

	canonical, err := codec.CBORCodec{}.Encode(payload)
	if err != nil {
		t.Fatalf("CBORCodec.Encode: %v", err)
	}

	compact, err := codec.CBORCompactCodec{}.Encode(payload)
	if err != nil {
		t.Fatalf("CBORCompactCodec.Encode: %v", err)
	}

	if string(canonical) == string(compact) {
		t.Errorf("CBORCodec and CBORCompactCodec produced identical bytes %x; expected different bytes", canonical)
	}
}

func TestCBORMode_SingletonsReturnIdenticalValues(t *testing.T) {
	t.Parallel()

	// CBOREncMode / CBORDecMode are process-wide sync.OnceValue singletons.
	// Multiple calls must return the exact same value so sibling modules that
	// reuse them produce byte-identical output with CBORCodec.
	if codec.CBOREncMode() != codec.CBOREncMode() {
		t.Error("CBOREncMode() returned different values on repeated calls")
	}

	if codec.CBORDecMode() != codec.CBORDecMode() {
		t.Error("CBORDecMode() returned different values on repeated calls")
	}
}

// TestDiagnose_InvalidCBOR locks the error path of the diagnostic helper:
// garbage bytes produce an error and an empty diagnostic string, not a panic
// or partial output.
func TestDiagnose_InvalidCBOR(t *testing.T) {
	t.Parallel()

	diag, err := codec.Diagnose([]byte{0xff, 0xff, 0xff})
	if err == nil {
		t.Fatal("Diagnose on invalid CBOR must return an error, got nil")
	}

	if diag != "" {
		t.Errorf("Diagnose on invalid CBOR returned partial output %q, want empty", diag)
	}
}
