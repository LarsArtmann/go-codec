package codec

import (
	"bytes"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Encoding identifies the serialization format used for a payload.
type Encoding string

const (
	EncodingJSON Encoding = "json"
	EncodingCBOR Encoding = "cbor"
	EncodingRaw  Encoding = "raw"
)

// ErrUnknownEncoding is returned by [ForEncoding] when no built-in codec
// matches the requested encoding.
var ErrUnknownEncoding = errorfamily.NewRejection(
	"codec.unknown_encoding",
	"codec: unknown encoding (no built-in codec)",
)

// ForEncoding returns the built-in [Codec] for the given [Encoding].
// It resolves [EncodingJSON] → [JSONCodec], [EncodingCBOR] → [CBORCodec],
// and [EncodingRaw] → [RawCodec].
//
// For unknown encodings (including custom values like "encrypted"), it returns
// [ErrUnknownEncoding]. Callers that need custom encoding support should build
// their own dispatch table.
//
// ForEncoding is the codec-level counterpart to [AutoDetect]: AutoDetect
// infers the encoding from raw bytes, ForEncoding resolves a known encoding
// stamp to its codec. Together they enable mixed-stream decoding — see
// [event.DecodePayloadAuto].
func ForEncoding(enc Encoding) (Codec, error) {
	switch enc {
	case EncodingJSON:
		return JSONCodec{}, nil
	case EncodingCBOR:
		return CBORCodec{}, nil
	case EncodingRaw:
		return RawCodec{}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownEncoding, enc)
	}
}

// Codec serializes and deserializes values with a declared encoding.
type Codec interface {
	Encoding() Encoding
	Encode(v any) ([]byte, error)
	Decode(data []byte, v any) error
}

// BufferEncoder is an optional interface that Codecs can implement for
// zero-allocation encoding. Instead of allocating a new []byte on every
// Encode call, EncodeToBuffer writes directly into a caller-provided buffer.
// Callers can reuse the buffer across calls to eliminate GC pressure in
// hot paths (batch event publishing, bulk snapshot saving).
type BufferEncoder interface {
	EncodeToBuffer(v any, buf *bytes.Buffer) error
}

// DeterministicCodec is a marker interface implemented by codecs whose Encode
// produces byte-identical output for equal inputs on every call. Codecs that
// satisfy this interface are safe for cryptographic signing: the signing
// module can assert this at compile time, turning silent signature corruption
// (e.g. using JSONCodec in v1 mode, whose map key order is non-deterministic)
// into a build error.
//
// The unexported method prevents external packages from implementing the
// interface — only the built-in codecs in this package can be signing-safe.
//
// CBORCodec and CBORCompactCodec always implement DeterministicCodec.
// JSONCodec implements it only in the v2 build (encoding/json/v2 with
// Deterministic mode); the v1 default does not because map key ordering is
// non-deterministic. RawCodec never implements it because it performs no
// encoding at all.
type DeterministicCodec interface {
	Codec
	signingSafe()
}
