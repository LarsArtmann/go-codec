package codec_test

import (
	"errors"
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestForEncoding_JSON(t *testing.T) {
	t.Parallel()

	c, err := codec.ForEncoding(codec.EncodingJSON)
	if err != nil {
		t.Fatalf("codec.ForEncoding(JSON): %v", err)
	}

	if c.Encoding() != codec.EncodingJSON {
		t.Errorf("codec.Encoding() = %q, want %q", c.Encoding(), codec.EncodingJSON)
	}
}

func TestForEncoding_CBOR(t *testing.T) {
	t.Parallel()

	c, err := codec.ForEncoding(codec.EncodingCBOR)
	if err != nil {
		t.Fatalf("codec.ForEncoding(CBOR): %v", err)
	}

	if c.Encoding() != codec.EncodingCBOR {
		t.Errorf("codec.Encoding() = %q, want %q", c.Encoding(), codec.EncodingCBOR)
	}
}

func TestForEncoding_UnknownReturnsError(t *testing.T) {
	t.Parallel()

	cases := []codec.Encoding{
		"encrypted",
		"msgpack",
		"",
	}

	for _, enc := range cases {
		c, err := codec.ForEncoding(enc)
		if err == nil {
			t.Errorf("codec.ForEncoding(%q) = %v, want error", enc, c)

			continue
		}

		if !errors.Is(err, codec.ErrUnknownEncoding) {
			t.Errorf("codec.ForEncoding(%q) err = %v, want codec.ErrUnknownEncoding", enc, err)
		}
	}
}

func TestForEncoding_Raw(t *testing.T) {
	t.Parallel()

	c, err := codec.ForEncoding(codec.EncodingRaw)
	if err != nil {
		t.Fatalf("codec.ForEncoding(Raw): %v", err)
	}

	if c.Encoding() != codec.EncodingRaw {
		t.Errorf("codec.Encoding() = %q, want %q", c.Encoding(), codec.EncodingRaw)
	}
}

func TestForEncoding_RoundTripsWithAutoDetect(t *testing.T) {
	t.Parallel()

	type payload struct{ Name string }

	for _, c := range []codec.Codec{codec.JSONCodec{}, codec.CBORCodec{}} {
		data, err := c.Encode(payload{Name: testName})
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}

		detected := codec.AutoDetect(data)

		resolved, err := codec.ForEncoding(detected)
		if err != nil {
			t.Fatalf("codec.ForEncoding(%s): %v", detected, err)
		}

		if resolved.Encoding() != c.Encoding() {
			t.Errorf(
				"codec.ForEncoding(codec.AutoDetect(data)) = %s, want %s",
				resolved.Encoding(),
				c.Encoding(),
			)
		}

		var got payload
		if err := resolved.Decode(data, &got); err != nil {
			t.Fatalf("Decode: %v", err)
		}

		if got.Name != testName {
			t.Errorf("Name = %q, want %q", got.Name, testName)
		}
	}
}
