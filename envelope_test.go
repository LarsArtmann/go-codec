package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestWrapEncode_JSON_RoundTrip(t *testing.T) {
	t.Parallel()

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	original := user{Name: testName, Email: testEmail}

	wrapped, err := codec.WrapEncode(original, codec.JSONCodec{})
	if err != nil {
		t.Fatalf("codec.WrapEncode: %v", err)
	}

	c, inner := codec.UnwrapDecode(wrapped, codec.CBORCodec{})
	if c.Encoding() != codec.EncodingJSON {
		t.Fatalf("expected json codec, got %s", c.Encoding())
	}

	var decoded user
	if err := c.Decode(inner, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded != original {
		t.Fatalf("got %+v, want %+v", decoded, original)
	}
}

func TestWrapEncode_CBOR_RoundTrip(t *testing.T) {
	t.Parallel()

	type item struct {
		SKU   string `json:"sku"`
		Price int    `json:"price"`
	}

	original := item{SKU: "WIDGET-001", Price: 4999}

	wrapped, err := codec.WrapEncode(original, codec.CBORCodec{})
	if err != nil {
		t.Fatalf("codec.WrapEncode: %v", err)
	}

	c, inner := codec.UnwrapDecode(wrapped, codec.JSONCodec{})
	if c.Encoding() != codec.EncodingCBOR {
		t.Fatalf("expected cbor codec, got %s", c.Encoding())
	}

	var decoded item
	if err := c.Decode(inner, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded != original {
		t.Fatalf("got %+v, want %+v", decoded, original)
	}
}

func TestUnwrapDecode_BackwardCompat_RawJSON(t *testing.T) {
	t.Parallel()

	// Simulate old unenveloped JSON data.
	rawJSON := []byte(`{"name":"Bob","email":"bob@example.com"}`)
	fallback := codec.JSONCodec{}

	c, inner := codec.UnwrapDecode(rawJSON, fallback)
	if c.Encoding() != codec.EncodingJSON {
		t.Fatalf("expected fallback json codec, got %s", c.Encoding())
	}

	if string(inner) != string(rawJSON) {
		t.Fatalf("inner data should be unchanged for raw data")
	}

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var decoded user
	if err := c.Decode(inner, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.Name != testBob {
		t.Fatalf("got name %q, want %q", decoded.Name, testBob)
	}
}

func TestUnwrapDecode_BackwardCompat_RawCBOR(t *testing.T) {
	t.Parallel()

	type item struct {
		SKU string `json:"sku"`
	}

	original := item{SKU: "CBOR-RAW"}

	rawCBOR, err := (codec.CBORCodec{}).Encode(original)
	if err != nil {
		t.Fatalf("encode raw cbor: %v", err)
	}

	// Old CBOR data without codec.Envelope — should fall back to provided codec.
	c, inner := codec.UnwrapDecode(rawCBOR, codec.CBORCodec{})
	if c.Encoding() != codec.EncodingCBOR {
		t.Fatalf("expected fallback cbor codec, got %s", c.Encoding())
	}

	if string(inner) != string(rawCBOR) {
		t.Fatalf("inner data should be unchanged for raw data")
	}
}

func TestUnwrapDecode_NonJSONData(t *testing.T) {
	t.Parallel()

	// Random non-JSON bytes should fall back gracefully.
	weird := []byte{0x00, 0x01, 0x02, 0xFF}
	fallback := codec.JSONCodec{}

	c, inner := codec.UnwrapDecode(weird, fallback)
	if c.Encoding() != codec.EncodingJSON {
		t.Fatalf("expected fallback codec")
	}

	if string(inner) != string(weird) {
		t.Fatalf("data should be unchanged")
	}
}

func TestUnwrapDecode_FirstByteSniff(t *testing.T) {
	t.Parallel()

	fallback := codec.JSONCodec{}

	// Every first byte >= 0x80 can never begin valid JSON, so the envelope
	// parse must be skipped entirely — even when the tail is crafted to look
	// envelope-shaped.
	crafted := append([]byte{0xA2}, []byte(`{"$":"gcdc","enc":"json","dat":"eA=="}`)...)

	c, inner := codec.UnwrapDecode(crafted, fallback)
	if c.Encoding() != codec.EncodingJSON {
		t.Fatalf("expected fallback codec for CBOR-prefixed data")
	}

	if string(inner) != string(crafted) {
		t.Fatalf("data should be unchanged")
	}

	for b := byte(0x80); b != 0; b++ { // wraps at 0xFF -> 0x00 terminates
		data := []byte{b, '{', '"', '$'}

		c, inner := codec.UnwrapDecode(data, fallback)
		if c.Encoding() != codec.EncodingJSON {
			t.Fatalf("first byte 0x%02x: expected fallback codec", b)
		}

		if string(inner) != string(data) {
			t.Fatalf("first byte 0x%02x: data should be unchanged", b)
		}
	}
}

func TestUnwrapDecode_EmptyData(t *testing.T) {
	t.Parallel()

	fallback := codec.CBORCodec{}

	c, inner := codec.UnwrapDecode(nil, fallback)
	if c.Encoding() != codec.EncodingCBOR {
		t.Fatalf("expected fallback codec for empty data")
	}

	if len(inner) != 0 {
		t.Fatalf("data should be unchanged")
	}
}

func TestUnwrapDecode_RawCBORScalarsBelowSniffThreshold(t *testing.T) {
	t.Parallel()

	// CBOR text strings (0x60-0x7f) and ints (0x00-0x1b) sit below the 0x80
	// sniff threshold: they stay on the parse-fail path but must still fall
	// back unchanged.
	fallback := codec.CBORCodec{}

	rawString, err := (codec.CBORCodec{}).Encode(testName)
	if err != nil {
		t.Fatalf("encode cbor string: %v", err)
	}

	if rawString[0] >= 0x80 {
		t.Fatalf("fixture must start below 0x80, got 0x%02x", rawString[0])
	}

	c, inner := codec.UnwrapDecode(rawString, fallback)
	if c.Encoding() != codec.EncodingCBOR {
		t.Fatalf("expected fallback codec for raw CBOR scalar")
	}

	if string(inner) != string(rawString) {
		t.Fatalf("data should be unchanged")
	}

	rawInt, err := (codec.CBORCodec{}).Encode(int64(42))
	if err != nil {
		t.Fatalf("encode cbor int: %v", err)
	}

	c, inner = codec.UnwrapDecode(rawInt, fallback)
	if c.Encoding() != codec.EncodingCBOR {
		t.Fatalf("expected fallback codec for raw CBOR int")
	}

	if string(inner) != string(rawInt) {
		t.Fatalf("data should be unchanged")
	}
}

func TestWrapEncode_EnvelopeStructure(t *testing.T) {
	t.Parallel()

	wrapped, err := codec.WrapEncode(map[string]string{"k": "v"}, codec.JSONCodec{})
	if err != nil {
		t.Fatalf("codec.WrapEncode: %v", err)
	}

	// The codec.Envelope should always be JSON, even for CBOR inner data.
	var env codec.Envelope
	if err := (codec.JSONCodec{}).Decode(wrapped, &env); err != nil {
		t.Fatalf("codec.Envelope should be JSON-decodable: %v", err)
	}

	if env.Magic != codec.EnvelopeMagic {
		t.Fatalf("magic = %q, want %q", env.Magic, codec.EnvelopeMagic)
	}

	if env.Encoding != codec.EncodingJSON {
		t.Fatalf("encoding = %s, want %s", env.Encoding, codec.EncodingJSON)
	}

	if len(env.Data) == 0 {
		t.Fatal("inner data should not be empty")
	}
}

func TestWrapEncode_Deterministic(t *testing.T) {
	t.Parallel()

	type pair struct {
		A string `json:"a"`
		B string `json:"b"`
	}

	val := pair{A: "1", B: "2"}

	a, _ := codec.WrapEncode(val, codec.JSONCodec{})
	b, _ := codec.WrapEncode(val, codec.JSONCodec{})

	if string(a) != string(b) {
		t.Fatalf("codec.Envelope encoding is not deterministic:\n%q\n%q", a, b)
	}
}
