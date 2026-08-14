// Package codec provides payload encoding and decoding for event sourcing.
//
// The Codec interface abstracts serialization so that stores, snapshots, and
// event construction can work with any encoding format. Four implementations
// are provided:
//
//   - CBORCodec: recommended default — canonical CBOR (deterministic, signing-safe)
//   - JSONCodec: standard encoding/json — universal interop, human-readable
//   - CBORCompactCodec: stricter CBOR with unknown-field rejection (schema drift guard)
//   - RawCodec: passthrough for pre-encoded []byte payloads
//
// CBORCodec and CBORCompactCodec are NOT interchangeable: they sort map keys
// differently (RFC 7049 length-first vs RFC 8949 bytewise-lexical) and the
// compact codec rejects unknown fields, so data written by one does not
// round-trip through the other. Choose one codec per store and keep it
// documented.
//
// # Choosing a Codec
//
// Both JSON and CBOR are fully supported across the library. CBOR is recommended
// for internal serialization: smaller payloads (see BenchmarkTagTradeoffs for
// measured size reductions across payload shapes), faster to encode and decode,
// and deterministic (same input always produces the same output bytes — safe for
// cryptographic signing). JSON is the right choice for external interop, HTTP
// APIs, debugging, and any case where human-readability matters.
//
// Payloads that will be signed or content-addressed require byte-deterministic
// encoding. The DeterministicCodec interface makes that a compile-time
// guarantee: CBORCodec and CBORCompactCodec always satisfy it, JSONCodec
// satisfies it only in the GOEXPERIMENT=jsonv2 build, and RawCodec never does.
// Signing modules should accept DeterministicCodec rather than Codec so that a
// non-deterministic codec choice fails to build instead of silently producing
// signatures that break on re-encoding.
//
// # Usage
//
//	codec := codec.JSONCodec{}
//	data, err := codec.Encode(MyPayload{Name: "Alice"})
//	var decoded MyPayload
//	err = codec.Decode(data, &decoded)
//
// # Integration
//
// The Codec is used by event.New (auto-marshal payloads), event.DecodePayload[T]
// (typed decode), and snapshot stores (serialize stream state).
//
// The encryption module provides a composable codec wrapper (encryption.NewCodec)
// that wraps any Codec with transparent encrypt-on-encode / decrypt-on-decode.
// It reports its own encoding ("encrypted") and is used with event.WithCodec
// to create events with encrypted payloads.
//
// # Cross-Codec Transcoding
//
// TranscodeToJSON converts a payload from its stamped encoding into JSON bytes
// — the schema-free bridge for consumers that store events in CBOR but must
// serve JSON to browsers or REST clients. Non-CBOR payloads pass through
// unchanged; CBOR is decoded into a generic Go value and re-encoded as JSON.
//
//	out, err := codec.TranscodeToJSON(payload, evt.Encoding())
//
// It is schema-free: CBOR maps become JSON objects, but structs encoded with
// the cbor:",toarray" tag become JSON arrays (field names are lost). For
// schema-aware JSON, decode with the concrete type first, then json.Marshal.
// The transport/http package provides CBORToJSONTransform, a ready-made
// adapter that wraps TranscodeToJSON for use with WithPayloadTransform.
//
// TranscodeToJSON is a one-way operation: CBOR → JSON. There is intentionally
// no TranscodeToCBOR because the reverse direction cannot reconstruct CBOR
// type information (map key order, toarray positions, keyasint integers) from
// JSON without the original Go type. To produce CBOR, decode into the concrete
// type and encode with CBORCodec directly.
//
// # CBOR Compact Encoding (toarray)
//
// For maximum payload size reduction (~30-40%), add the cbor:",toarray" struct
// tag to event payload types. This encodes structs as positional CBOR arrays
// instead of keyed maps, eliminating field-name string overhead entirely.
//
//	type UserCreated struct {
//	    _     struct{} `cbor:",toarray"`
//	    Name  string
//	    Email string
//	    Time  int64
//	}
//
// Without toarray (map encoding):  {"Name":"Alice","Email":"a@b.com","Time":1700000000}
// With toarray (array encoding):   ["Alice","a@b.com",1700000000]
//
// The toarray tag works with both CBORCodec and CBORCompactCodec. It is a
// per-type decision — mix array-encoded and map-encoded types freely. Once a
// struct uses toarray, field ORDER is part of the wire format and cannot be
// reordered without breaking existing data. Add new fields only at the end.
//
// For additional compactness, use CBORCompactCodec (CoreDetEncOptions +
// ExtraReturnErrors for schema drift detection). See CBORCompactCodec docs.
//
// # Zero-Allocation Encoding
//
// Codecs that implement the BufferEncoder interface can write directly into a
// caller-provided bytes.Buffer, avoiding the allocation returned by Encode.
// This is useful in hot paths where buffer reuse eliminates GC pressure.
//
//	buf := &bytes.Buffer{}
//	if be, ok := codec.(codec.BufferEncoder); ok {
//	    _ = be.EncodeToBuffer(payload, buf)
//	}
//
// For even simpler pool-backed encoding, use [EncodePooled], which manages a
// sync.Pool-backed buffer automatically via a callback:
//
//	err := codec.EncodePooled(cborCodec, event, func(data []byte) error {
//	    _, werr := store.Write(data)
//	    return werr
//	})
//
// # Streaming
//
// For large event batches, streaming encoders/decoders avoid materializing the
// full byte slice in memory. Both CBOR and JSON are supported:
//
//	enc := codec.NewCBOREncoder(w)  // or codec.NewJSONEncoder(w)
//	_ = enc.Encode(event1)
//	_ = enc.Encode(event2)
//
//	dec := codec.NewCBORDecoder(r)  // or codec.NewJSONDecoder(r)
//	for {
//	    var evt Event
//	    if err := dec.Decode(&evt); err != nil { break }
//	    events = append(events, evt)
//	}
//
// JSON streaming uses newline-delimited JSON (NDJSON / JSON Lines): each Encode
// call writes one JSON value followed by a newline, and Decode reads one value
// at a time, skipping whitespace between values.
//
// # Observability
//
// ObserveCodec wraps any Codec with opt-in telemetry: per-operation call counts,
// byte totals, error counts, and last errors (CodecMetrics), plus a push-style
// MetricsHook for Prometheus/OpenTelemetry-style exporters. The wrapper is
// goroutine-safe and preserves the BufferEncoder fast path when present.
//
//	obs := codec.ObserveCodec(codec.CBORCodec{},
//	    codec.WithMetricsHook(func(op codec.Operation, enc codec.Encoding, n int, err error) {
//	        // emit to your metrics backend
//	    }),
//	)
//
// # Format Detection
//
// AutoDetect infers the encoding (json/cbor/raw) of unknown payload bytes from
// the leading byte, with trial-decode fallback for ambiguous starts. It is a
// best-effort heuristic for diagnostics and tooling, NOT a security boundary.
// AutoDetectDebug additionally returns a stable DetectionReason to branch on
// and a human-readable Detail string for logs (unstable wording — never parse).
//
// # Performance
//
// fxamacker/cbor caches type metadata in a process-wide sync.Map keyed by
// reflect.Type. The first encode/decode of each type pays a one-time reflection
// cost (~100µs); subsequent operations reuse the cached metadata (~300ns for a
// 7-field struct). Code generation is NOT needed for typical use cases — the
// cache handles it. Use [EncodePooled] or [BufferEncoder] to eliminate
// allocation overhead on hot paths.
//
// The toarray and keyasint struct tags reduce payload size by eliminating
// field-name strings. Measured tradeoffs across payload shapes:
//
//	Payload      map (bytes)   toarray (bytes)   keyasint (bytes)
//	small (3)    56            43 (-23%)          46 (-18%)
//	medium (7)   218           156 (-28%)         168 (-23%)
//	large (12)   270           159 (-41%)         171 (-37%)
//
// toarray produces the smallest payloads but encodes field order as positional
// (reordering fields breaks stored data). keyasint is slightly larger but
// preserves field identity via stable integer keys (safer for schema evolution).
// Run BenchmarkTagTradeoffs_Encode / BenchmarkTagTradeoffs_Decode for details.
//
// # CBOR Compact Struct Tags
//
// fxamacker/cbor supports two struct tags for further payload optimization:
//
// keyasint — encode struct fields as integer keys instead of string keys.
// This eliminates field-name strings entirely, ideal for high-frequency
// events with many fields.
//
//	type OrderPlaced struct {
//	    _      struct{} `cbor:",keyasint"`
//	    UserID uint64   `cbor:"1,keyasint"`
//	    ItemID uint64   `cbor:"2,keyasint"`
//	    Qty    int      `cbor:"3,keyasint"`
//	}
//
// omitzero — omit fields that are zero-valued. Reduces payload size for
// events where many fields are optional.
//
//	type UserUpdated struct {
//	    Name  string `cbor:"name"`
//	    Email string `cbor:"email,omitempty"`
//	    Bio   string `cbor:"bio,omitzero"`
//	}
//
// Both tags work with CBORCodec and CBORCompactCodec. Once adopted, the
// integer key mapping is part of the wire format — changing key numbers
// breaks existing data.
package codec
