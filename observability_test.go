package codec_test

import (
	"bytes"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/larsartmann/go-codec"
)

func TestObservableCodec_EncodeDecode_RecordsMetrics(t *testing.T) {
	t.Parallel()

	obs := codec.ObserveCodec(codec.JSONCodec{})
	payload := map[string]string{testField: testName}

	data, err := obs.Encode(payload)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var decoded map[string]string
	if err := obs.Decode(data, &decoded); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	m := obs.Metrics().Snapshot()

	if m.EncodeCalls != 1 {
		t.Errorf("EncodeCalls = %d, want 1", m.EncodeCalls)
	}

	if m.EncodeBytes != int64(len(data)) {
		t.Errorf("EncodeBytes = %d, want %d", m.EncodeBytes, len(data))
	}

	if m.DecodeCalls != 1 {
		t.Errorf("DecodeCalls = %d, want 1", m.DecodeCalls)
	}

	if m.DecodeBytes != int64(len(data)) {
		t.Errorf("DecodeBytes = %d, want %d", m.DecodeBytes, len(data))
	}

	if m.EncodeErrors != 0 {
		t.Errorf("EncodeErrors = %d, want 0", m.EncodeErrors)
	}

	if m.DecodeErrors != 0 {
		t.Errorf("DecodeErrors = %d, want 0", m.DecodeErrors)
	}

	if m.LastEncodeError != nil {
		t.Errorf("LastEncodeError = %v, want nil", m.LastEncodeError)
	}

	if m.LastDecodeError != nil {
		t.Errorf("LastDecodeError = %v, want nil", m.LastDecodeError)
	}
}

func TestObservableCodec_ErrorsRecorded(t *testing.T) {
	t.Parallel()

	errSentinel := errors.New("encode error")
	obs := codec.ObserveCodec(failingCodec{encodeErr: errSentinel})

	if _, err := obs.Encode(map[string]string{}); !errors.Is(err, errSentinel) {
		t.Fatalf("expected encode error, got %v", err)
	}

	m := obs.Metrics().Snapshot()

	if m.EncodeCalls != 1 {
		t.Errorf("EncodeCalls = %d, want 1", m.EncodeCalls)
	}

	if m.EncodeErrors != 1 {
		t.Errorf("EncodeErrors = %d, want 1", m.EncodeErrors)
	}

	if !errors.Is(m.LastEncodeError, errSentinel) {
		t.Errorf("LastEncodeError = %v, want %v", m.LastEncodeError, errSentinel)
	}
}

func TestObservableCodec_HookCalled(t *testing.T) {
	t.Parallel()

	var calls []hookCall

	obs := codec.ObserveCodec(codec.JSONCodec{}, codec.WithMetricsHook(func(
		op codec.Operation,
		enc codec.Encoding,
		bytesProcessed int,
		err error,
	) {
		calls = append(calls, hookCall{op, enc, bytesProcessed, err})
	}))

	payload := map[string]string{testField: testName}

	data, err := obs.Encode(payload)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var decoded map[string]string
	if err := obs.Decode(data, &decoded); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("hook calls = %d, want 2", len(calls))
	}

	enc := calls[0]

	if enc.op != codec.OpEncode {
		t.Errorf("first call op = %v, want OpEncode", enc.op)
	}

	if enc.enc != codec.EncodingJSON {
		t.Errorf("first call encoding = %q, want %q", enc.enc, codec.EncodingJSON)
	}

	if enc.bytesProcessed != len(data) {
		t.Errorf("first call bytesProcessed = %d, want %d", enc.bytesProcessed, len(data))
	}

	if enc.err != nil {
		t.Errorf("first call err = %v, want nil", enc.err)
	}

	dec := calls[1]

	if dec.op != codec.OpDecode {
		t.Errorf("second call op = %v, want OpDecode", dec.op)
	}

	if dec.enc != codec.EncodingJSON {
		t.Errorf("second call encoding = %q, want %q", dec.enc, codec.EncodingJSON)
	}

	if dec.bytesProcessed != len(data) {
		t.Errorf("second call bytesProcessed = %d, want %d", dec.bytesProcessed, len(data))
	}

	if dec.err != nil {
		t.Errorf("second call err = %v, want nil", dec.err)
	}
}

func TestObservableCodec_SharedMetrics(t *testing.T) {
	t.Parallel()

	shared := &codec.CodecMetrics{}
	obs1 := codec.ObserveCodec(codec.JSONCodec{}, codec.WithMetrics(shared))
	obs2 := codec.ObserveCodec(codec.JSONCodec{}, codec.WithMetrics(shared))

	payload := map[string]string{testField: testName}

	if _, err := obs1.Encode(payload); err != nil {
		t.Fatalf("obs1 Encode: %v", err)
	}

	if _, err := obs2.Encode(payload); err != nil {
		t.Fatalf("obs2 Encode: %v", err)
	}

	m := shared.Snapshot()

	if m.EncodeCalls != 2 {
		t.Errorf("EncodeCalls = %d, want 2", m.EncodeCalls)
	}

	if m.EncodeBytes <= 0 {
		t.Errorf("EncodeBytes = %d, want > 0", m.EncodeBytes)
	}
}

func TestObservableCodec_BufferEncoder(t *testing.T) {
	t.Parallel()

	obs := codec.ObserveCodec(codec.CBORCodec{})

	// Compile-time check that ObservableCodec implements BufferEncoder when the
	// wrapped codec does.
	var _ codec.BufferEncoder = obs

	payload := map[string]string{testField: testName}
	buf := &bytes.Buffer{}

	if err := obs.EncodeToBuffer(payload, buf); err != nil {
		t.Fatalf("EncodeToBuffer: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("EncodeToBuffer wrote no bytes")
	}

	m := obs.Metrics().Snapshot()

	if m.EncodeCalls != 1 {
		t.Errorf("EncodeCalls = %d, want 1", m.EncodeCalls)
	}

	if m.EncodeBytes != int64(buf.Len()) {
		t.Errorf("EncodeBytes = %d, want %d", m.EncodeBytes, buf.Len())
	}
}

func TestObservableCodec_NonBufferEncoderFallback(t *testing.T) {
	t.Parallel()

	obs := codec.ObserveCodec(codec.RawCodec{})
	payload := []byte{0x01, 0x02, 0x03}
	buf := &bytes.Buffer{}

	if err := obs.EncodeToBuffer(payload, buf); err != nil {
		t.Fatalf("EncodeToBuffer: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), payload) {
		t.Errorf("buffer = %v, want %v", buf.Bytes(), payload)
	}

	m := obs.Metrics().Snapshot()

	if m.EncodeCalls != 1 {
		t.Errorf("EncodeCalls = %d, want 1", m.EncodeCalls)
	}

	if m.EncodeBytes != int64(len(payload)) {
		t.Errorf("EncodeBytes = %d, want %d", m.EncodeBytes, len(payload))
	}
}

func TestObservableCodec_Reset(t *testing.T) {
	t.Parallel()

	obs := codec.ObserveCodec(codec.JSONCodec{})

	if _, err := obs.Encode(map[string]string{testField: testName}); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	obs.Metrics().Reset()

	m := obs.Metrics().Snapshot()

	if m.EncodeCalls != 0 {
		t.Errorf("EncodeCalls after Reset = %d, want 0", m.EncodeCalls)
	}

	if m.EncodeBytes != 0 {
		t.Errorf("EncodeBytes after Reset = %d, want 0", m.EncodeBytes)
	}
}

func TestObservableCodec_ConcurrentStress(t *testing.T) {
	t.Parallel()

	const goroutines = 32

	const iterations = 500

	shared := &codec.CodecMetrics{}

	var hookCalls int64

	hook := stressHook(t, &hookCalls)
	obs := codec.ObserveCodec(codec.JSONCodec{}, codec.WithMetrics(shared), codec.WithMetricsHook(hook))

	payload := map[string]string{testField: testName}

	var wg sync.WaitGroup

	wg.Add(goroutines * iterations)

	for range goroutines * iterations {
		go func() {
			defer wg.Done()

			data, err := obs.Encode(payload)
			if err != nil {
				t.Errorf("Encode: %v", err)

				return
			}

			var decoded map[string]string

			if err := obs.Decode(data, &decoded); err != nil {
				t.Errorf("Decode: %v", err)

				return
			}
		}()
	}

	wg.Wait()

	assertStressMetrics(t, shared.Snapshot(), atomic.LoadInt64(&hookCalls), goroutines*iterations)
}

// stressHook returns a MetricsHook that validates every call and counts
// invocations into calls atomically.
func stressHook(t *testing.T, calls *int64) codec.MetricsHook {
	t.Helper()

	return func(op codec.Operation, enc codec.Encoding, bytesProcessed int, err error) {
		if err != nil {
			t.Errorf("hook saw unexpected error: %v", err)

			return
		}

		if enc != codec.EncodingJSON {
			t.Errorf("hook encoding = %q, want %q", enc, codec.EncodingJSON)

			return
		}

		if bytesProcessed <= 0 {
			t.Errorf("hook bytesProcessed = %d, want > 0", bytesProcessed)

			return
		}

		if op != codec.OpEncode && op != codec.OpDecode {
			t.Errorf("hook op = %v, want encode or decode", op)

			return
		}

		atomic.AddInt64(calls, 1)
	}
}

func assertStressMetrics(t *testing.T, m codec.MetricsSnapshot, hookCalls int64, ops int) {
	t.Helper()

	if m.EncodeCalls != int64(ops) {
		t.Errorf("EncodeCalls = %d, want %d", m.EncodeCalls, ops)
	}

	if m.DecodeCalls != int64(ops) {
		t.Errorf("DecodeCalls = %d, want %d", m.DecodeCalls, ops)
	}

	if m.EncodeErrors != 0 || m.DecodeErrors != 0 {
		t.Errorf("EncodeErrors = %d, DecodeErrors = %d, want 0", m.EncodeErrors, m.DecodeErrors)
	}

	if hookCalls != int64(2*ops) {
		t.Errorf("hookCalls = %d, want %d", hookCalls, 2*ops)
	}

	if m.EncodeBytes != m.DecodeBytes {
		t.Errorf("EncodeBytes %d != DecodeBytes %d", m.EncodeBytes, m.DecodeBytes)
	}
}

// TestObservableCodec_HookPanicPropagates locks the documented panic policy:
// hook panics are NOT recovered; they propagate to the caller. Metrics are
// recorded before the hook runs, so counters stay consistent despite the panic.
func TestObservableCodec_HookPanicPropagates(t *testing.T) {
	t.Parallel()

	obs := codec.ObserveCodec(codec.JSONCodec{},
		codec.WithMetricsHook(func(codec.Operation, codec.Encoding, int, error) {
			panic("hook exploded")
		}))

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Encode did not propagate the hook panic")

			return
		}

		if r != "hook exploded" {
			t.Fatalf("recovered %v, want %q", r, "hook exploded")
		}

		m := obs.Metrics().Snapshot()

		if m.EncodeCalls != 1 {
			t.Errorf("EncodeCalls after hook panic = %d, want 1 (metrics recorded before hook)", m.EncodeCalls)
		}
	}()

	_, _ = obs.Encode(map[string]string{testField: testName})

	t.Fatal("unreachable: Encode should have panicked")
}

type hookCall struct {
	op             codec.Operation
	enc            codec.Encoding
	bytesProcessed int
	err            error
}

type failingCodec struct {
	encodeErr error
	decodeErr error
}

func (c failingCodec) Encoding() codec.Encoding {
	return codec.EncodingRaw
}

func (c failingCodec) Encode(any) ([]byte, error) {
	return nil, c.encodeErr
}

func (c failingCodec) Decode([]byte, any) error {
	return c.decodeErr
}
