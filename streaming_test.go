package codec_test

import (
	"bytes"
	"io"
	"strings"
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

func TestStreaming_JSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)
	if enc == nil {
		t.Fatal("codec.NewJSONEncoder returned nil")
	}

	type payload struct {
		Name string
		Age  int
	}

	if err := enc.Encode(payload{Name: testName, Age: 30}); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	dec := codec.NewJSONDecoder(&buf)
	if dec == nil {
		t.Fatal("codec.NewJSONDecoder returned nil")
	}

	var got payload
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got.Name != testName || got.Age != 30 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestStreaming_JSONEncoderMultiple(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)

	type item struct{ ID int }

	_ = enc.Encode(item{ID: 1})
	_ = enc.Encode(item{ID: 2})

	dec := codec.NewJSONDecoder(&buf)

	var first, second item

	_ = dec.Decode(&first)
	_ = dec.Decode(&second)

	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("expected 1,2 got %d,%d", first.ID, second.ID)
	}
}

func TestStreaming_JSONNewlineDelimited(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)

	type event struct {
		Type string
		Data string
	}

	events := []event{
		{Type: "created", Data: testGreeting},
		{Type: "updated", Data: testUserName},
		{Type: "deleted", Data: testBob},
	}

	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("Encode: %v", err)
		}
	}

	dec := codec.NewJSONDecoder(&buf)

	var got []event

	for {
		var e event
		if err := dec.Decode(&e); err != nil {
			break
		}

		got = append(got, e)
	}

	if len(got) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(got))
	}

	for i, e := range got {
		if e.Type != events[i].Type || e.Data != events[i].Data {
			t.Fatalf("event %d mismatch: got %+v, want %+v", i, e, events[i])
		}
	}
}

// TestStreaming_JSONNonBufferReader verifies that the JSON streaming decoder
// works correctly with a non-buffer reader (strings.Reader). bytes.Buffer
// masks over-read bugs because its Read method is more forgiving than a
// real stream — strings.Reader enforces strict sequential reading, exposing
// any decoder that over-reads past a JSON value boundary.
func TestStreaming_JSONNonBufferReader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)

	type item struct {
		ID   int
		Name string
	}

	items := []item{
		{ID: 1, Name: testName},
		{ID: 2, Name: testBob},
		{ID: 3, Name: testUserName},
	}

	for _, it := range items {
		if err := enc.Encode(it); err != nil {
			t.Fatalf("Encode: %v", err)
		}
	}

	// Use strings.Reader instead of bytes.Buffer — it does not support
	// UnreadByte or peeking, so any over-read by the decoder corrupts
	// subsequent Decode calls.
	dec := codec.NewJSONDecoder(strings.NewReader(buf.String()))

	var got []item

	for {
		var it item
		if err := dec.Decode(&it); err != nil {
			break
		}

		got = append(got, it)
	}

	if len(got) != len(items) {
		t.Fatalf("expected %d items, got %d", len(items), len(got))
	}

	for i, it := range got {
		if it.ID != items[i].ID || it.Name != items[i].Name {
			t.Fatalf("item %d mismatch: got %+v, want %+v", i, it, items[i])
		}
	}
}

// TestStreaming_JSONByteAtATimeReader uses a reader that yields one byte per
// Read call, stress-testing the decoder's buffering logic. This catches
// boundary bugs that strings.Reader might not expose.
func TestStreaming_JSONByteAtATimeReader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)

	type payload struct {
		Value string
	}

	for _, v := range []string{testGreeting, testName, testBob} {
		if err := enc.Encode(payload{Value: v}); err != nil {
			t.Fatalf("Encode: %v", err)
		}
	}

	dec := codec.NewJSONDecoder(&byteAtATimeReader{data: buf.Bytes()})

	var got []payload

	for {
		var p payload
		if err := dec.Decode(&p); err != nil {
			break
		}

		got = append(got, p)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 values, got %d", len(got))
	}

	expected := []string{testGreeting, testName, testBob}

	for i, p := range got {
		if p.Value != expected[i] {
			t.Fatalf("value %d mismatch: got %q, want %q", i, p.Value, expected[i])
		}
	}
}

// byteAtATimeReader yields exactly one byte per Read call, forcing the
// decoder to handle arbitrarily small reads.
type byteAtATimeReader struct {
	data []byte
	pos  int
}

func (r *byteAtATimeReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	p[0] = r.data[r.pos]
	r.pos++

	return 1, nil
}
