package codec

import (
	"bytes"
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
