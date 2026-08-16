package codec_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestDecodeEnvelopeOrLegacy_EnvelopeStampedData(t *testing.T) {
	t.Parallel()

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	wrapped, err := codec.WrapEncode(user{Name: testName, Email: testEmail}, codec.CBORCodec{})
	if err != nil {
		t.Fatalf("WrapEncode: %v", err)
	}

	// The envelope wins regardless of what is configured.
	got, err := codec.DecodeEnvelopeOrLegacy[user](wrapped, codec.JSONCodec{})
	if err != nil {
		t.Fatalf("DecodeEnvelopeOrLegacy: %v", err)
	}

	if got.Name != testName || got.Email != testEmail {
		t.Fatalf("decoded = %+v", got)
	}
}

func TestDecodeEnvelopeOrLegacy_RawJSONUnderCBORConfigured(t *testing.T) {
	t.Parallel()

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	raw := []byte(`{"name":"` + testName + `","email":"` + testEmail + `"}`)

	got, err := codec.DecodeEnvelopeOrLegacy[user](raw, codec.CBORCodec{})
	if err != nil {
		t.Fatalf("legacy raw JSON under CBOR config: %v", err)
	}

	if got.Name != testName || got.Email != testEmail {
		t.Fatalf("decoded = %+v", got)
	}
}

func TestDecodeEnvelopeOrLegacy_RawCBORUnderJSONConfigured(t *testing.T) {
	t.Parallel()

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	raw, err := codec.CBORCodec{}.Encode(user{Name: testName, Email: testEmail})
	if err != nil {
		t.Fatalf("encode raw CBOR: %v", err)
	}

	got, err := codec.DecodeEnvelopeOrLegacy[user](raw, codec.JSONCodec{})
	if err != nil {
		t.Fatalf("legacy raw CBOR under JSON config: %v", err)
	}

	if got.Name != testName || got.Email != testEmail {
		t.Fatalf("decoded = %+v", got)
	}
}

func TestDecodeEnvelopeOrLegacy_GarbageStillErrors(t *testing.T) {
	t.Parallel()

	type user struct {
		Name string `json:"name"`
	}

	if _, err := codec.DecodeEnvelopeOrLegacy[user](
		[]byte{0xc1, 0xff, 0xfe, 0x00}, codec.CBORCodec{},
	); err == nil {
		t.Fatal("expected decode error for garbage data")
	}
}

type prefixCodec struct{}

func (prefixCodec) Encoding() codec.Encoding { return codec.Encoding("prefix") }

func (prefixCodec) Encode(v any) ([]byte, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("prefixCodec: unsupported type %T", v)
	}

	return []byte("PREFIX:" + s), nil
}

func (prefixCodec) Decode(data []byte, v any) error {
	p, ok := v.(*string)
	if !ok {
		return fmt.Errorf("prefixCodec: unsupported target %T", v)
	}

	if !strings.HasPrefix(string(data), "PREFIX:") {
		return errors.New("prefixCodec: missing PREFIX: marker")
	}

	*p = strings.TrimPrefix(string(data), "PREFIX:")

	return nil
}

func TestDecodeEnvelopeOrLegacy_CustomCodecNoCrossRetry(t *testing.T) {
	t.Parallel()

	// A custom codec whose data only decodes with itself: the helper must
	// attempt it (and only it) rather than silently switching to JSON/CBOR.
	got, err := codec.DecodeEnvelopeOrLegacy[string]([]byte("PREFIX:hello"), prefixCodec{})
	if err != nil {
		t.Fatalf("custom codec decode: %v", err)
	}

	if got != "hello" {
		t.Fatalf("decoded = %q", got)
	}

	// Data the custom codec rejects errors — no standard cross-retry happens.
	if _, err := codec.DecodeEnvelopeOrLegacy[string](
		[]byte("garbage"), prefixCodec{},
	); err == nil {
		t.Fatal("expected decode error for custom-codec rejection")
	}
}
