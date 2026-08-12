package codec_test

import (
	"errors"
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestGetBuffer_ReturnsResetBuffer(t *testing.T) {
	t.Parallel()

	buf := codec.GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer returned nil")
	}

	if buf.Len() != 0 {
		t.Fatalf("buffer not reset: len=%d", buf.Len())
	}

	buf.WriteString("residual data")
	codec.PutBuffer(buf)

	buf2 := codec.GetBuffer()
	if buf2.Len() != 0 {
		t.Fatalf("pooled buffer not reset: len=%d", buf2.Len())
	}

	codec.PutBuffer(buf2)
}

func TestPutBuffer_NilIsSafe(t *testing.T) {
	t.Parallel()

	codec.PutBuffer(nil) // must not panic
}

func TestGetBuffer_RoundTrip(t *testing.T) {
	t.Parallel()

	buf := codec.GetBuffer()
	buf.Grow(4096)
	buf.WriteString("x")
	codec.PutBuffer(buf)

	// sync.Pool may or may not return the same buffer, but it must be usable.
	buf2 := codec.GetBuffer()
	if buf2 == nil {
		t.Fatal("GetBuffer returned nil after round-trip")
	}

	if buf2.Len() != 0 {
		t.Fatalf("pooled buffer not reset: len=%d", buf2.Len())
	}

	buf2.WriteString("test")

	if buf2.Len() != 4 {
		t.Fatalf("buffer not writable after pool round-trip: len=%d", buf2.Len())
	}

	codec.PutBuffer(buf2)
}

func TestEncodePooled_CBORRoundTrip(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name  string
		Value int
	}

	c := codec.CBORCodec{}
	want := payload{Name: "test", Value: 42}

	var encoded []byte

	err := codec.EncodePooled(c, want, func(data []byte) error {
		encoded = make([]byte, len(data)) //nolint:makezero // intentional: allocate exact size then fill via copy
		copy(encoded, data)

		return nil
	})
	if err != nil {
		t.Fatalf("EncodePooled: %v", err)
	}

	var got payload
	if err := c.Decode(encoded, &got); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got.Name != want.Name || got.Value != want.Value {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestEncodePooled_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name  string
		Value int
	}

	c := codec.JSONCodec{}
	want := payload{Name: "test", Value: 42}

	var encoded []byte

	err := codec.EncodePooled(c, want, func(data []byte) error {
		encoded = make([]byte, len(data)) //nolint:makezero // intentional: allocate exact size then fill via copy
		copy(encoded, data)

		return nil
	})
	if err != nil {
		t.Fatalf("EncodePooled: %v", err)
	}

	var got payload
	if err := c.Decode(encoded, &got); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if got.Name != want.Name || got.Value != want.Value {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestEncodePooled_CallbackError(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}
	errSentinel := errors.New("callback error")

	err := codec.EncodePooled(c, map[string]string{"k": "v"}, func([]byte) error {
		return errSentinel
	})
	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected callback error, got: %v", err)
	}
}

func TestEncodePooled_CallbackMustNotRetainBuffer(t *testing.T) {
	t.Parallel()

	c := codec.CBORCodec{}

	var stale []byte

	_ = codec.EncodePooled(c, map[string]string{testAlpha: "1"}, func(data []byte) error {
		stale = data // intentionally retain to prove it becomes stale

		return nil
	})

	// Right after EncodePooled returns, stale still points to the first encode's
	// bytes because the buffer has not yet been reused. Reading it now is safe
	// only because no other pool user has run yet.
	if len(stale) == 0 {
		t.Error("first callback received empty data")
	}

	// A second encode may reuse the same pooled buffer, overwriting stale's
	// backing array. The callback must copy data if it needs to keep it.
	_ = codec.EncodePooled(c, map[string]string{testBeta: "2"}, func(data []byte) error {
		if len(data) == 0 {
			t.Error("second encode produced empty output")
		}

		return nil
	})

	// stale is intentionally not read again — it may point to reused pool memory.
}
