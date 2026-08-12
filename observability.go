package codec

import (
	"bytes"
	"fmt"
	"sync"
)

// Operation identifies the kind of observed codec operation.
type Operation int

const (
	// OpEncode is an encoding operation.
	OpEncode Operation = iota

	// OpDecode is a decoding operation.
	OpDecode
)

// MetricsHook receives a notification after each observed operation.
// bytesProcessed is the length of the encoded payload for encodes or the
// input length for decodes. err is the result of the operation (nil on success).
type MetricsHook func(op Operation, enc Encoding, bytesProcessed int, err error)

// CodecMetrics records per-instance codec operation counters and last results.
// It is safe for concurrent use by multiple goroutines.
type CodecMetrics struct {
	mu              sync.RWMutex
	encodeCalls     int64
	encodeBytes     int64
	encodeErrors    int64
	lastEncodeError error
	decodeCalls     int64
	decodeBytes     int64
	decodeErrors    int64
	lastDecodeError error
}

func (m *CodecMetrics) recordEncode(bytesProcessed int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.encodeCalls++
	m.encodeBytes += int64(bytesProcessed)

	if err != nil {
		m.encodeErrors++
		m.lastEncodeError = err
	}
}

func (m *CodecMetrics) recordDecode(bytesProcessed int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.decodeCalls++
	m.decodeBytes += int64(bytesProcessed)

	if err != nil {
		m.decodeErrors++
		m.lastDecodeError = err
	}
}

// Snapshot returns a point-in-time copy of the metrics.
func (m *CodecMetrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsSnapshot{
		EncodeCalls:     m.encodeCalls,
		EncodeBytes:     m.encodeBytes,
		EncodeErrors:    m.encodeErrors,
		LastEncodeError: m.lastEncodeError,
		DecodeCalls:     m.decodeCalls,
		DecodeBytes:     m.decodeBytes,
		DecodeErrors:    m.decodeErrors,
		LastDecodeError: m.lastDecodeError,
	}
}

// Reset clears all recorded metrics.
func (m *CodecMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.encodeCalls = 0
	m.encodeBytes = 0
	m.encodeErrors = 0
	m.lastEncodeError = nil
	m.decodeCalls = 0
	m.decodeBytes = 0
	m.decodeErrors = 0
	m.lastDecodeError = nil
}

// MetricsSnapshot is a point-in-time copy of CodecMetrics.
type MetricsSnapshot struct {
	EncodeCalls     int64
	EncodeBytes     int64
	EncodeErrors    int64
	LastEncodeError error
	DecodeCalls     int64
	DecodeBytes     int64
	DecodeErrors    int64
	LastDecodeError error
}

// ObserveOption configures an ObservableCodec.
type ObserveOption func(*ObservableCodec)

// WithMetricsHook registers a hook that is called after every Encode/Decode.
func WithMetricsHook(hook MetricsHook) ObserveOption {
	return func(obs *ObservableCodec) {
		obs.hook = hook
	}
}

// WithMetrics stores metrics in the supplied CodecMetrics instead of creating
// a private one. This lets multiple ObservableCodecs share a metrics sink or
// lets callers inspect metrics directly.
func WithMetrics(metrics *CodecMetrics) ObserveOption {
	return func(obs *ObservableCodec) {
		obs.metrics = metrics
	}
}

// ObservableCodec wraps a Codec and records metrics for every operation.
// It implements Codec and, when the wrapped codec implements BufferEncoder,
// BufferEncoder as well.
type ObservableCodec struct {
	codec   Codec
	metrics *CodecMetrics
	hook    MetricsHook
}

// ObserveCodec returns an observable wrapper around c.
func ObserveCodec(c Codec, opts ...ObserveOption) *ObservableCodec {
	obs := &ObservableCodec{
		codec:   c,
		metrics: &CodecMetrics{},
	}

	for _, opt := range opts {
		opt(obs)
	}

	return obs
}

// Encoding returns the wrapped codec's encoding.
func (obs *ObservableCodec) Encoding() Encoding {
	return obs.codec.Encoding()
}

// Metrics returns the live metrics sink for this wrapper.
func (obs *ObservableCodec) Metrics() *CodecMetrics {
	return obs.metrics
}

// Encode marshals v and records metrics.
func (obs *ObservableCodec) Encode(v any) ([]byte, error) {
	data, err := obs.codec.Encode(v)

	obs.metrics.recordEncode(len(data), err)

	if obs.hook != nil {
		obs.hook(OpEncode, obs.codec.Encoding(), len(data), err)
	}

	//nolint:wrapcheck // thin wrapper around the wrapped codec's Encode
	return data, err
}

// Decode unmarshals data and records metrics.
func (obs *ObservableCodec) Decode(data []byte, v any) error {
	err := obs.codec.Decode(data, v)

	obs.metrics.recordDecode(len(data), err)

	if obs.hook != nil {
		obs.hook(OpDecode, obs.codec.Encoding(), len(data), err)
	}

	//nolint:wrapcheck // thin wrapper around the wrapped codec's Decode
	return err
}

// EncodeToBuffer writes v into buf using the wrapped codec if it implements
// BufferEncoder; otherwise it encodes and writes the bytes.
func (obs *ObservableCodec) EncodeToBuffer(v any, buf *bytes.Buffer) error {
	if be, ok := obs.codec.(BufferEncoder); ok {
		start := buf.Len()
		err := be.EncodeToBuffer(v, buf)
		processed := buf.Len() - start

		obs.metrics.recordEncode(processed, err)

		if obs.hook != nil {
			obs.hook(OpEncode, obs.codec.Encoding(), processed, err)
		}

		//nolint:wrapcheck // thin wrapper around the wrapped codec's EncodeToBuffer
		return err
	}

	data, err := obs.codec.Encode(v)
	if err != nil {
		obs.metrics.recordEncode(0, err)

		if obs.hook != nil {
			obs.hook(OpEncode, obs.codec.Encoding(), 0, err)
		}

		//nolint:wrapcheck // thin wrapper around the wrapped codec's Encode
		return err
	}

	if _, writeErr := buf.Write(data); writeErr != nil {
		obs.metrics.recordEncode(len(data), writeErr)

		if obs.hook != nil {
			obs.hook(OpEncode, obs.codec.Encoding(), len(data), writeErr)
		}

		return fmt.Errorf("codec: observable write to buffer: %w", writeErr)
	}

	obs.metrics.recordEncode(len(data), nil)

	if obs.hook != nil {
		obs.hook(OpEncode, obs.codec.Encoding(), len(data), nil)
	}

	return nil
}
