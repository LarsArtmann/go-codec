package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// FuzzObservableCodec_HookSafety locks the observability contract under
// arbitrary payloads and operation counts:
//
//   - the MetricsHook fires exactly once per codec operation,
//   - EncodeCalls/DecodeCalls counters equal the number of operations regardless
//     of whether the operation succeeded,
//   - nothing panics or deadlocks for any input.
//
// Hook panics intentionally propagate (documented policy, locked by
// TestObservableCodec_HookPanicPropagates); this target fuzzes the non-panic
// guarantee.
func FuzzObservableCodec_HookSafety(f *testing.F) {
	f.Add([]byte(`{"a":1}`), 3)
	f.Add([]byte{0xff, 0x00}, 1)
	f.Add([]byte(`[1,2,3]`), 7)
	f.Add([]byte(`not json`), 2)

	f.Fuzz(func(t *testing.T, data []byte, ops int) {
		if ops < 1 || ops > 16 {
			t.Skip() // keep the run bounded; repetition adds no coverage
		}

		hookCalls := 0

		obs := codec.ObserveCodec(codec.JSONCodec{}, codec.WithMetricsHook(
			func(_ codec.Operation, _ codec.Encoding, _ int, _ error) {
				hookCalls++
			},
		))

		for range ops {
			_, _ = obs.Encode(codec.RawJSONValue(data))
			_ = obs.Decode(data, &map[string]any{})
		}

		snap := obs.Metrics().Snapshot()

		if hookCalls != 2*ops {
			t.Fatalf("hook called %d times, want exactly %d (once per operation)", hookCalls, 2*ops)
		}

		if snap.EncodeCalls != int64(ops) {
			t.Fatalf("EncodeCalls = %d, want %d", snap.EncodeCalls, ops)
		}

		if snap.DecodeCalls != int64(ops) {
			t.Fatalf("DecodeCalls = %d, want %d", snap.DecodeCalls, ops)
		}

		if snap.EncodeErrors > snap.EncodeCalls || snap.DecodeErrors > snap.DecodeCalls {
			t.Fatalf("error counts exceed call counts: encode %d/%d, decode %d/%d",
				snap.EncodeErrors, snap.EncodeCalls, snap.DecodeErrors, snap.DecodeCalls)
		}
	})
}
