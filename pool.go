package codec

import (
	"bytes"
	"fmt"
	"sync"
)

// bufferPool reuse *bytes.Buffer across EncodeToBuffer calls to eliminate
// allocation pressure in hot paths. Callers must call PutBuffer when done.
//
//nolint:gochecknoglobals // intentional pool, standard sync.Pool pattern
var bufferPool = sync.Pool{
	New: func() any { return &bytes.Buffer{} },
}

// GetBuffer returns a zeroed *bytes.Buffer from the pool. The caller owns the
// buffer until PutBuffer is called.
func GetBuffer() *bytes.Buffer {
	b, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok {
		b = &bytes.Buffer{}
	}

	b.Reset()

	return b
}

// PutBuffer returns a buffer to the pool for reuse. Do not use the buffer
// after calling PutBuffer.
func PutBuffer(b *bytes.Buffer) {
	if b == nil {
		return
	}

	bufferPool.Put(b)
}

// EncodePooled encodes v using a pooled *bytes.Buffer and passes the encoded
// bytes to fn. The buffer is returned to the pool after fn returns, so fn must
// not retain references to the byte slice after returning — if it needs to keep
// the bytes, it must copy them.
//
// This eliminates the per-call []byte allocation of [Codec.Encode] in hot paths
// where the caller processes the encoded bytes immediately (e.g., writing to a
// store, sending over the wire). For callers that need to keep the bytes, use
// [Codec.Encode] instead.
//
//	cbor := codec.CBORCodec{}
//	err := codec.EncodePooled(cbor, event, func(data []byte) error {
//	    _, werr := store.Write(data)
//	    return werr
//	})
func EncodePooled(enc BufferEncoder, v any, fn func([]byte) error) error {
	buf := GetBuffer()
	defer PutBuffer(buf)

	if err := enc.EncodeToBuffer(v, buf); err != nil {
		return fmt.Errorf("codec: pooled encode failed: %w", err)
	}

	return fn(buf.Bytes())
}
