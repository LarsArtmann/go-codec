package codec_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/larsartmann/go-codec"
)

func ExampleJSONCodec() {
	c := codec.JSONCodec{}

	type User struct {
		Name string `json:"name"`
	}

	data, err := c.Encode(User{Name: testName})
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	var user User

	err = c.Decode(data, &user)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(user.Name)

	// Output:
	// Alice
}

func ExampleCBORCodec() {
	c := codec.CBORCodec{}

	type User struct {
		Name string
	}

	data, err := c.Encode(User{Name: testName})
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	var user User

	err = c.Decode(data, &user)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(user.Name)

	// Output:
	// Alice
}

func ExampleRawCodec() {
	c := codec.RawCodec{}

	raw := []byte(`{"name":"Bob"}`)

	data, err := c.Encode(raw)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	var decoded []byte

	err = c.Decode(data, &decoded)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(string(decoded))

	// Output:
	// {"name":"Bob"}
}

// ExampleCBORCompactCodec demonstrates the strict CBOR codec that rejects
// unknown fields on decode — a schema drift detection mechanism.
func ExampleCBORCompactCodec() {
	c := codec.CBORCompactCodec{}

	type UserCreated struct {
		Name  string
		Email string
	}

	data, _ := c.Encode(UserCreated{Name: testName, Email: testEmail})

	var result UserCreated

	_ = c.Decode(data, &result)

	fmt.Println(result.Name, result.Email)

	// Output:
	// Alice alice@example.com
}

// ExampleCBORCodec_toarray demonstrates the toarray struct tag that encodes
// structs as positional CBOR arrays instead of keyed maps, reducing payload
// size by 30-40% by eliminating field-name string overhead.
func ExampleCBORCodec_toarray() {
	c := codec.CBORCodec{}

	// The _ field with `cbor:",toarray"` tag enables positional encoding.
	// Field ORDER becomes part of the wire format — add new fields only at end.
	type PaymentProcessed struct {
		_           struct{} `cbor:",toarray"`
		PaymentID   string
		AmountCents int64
		Currency    string
		OccurredAt  int64
	}

	payload := PaymentProcessed{
		PaymentID:   "pay_abc123",
		AmountCents: 4999,
		Currency:    "USD",
		OccurredAt:  1700000000,
	}

	mapCodec := codec.JSONCodec{}
	mapData, _ := mapCodec.Encode(payload)

	arrayData, _ := c.Encode(payload)

	fmt.Printf("JSON: %d bytes, CBOR+toarray: %d bytes (%.0f%% smaller)\n",
		len(mapData), len(arrayData),
		float64(len(mapData)-len(arrayData))/float64(len(mapData))*100)

	var decoded PaymentProcessed

	_ = c.Decode(arrayData, &decoded)
	fmt.Println(decoded.PaymentID, decoded.AmountCents)

	// Output:
	// JSON: 86 bytes, CBOR+toarray: 24 bytes (72% smaller)
	// pay_abc123 4999
}

// ExampleBufferEncoder demonstrates zero-allocation encoding by writing
// directly into a caller-provided buffer. Useful in hot paths where buffer
// reuse eliminates GC pressure.
func ExampleBufferEncoder() {
	type Metric struct {
		Name  string
		Value float64
	}

	c := codec.CBORCodec{}
	buf := &bytes.Buffer{}

	// Reuse the same buffer across multiple encode calls
	for _, m := range []Metric{
		{Name: "cpu", Value: 0.42},
		{Name: "mem", Value: 0.87},
	} {
		buf.Reset()

		if be, ok := any(c).(codec.BufferEncoder); ok {
			_ = be.EncodeToBuffer(m, buf)
		}

		fmt.Printf("%d bytes ", buf.Len())
	}

	fmt.Println("done")

	// Output:
	// 25 bytes 25 bytes done
}

// ExampleNewCBOREncoder demonstrates streaming CBOR encoding for large
// event batches without materializing the full byte slice in memory.
func ExampleNewCBOREncoder() {
	type Event struct {
		Type string
		Data string
	}

	var buf bytes.Buffer

	enc := codec.NewCBOREncoder(&buf)
	_ = enc.Encode(Event{Type: testEventType, Data: testUserName})
	_ = enc.Encode(Event{Type: testEventType, Data: "bob"})

	// Decode the stream
	dec := codec.NewCBORDecoder(&buf)

	var events []Event

	for {
		var evt Event
		if err := dec.Decode(&evt); err != nil {
			break
		}

		events = append(events, evt)
	}

	fmt.Printf("%d events decoded from stream\n", len(events))

	// Output:
	// 2 events decoded from stream
}

// ExampleNewJSONEncoder demonstrates streaming JSON encoding for large
// event batches using newline-delimited JSON (NDJSON). Each Encode call
// writes one JSON value followed by a newline, enabling incremental
// consumption on the reader side.
func ExampleNewJSONEncoder() {
	type Event struct {
		Type string
		Data string
	}

	var buf bytes.Buffer

	enc := codec.NewJSONEncoder(&buf)
	_ = enc.Encode(Event{Type: testEventType, Data: testUserName})
	_ = enc.Encode(Event{Type: testEventType, Data: "bob"})

	// Decode the NDJSON stream
	dec := codec.NewJSONDecoder(&buf)

	var events []Event

	for {
		var evt Event
		if err := dec.Decode(&evt); err != nil {
			break
		}

		events = append(events, evt)
	}

	fmt.Printf("%d events decoded from JSON stream\n", len(events))

	// Output:
	// 2 events decoded from JSON stream
}

// ExampleObserveCodec demonstrates wrapping a codec with telemetry. The
// ObservableCodec decorator records per-operation metrics and invokes a
// MetricsHook after every encode/decode — useful for Prometheus or
// OpenTelemetry-style push telemetry without polling.
func ExampleObserveCodec() {
	obs := codec.ObserveCodec(codec.CBORCodec{},
		codec.WithMetricsHook(func(op codec.Operation, enc codec.Encoding, bytes int, err error) {
			fmt.Printf("op=%d enc=%s err=%v\n", op, enc, err)
		}))

	data, _ := obs.Encode(map[string]string{testFieldName: testName})

	var decoded map[string]string

	_ = obs.Decode(data, &decoded)

	m := obs.Metrics().Snapshot()
	fmt.Printf("encode calls: %d, decode calls: %d\n", m.EncodeCalls, m.DecodeCalls)

	// Output:
	// op=0 enc=cbor err=<nil>
	// op=1 enc=cbor err=<nil>
	// encode calls: 1, decode calls: 1
}

// ExampleAutoDetectDebug demonstrates explainable format detection: beyond the
// inferred encoding, it returns a stable DetectionReason to branch on and a
// human-readable Detail for logs. Detail is NOT a stable contract — never
// parse it.
func ExampleAutoDetectDebug() {
	for _, data := range [][]byte{
		[]byte(`{"name":"Alice"}`),
		{0xa1, 0x64, 'n', 'a', 'm', 'e', 0x65, 'A', 'l', 'i', 'c', 'e'}, // CBOR map
		{0x1f}, // unrecognized
	} {
		result := codec.AutoDetectDebug(data)

		switch result.Reason {
		case codec.DetectionReasonJSONStructure, codec.DetectionReasonJSONTrialDecode:
			fmt.Println("json")
		case codec.DetectionReasonCBORMajorType, codec.DetectionReasonCBORTrialDecode:
			fmt.Println("cbor")
		case codec.DetectionReasonEmpty,
			codec.DetectionReasonOversized,
			codec.DetectionReasonUnknown:
			fmt.Printf("other (%s): %s\n", result.Reason, result.Detail)
		}
	}

	// Output:
	// json
	// cbor
	// other (unknown): first byte 0x1f does not match any known format
}

// ExampleDiagnose converts CBOR bytes to human-readable diagnostic notation.
// Useful for debugging corrupt events or inspecting raw CBOR payloads.
func ExampleDiagnose() {
	c := codec.CBORCodec{}

	type User struct {
		Name  string
		Email string
	}

	data, _ := c.Encode(User{Name: testName, Email: testEmail})

	diag, _ := codec.Diagnose(data)

	// Diagnostic notation is a map-like representation
	fmt.Println(strings.Contains(diag, testName))

	// Output:
	// true
}

// ExampleCBOREncMode demonstrates using the exported canonical encoding mode
// directly. Storage backends should use this instead of creating their own
// CBOR mode, ensuring all modules share one deterministic encoding.
func ExampleCBOREncMode() {
	type Snapshot struct {
		State string
		N     int
	}

	snap := Snapshot{State: "active", N: 42}

	// Same canonical EncMode used by CBORCodec internally
	data, err := codec.CBOREncMode().Marshal(snap)
	if err != nil {
		log.Fatal(err)
	}

	var decoded Snapshot

	_ = codec.CBORDecMode().Unmarshal(data, &decoded)

	fmt.Printf("%s/%d (%d bytes)\n", decoded.State, decoded.N, len(data))

	// Output:
	// active/42 (18 bytes)
}

// ExampleCBORCodec_keyasint demonstrates the keyasint struct tag that encodes
// field names as compact integer keys instead of strings. This is the COSE/CWT
// (CBOR Web Token) pattern used by JWT-like claims: each field gets a small
// integer key, shrinking payloads without losing the self-describing map
// structure (unlike toarray, field order doesn't matter).
func ExampleCBORCodec_keyasint() {
	c := codec.CBORCodec{}

	// Integer keys follow the CWT claim registry (RFC 8392).
	type Claims struct {
		Iss string `cbor:"1,keyasint"`
		Sub string `cbor:"2,keyasint"`
		Aud string `cbor:"3,keyasint"`
		Exp int64  `cbor:"4,keyasint"`
	}

	// String-keyed equivalent for size comparison
	type ClaimsString struct {
		Iss string
		Sub string
		Aud string
		Exp int64
	}

	claims := Claims{Iss: "s6BhdRkqt3", Sub: "24400320", Aud: "s6BhdRkqt3", Exp: 1735689600}

	data, _ := c.Encode(claims)
	stringData, _ := c.Encode(ClaimsString(claims))

	fmt.Printf("keyasint: %d bytes, string keys: %d bytes\n", len(data), len(stringData))

	var decoded Claims

	_ = c.Decode(data, &decoded)
	fmt.Println(decoded.Iss)

	// Output:
	// keyasint: 41 bytes, string keys: 53 bytes
	// s6BhdRkqt3
}

// ExampleEncodePooled demonstrates zero-allocation encoding using a pooled
// buffer. The callback receives the encoded bytes and must copy them if it
// needs to retain them after the callback returns — the buffer is returned
// to the pool immediately.
func ExampleEncodePooled() {
	type Event struct {
		Type string
		Data string
	}

	c := codec.CBORCodec{}
	evt := Event{Type: "created", Data: "hello"}

	var encoded []byte

	err := codec.EncodePooled(c, evt, func(data []byte) error {
		encoded = make([]byte, len(data)) // copy: data is invalid after callback
		copy(encoded, data)
		return nil
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	var decoded Event
	_ = c.Decode(encoded, &decoded)
	fmt.Println(decoded.Type, decoded.Data)

	// Output:
	// created hello
}

// ExampleSize compares the serialized byte sizes of a value under JSON and
// CBOR, helping decide whether a format switch is worthwhile before committing.
func ExampleSize() {
	type UserCreated struct {
		Name  string
		Email string
	}

	s := codec.Size(UserCreated{Name: testName, Email: testEmail})

	if s.CBOR < s.JSON {
		fmt.Printf("CBOR saves %d bytes (%.0f%% smaller)\n",
			s.JSON-s.CBOR,
			float64(s.JSON-s.CBOR)/float64(s.JSON)*100)
	} else {
		fmt.Println("CBOR is not smaller for this payload")
	}

	// Output:
	// CBOR saves 8 bytes (18% smaller)
}
