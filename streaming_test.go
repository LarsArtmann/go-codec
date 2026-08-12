package codec_test

import (
	"bytes"
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestStreaming_CBOR(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	enc := codec.NewCBOREncoder(&buf)
	if enc == nil {
		t.Fatal("codec.NewCBOREncoder returned nil")
	}

	type payload struct {
		Name string
		Age  int
	}

	if err := enc.Encode(payload{Name: testName, Age: 30}); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	dec := codec.NewCBORDecoder(&buf)
	if dec == nil {
		t.Fatal("codec.NewCBORDecoder returned nil")
	}

	var got payload
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got.Name != testName || got.Age != 30 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestStreaming_CBOREncoderMultiple(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	enc := codec.NewCBOREncoder(&buf)

	type item struct{ ID int }

	_ = enc.Encode(item{ID: 1})
	_ = enc.Encode(item{ID: 2})

	dec := codec.NewCBORDecoder(&buf)

	var first, second item

	_ = dec.Decode(&first)
	_ = dec.Decode(&second)

	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("expected 1,2 got %d,%d", first.ID, second.ID)
	}
}

func TestCanonicalEncMode(t *testing.T) {
	t.Parallel()

	mode := codec.CanonicalEncMode()
	if mode == nil {
		t.Fatal("codec.CanonicalEncMode should not be nil")
	}
}

func TestCanonicalDecMode(t *testing.T) {
	t.Parallel()

	mode := codec.CanonicalDecMode()
	if mode == nil {
		t.Fatal("codec.CanonicalDecMode should not be nil")
	}
}
