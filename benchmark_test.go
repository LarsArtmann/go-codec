package codec_test

import (
	"bytes"
	"testing"

	"github.com/larsartmann/go-codec"
)

const (
	benchNameJSON = "JSON"
	benchNameCBOR = "CBOR"
)

func BenchmarkJSONCodec_Encode(b *testing.B) {
	b.ReportAllocs()

	c := codec.JSONCodec{}
	payload := map[string]string{testFieldName: testName, testFieldEmail: testEmail}

	b.ResetTimer()

	for b.Loop() {
		_, err := c.Encode(payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONCodec_Decode(b *testing.B) {
	b.ReportAllocs()

	c := codec.JSONCodec{}
	data, _ := c.Encode(map[string]string{testFieldName: testName, testFieldEmail: testEmail})

	b.ResetTimer()

	for b.Loop() {
		var result map[string]string
		if err := c.Decode(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCBORCodec_Encode(b *testing.B) {
	b.ReportAllocs()

	c := codec.CBORCodec{}
	payload := map[string]string{testFieldName: testName, testFieldEmail: testEmail}

	b.ResetTimer()

	for b.Loop() {
		_, err := c.Encode(payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCBORCodec_Decode(b *testing.B) {
	b.ReportAllocs()

	c := codec.CBORCodec{}
	data, _ := c.Encode(map[string]string{testFieldName: testName, testFieldEmail: testEmail})

	b.ResetTimer()

	for b.Loop() {
		var result map[string]string
		if err := c.Decode(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCodecComparison_Encode(b *testing.B) {
	b.ReportAllocs()

	jsonCodec := codec.JSONCodec{}
	cborCodec := codec.CBORCodec{}
	payload := map[string]string{testFieldName: testName, testFieldEmail: testEmail}

	b.Run(benchNameJSON, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_, err := jsonCodec.Encode(payload)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run(benchNameCBOR, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			_, err := cborCodec.Encode(payload)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCodecComparison_Decode(b *testing.B) {
	b.ReportAllocs()

	jsonCodec := codec.JSONCodec{}
	cborCodec := codec.CBORCodec{}
	payload := map[string]string{testFieldName: testName, testFieldEmail: testEmail}

	jsonData, _ := jsonCodec.Encode(payload)
	cborData, _ := cborCodec.Encode(payload)

	b.Run(benchNameJSON, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result map[string]string
			if err := jsonCodec.Decode(jsonData, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run(benchNameCBOR, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result map[string]string
			if err := cborCodec.Decode(cborData, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkRawCodec_Encode(b *testing.B) {
	b.ReportAllocs()

	c := codec.RawCodec{}
	data := []byte("raw payload bytes")

	b.ResetTimer()

	for b.Loop() {
		_, err := c.Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRawCodec_Decode(b *testing.B) {
	b.ReportAllocs()

	c := codec.RawCodec{}
	data := []byte("raw payload bytes")

	b.ResetTimer()

	for b.Loop() {
		var result []byte
		if err := c.Decode(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCBORCompact_vs_Canon_Size(b *testing.B) {
	type eventPayload struct {
		Name    string
		Email   string
		Version int
		Active  bool
	}

	payload := eventPayload{Name: testName, Email: testEmail, Version: 42, Active: true}

	canonical := codec.CBORCodec{}
	compact := codec.CBORCompactCodec{}

	canonicalData, _ := canonical.Encode(payload)
	compactData, _ := compact.Encode(payload)

	b.Logf(
		"CBOR (canonical): %d bytes, CBOR (compact): %d bytes, savings: %.1f%%",
		len(canonicalData), len(compactData),
		float64(len(canonicalData)-len(compactData))/float64(len(canonicalData))*100,
	)

	codecs := []struct {
		name string
		c    codec.Codec
	}{
		{"Canonical", canonical},
		{"Compact", compact},
	}

	for _, tc := range codecs {
		b.Run(tc.name+"/Encode", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_, err := tc.c.Encode(payload)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// realisticOrder simulates a real-world event payload with mixed field types.
type realisticOrder struct {
	OrderID    string
	CustomerID string
	Items      []orderItem
	TotalCents int64
	Currency   string
	Status     string
	CreatedAt  int64
}

type orderItem struct {
	SKU       string
	Quantity  int
	UnitPrice int64
}

// realisticOrderArray uses the toarray tag for positional CBOR encoding.
type realisticOrderArray struct {
	_          struct{} `cbor:",toarray"`
	OrderID    string
	CustomerID string
	Items      []orderItem
	TotalCents int64
	Currency   string
	Status     string
	CreatedAt  int64
}

func sampleOrder() realisticOrder {
	return realisticOrder{
		OrderID:    testOrderID,
		CustomerID: testCustID,
		Items: []orderItem{
			{SKU: "WIDGET-001", Quantity: 2, UnitPrice: 1999},
			{SKU: "GADGET-042", Quantity: 1, UnitPrice: 4999},
		},
		TotalCents: 8997,
		Currency:   "USD",
		Status:     "pending",
		CreatedAt:  1700000000,
	}
}

func BenchmarkRealisticPayload_Encode(b *testing.B) {
	order := sampleOrder()
	orderArr := realisticOrderArray{
		OrderID:    order.OrderID,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		TotalCents: order.TotalCents,
		Currency:   order.Currency,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
	}

	jsonCodec := codec.JSONCodec{}
	cborCodec := codec.CBORCodec{}
	compactCodec := codec.CBORCompactCodec{}

	jsonData, _ := jsonCodec.Encode(order)
	cborData, _ := cborCodec.Encode(order)
	compactData, _ := compactCodec.Encode(orderArr)

	b.Logf("Realistic order payload sizes:")
	b.Logf("  JSON:              %d bytes", len(jsonData))
	b.Logf("  CBOR canonical:    %d bytes (%.1f%% of JSON)",
		len(cborData), float64(len(cborData))/float64(len(jsonData))*100)
	b.Logf("  CBOR compact+toarray: %d bytes (%.1f%% of JSON)",
		len(compactData), float64(len(compactData))/float64(len(jsonData))*100)

	codecs := []struct {
		name string
		c    codec.Codec
		v    any
	}{
		{benchNameJSON, jsonCodec, order},
		{benchNameCBOR, cborCodec, order},
		{"CBOR_compact_toarray", compactCodec, orderArr},
	}

	for _, tc := range codecs {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				_, err := tc.c.Encode(tc.v)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRealisticPayload_Decode(b *testing.B) {
	order := sampleOrder()
	orderArr := realisticOrderArray{
		OrderID:    order.OrderID,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		TotalCents: order.TotalCents,
		Currency:   order.Currency,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
	}

	jsonCodec := codec.JSONCodec{}
	cborCodec := codec.CBORCodec{}
	compactCodec := codec.CBORCompactCodec{}

	jsonData, _ := jsonCodec.Encode(order)
	cborData, _ := cborCodec.Encode(order)
	compactData, _ := compactCodec.Encode(orderArr)

	b.Run(benchNameJSON, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result realisticOrder
			if err := jsonCodec.Decode(jsonData, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run(benchNameCBOR, func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result realisticOrder
			if err := cborCodec.Decode(cborData, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CBOR_compact_toarray", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result realisticOrderArray
			if err := compactCodec.Decode(compactData, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkBufferEncoder(b *testing.B) {
	order := sampleOrder()

	codecs := []struct {
		name string
		c    codec.BufferEncoder
	}{
		{benchNameJSON, codec.JSONCodec{}},
		{benchNameCBOR, codec.CBORCodec{}},
		{"CBOR_compact", codec.CBORCompactCodec{}},
	}

	for _, tc := range codecs {
		b.Run(tc.name, func(b *testing.B) {
			buf := &bytes.Buffer{}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				buf.Reset()

				if err := tc.c.EncodeToBuffer(order, buf); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEncodePooled(b *testing.B) {
	order := sampleOrder()

	codecs := []struct {
		name string
		c    codec.BufferEncoder
	}{
		{benchNameJSON, codec.JSONCodec{}},
		{benchNameCBOR, codec.CBORCodec{}},
		{"CBOR_compact", codec.CBORCompactCodec{}},
	}

	for _, tc := range codecs {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				err := codec.EncodePooled(tc.c, order, func([]byte) error {
					return nil
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// --- toarray / keyasint tradeoff benchmarks ---

// smallEvent is a minimal 3-field event for measuring overhead on tiny payloads.
type smallEvent struct {
	ID   string
	Type string
	Data string
}

type smallEventArray struct {
	_    struct{} `cbor:",toarray"`
	ID   string
	Type string
	Data string
}

type smallEventKeyInt struct {
	_    struct{} `cbor:",keyasint"`
	ID   string   `cbor:"1,keyasint"`
	Type string   `cbor:"2,keyasint"`
	Data string   `cbor:"3,keyasint"`
}

func sampleSmallEvent() smallEvent {
	return smallEvent{ID: testOrderID, Type: "created", Data: testUserName}
}

// realisticOrderKeyInt uses keyasint for integer-keyed CBOR encoding.
type realisticOrderKeyInt struct {
	_          struct{}    `cbor:",keyasint"`
	OrderID    string      `cbor:"1,keyasint"`
	CustomerID string      `cbor:"2,keyasint"`
	Items      []orderItem `cbor:"3,keyasint"`
	TotalCents int64       `cbor:"4,keyasint"`
	Currency   string      `cbor:"5,keyasint"`
	Status     string      `cbor:"6,keyasint"`
	CreatedAt  int64       `cbor:"7,keyasint"`
}

// largeEvent is a wide 12-field event for measuring overhead on larger payloads.
type largeEvent struct {
	EventID       string
	AggregateID   string
	EventType     string
	Version       int64
	Timestamp     int64
	UserID        string
	TraceID       string
	SpanID        string
	Source        string
	TenantID      string
	CorrelationID string
	OccurredAt    int64
}

type largeEventArray struct {
	_             struct{} `cbor:",toarray"`
	EventID       string
	AggregateID   string
	EventType     string
	Version       int64
	Timestamp     int64
	UserID        string
	TraceID       string
	SpanID        string
	Source        string
	TenantID      string
	CorrelationID string
	OccurredAt    int64
}

type largeEventKeyInt struct {
	_             struct{} `cbor:",keyasint"`
	EventID       string   `cbor:"1,keyasint"`
	AggregateID   string   `cbor:"2,keyasint"`
	EventType     string   `cbor:"3,keyasint"`
	Version       int64    `cbor:"4,keyasint"`
	Timestamp     int64    `cbor:"5,keyasint"`
	UserID        string   `cbor:"6,keyasint"`
	TraceID       string   `cbor:"7,keyasint"`
	SpanID        string   `cbor:"8,keyasint"`
	Source        string   `cbor:"9,keyasint"`
	TenantID      string   `cbor:"10,keyasint"`
	CorrelationID string   `cbor:"11,keyasint"`
	OccurredAt    int64    `cbor:"12,keyasint"`
}

func sampleLargeEvent() largeEvent {
	return largeEvent{
		EventID:       testOrderID,
		AggregateID:   testCustID,
		EventType:     "order.placed",
		Version:       42,
		Timestamp:     1700000000,
		UserID:        "user_abc123",
		TraceID:       "trace_xyz789",
		SpanID:        "span_def456",
		Source:        "checkout-service",
		TenantID:      "tenant_001",
		CorrelationID: "corr_999888",
		OccurredAt:    1700000001,
	}
}

// BenchmarkTagTradeoffs_Encode measures encode speed and allocation across
// default (map), toarray, and keyasint for small, medium, and large payloads.
func BenchmarkTagTradeoffs_Encode(b *testing.B) {
	cborCodec := codec.CBORCodec{}

	small := sampleSmallEvent()
	smallArr := smallEventArray{ID: small.ID, Type: small.Type, Data: small.Data}
	smallKI := smallEventKeyInt{ID: small.ID, Type: small.Type, Data: small.Data}

	order := sampleOrder()
	orderArr := realisticOrderArray{
		OrderID: order.OrderID, CustomerID: order.CustomerID, Items: order.Items,
		TotalCents: order.TotalCents, Currency: order.Currency, Status: order.Status, CreatedAt: order.CreatedAt,
	}
	orderKI := realisticOrderKeyInt{
		OrderID: order.OrderID, CustomerID: order.CustomerID, Items: order.Items,
		TotalCents: order.TotalCents, Currency: order.Currency, Status: order.Status, CreatedAt: order.CreatedAt,
	}

	large := sampleLargeEvent()
	largeArr := largeEventArray{
		EventID: large.EventID, AggregateID: large.AggregateID, EventType: large.EventType,
		Version: large.Version, Timestamp: large.Timestamp, UserID: large.UserID,
		TraceID: large.TraceID, SpanID: large.SpanID, Source: large.Source,
		TenantID: large.TenantID, CorrelationID: large.CorrelationID, OccurredAt: large.OccurredAt,
	}
	largeKI := largeEventKeyInt{
		EventID: large.EventID, AggregateID: large.AggregateID, EventType: large.EventType,
		Version: large.Version, Timestamp: large.Timestamp, UserID: large.UserID,
		TraceID: large.TraceID, SpanID: large.SpanID, Source: large.Source,
		TenantID: large.TenantID, CorrelationID: large.CorrelationID, OccurredAt: large.OccurredAt,
	}

	// Log sizes for all shapes and modes
	smallMap, _ := cborCodec.Encode(small)
	smallArrData, _ := cborCodec.Encode(smallArr)
	smallKIData, _ := cborCodec.Encode(smallKI)

	medMap, _ := cborCodec.Encode(order)
	medArrData, _ := cborCodec.Encode(orderArr)
	medKIData, _ := cborCodec.Encode(orderKI)

	largeMap, _ := cborCodec.Encode(large)
	largeArrData, _ := cborCodec.Encode(largeArr)
	largeKIData, _ := cborCodec.Encode(largeKI)

	b.Log("--- CBOR tag tradeoff: payload sizes (bytes) ---")
	b.Logf("  small  (3 fields):  map=%d  toarray=%d  keyasint=%d", len(smallMap), len(smallArrData), len(smallKIData))
	b.Logf("  medium (7 fields):  map=%d  toarray=%d  keyasint=%d", len(medMap), len(medArrData), len(medKIData))
	b.Logf("  large  (12 fields): map=%d  toarray=%d  keyasint=%d", len(largeMap), len(largeArrData), len(largeKIData))

	cases := []struct {
		name string
		v    any
	}{
		{"small/map", small},
		{"small/toarray", smallArr},
		{"small/keyasint", smallKI},
		{"medium/map", order},
		{"medium/toarray", orderArr},
		{"medium/keyasint", orderKI},
		{"large/map", large},
		{"large/toarray", largeArr},
		{"large/keyasint", largeKI},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if _, err := cborCodec.Encode(tc.v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkTagTradeoffs_Decode measures decode speed and allocation across
// default (map), toarray, and keyasint for small, medium, and large payloads.
func BenchmarkTagTradeoffs_Decode(b *testing.B) {
	cborCodec := codec.CBORCodec{}

	small := sampleSmallEvent()
	smallArr := smallEventArray{ID: small.ID, Type: small.Type, Data: small.Data}
	smallKI := smallEventKeyInt{ID: small.ID, Type: small.Type, Data: small.Data}

	order := sampleOrder()
	orderArr := realisticOrderArray{
		OrderID: order.OrderID, CustomerID: order.CustomerID, Items: order.Items,
		TotalCents: order.TotalCents, Currency: order.Currency, Status: order.Status, CreatedAt: order.CreatedAt,
	}
	orderKI := realisticOrderKeyInt{
		OrderID: order.OrderID, CustomerID: order.CustomerID, Items: order.Items,
		TotalCents: order.TotalCents, Currency: order.Currency, Status: order.Status, CreatedAt: order.CreatedAt,
	}

	large := sampleLargeEvent()
	largeArr := largeEventArray{
		EventID: large.EventID, AggregateID: large.AggregateID, EventType: large.EventType,
		Version: large.Version, Timestamp: large.Timestamp, UserID: large.UserID,
		TraceID: large.TraceID, SpanID: large.SpanID, Source: large.Source,
		TenantID: large.TenantID, CorrelationID: large.CorrelationID, OccurredAt: large.OccurredAt,
	}
	largeKI := largeEventKeyInt{
		EventID: large.EventID, AggregateID: large.AggregateID, EventType: large.EventType,
		Version: large.Version, Timestamp: large.Timestamp, UserID: large.UserID,
		TraceID: large.TraceID, SpanID: large.SpanID, Source: large.Source,
		TenantID: large.TenantID, CorrelationID: large.CorrelationID, OccurredAt: large.OccurredAt,
	}

	smallMap, _ := cborCodec.Encode(small)
	smallArrData, _ := cborCodec.Encode(smallArr)
	smallKIData, _ := cborCodec.Encode(smallKI)

	medMap, _ := cborCodec.Encode(order)
	medArrData, _ := cborCodec.Encode(orderArr)
	medKIData, _ := cborCodec.Encode(orderKI)

	largeMap, _ := cborCodec.Encode(large)
	largeArrData, _ := cborCodec.Encode(largeArr)
	largeKIData, _ := cborCodec.Encode(largeKI)

	cases := []struct {
		name string
		data []byte
		new  func() any
	}{
		{"small/map", smallMap, func() any { return &smallEvent{} }},
		{"small/toarray", smallArrData, func() any { return &smallEventArray{} }},
		{"small/keyasint", smallKIData, func() any { return &smallEventKeyInt{} }},
		{"medium/map", medMap, func() any { return &realisticOrder{} }},
		{"medium/toarray", medArrData, func() any { return &realisticOrderArray{} }},
		{"medium/keyasint", medKIData, func() any { return &realisticOrderKeyInt{} }},
		{"large/map", largeMap, func() any { return &largeEvent{} }},
		{"large/toarray", largeArrData, func() any { return &largeEventArray{} }},
		{"large/keyasint", largeKIData, func() any { return &largeEventKeyInt{} }},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if err := cborCodec.Decode(tc.data, tc.new()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCBORReflectionCache measures the cost of CBOR encoding with a warm
// reflection cache vs a cold cache. fxamacker/cbor caches type metadata in a
// process-wide sync.Map (keyed by reflect.Type). The first encode of each type
// pays the reflection cost; subsequent encodes of the same type reuse the
// cached metadata.
//
// To compare cold vs warm:
//   - Run with -benchtime=1x to see first-encode (cold) cost
//   - Run with -benchtime=10000x to see warm-cache cost
//   - The ratio between the two shows the caching benefit
//
// In practice, the cache is always warm after the first event of each type,
// so steady-state performance is what matters for high-volume event streams.
// Code generation is NOT needed — the cache handles it.
func BenchmarkCBORReflectionCache(b *testing.B) {
	cborCodec := codec.CBORCodec{}
	order := sampleOrder()

	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			if _, err := cborCodec.Encode(order); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("decode", func(b *testing.B) {
		data, _ := cborCodec.Encode(order)

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result realisticOrder
			if err := cborCodec.Decode(data, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkObserveCodec quantifies the overhead of the ObservableCodec
// decorator vs. the raw codec. The decorator adds a mutex lock + counter
// increment per operation. This benchmark informs the atomics-vs-RWMutex
// refactor decision.
func BenchmarkObserveCodec(b *testing.B) {
	cborCodec := codec.CBORCodec{}
	obs := codec.ObserveCodec(cborCodec)
	order := sampleOrder()
	data, _ := cborCodec.Encode(order)

	b.Run("encode/raw", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			if _, err := cborCodec.Encode(order); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("encode/observed", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			if _, err := obs.Encode(order); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("decode/raw", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result realisticOrder
			if err := cborCodec.Decode(data, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("decode/observed", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			var result realisticOrder
			if err := obs.Decode(data, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("encode_pooled/observed", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			if err := codec.EncodePooled(obs, order, func([]byte) error { return nil }); err != nil {
				b.Fatal(err)
			}
		}
	})
}
