package codec_test

import (
	"bytes"
	"testing"
	"testing/iotest"

	"github.com/larsartmann/go-codec"
)

// FuzzStreamingJSON_NDJSONRoundtrip locks two NDJSON streaming invariants:
//
//  1. Framing — every JSONEncoder.Encode writes exactly one value plus a
//     newline, and exactly that many values can be decoded back, ending in a
//     clean EOF (no extra values, no data loss).
//  2. Buffering independence — the decode count is identical through a
//     byte-at-a-time reader. This is the over-read regression class that bit
//     the v2 decoder (json.UnmarshalRead over-reads from the io.Reader and
//     silently broke sequential Decode calls; see json_compat_v2.go).
func FuzzStreamingJSON_NDJSONRoundtrip(f *testing.F) {
	f.Add([]byte(`{"a":1}`))
	f.Add([]byte(`[1,2,3]`))
	f.Add([]byte(`"text"`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"nested":{"deep":[true,false]}}`))
	f.Add([]byte(`123.45`))

	f.Fuzz(func(t *testing.T, value []byte) {
		var buf bytes.Buffer

		enc := codec.NewJSONEncoder(&buf)

		if err := enc.Encode(codec.RawJSONValue(value)); err != nil {
			t.Skip() // not a valid JSON value — writer correctly rejects it
		}

		if err := enc.Encode(codec.RawJSONValue(value)); err != nil {
			t.Fatalf("second Encode: %v", err)
		}

		if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
			t.Fatalf("NDJSON stream must end with a newline, got %q", buf.Bytes())
		}

		dec := codec.NewJSONDecoder(bytes.NewReader(buf.Bytes()))

		for i := range 2 {
			var got codec.RawJSONValue
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("Decode #%d: %v", i, err)
			}
		}

		var extra codec.RawJSONValue
		if err := dec.Decode(&extra); err == nil {
			t.Fatal("expected EOF on third Decode, got a value")
		}

		slow := codec.NewJSONDecoder(iotest.OneByteReader(bytes.NewReader(buf.Bytes())))

		for i := range 2 {
			var got codec.RawJSONValue
			if err := slow.Decode(&got); err != nil {
				t.Fatalf("byte-at-a-time Decode #%d: %v", i, err)
			}
		}

		var extraSlow codec.RawJSONValue
		if err := slow.Decode(&extraSlow); err == nil {
			t.Fatal("expected EOF on third byte-at-a-time Decode, got a value")
		}
	})
}
