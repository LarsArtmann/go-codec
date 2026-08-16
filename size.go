package codec

// SizeResult reports the serialized byte sizes of a value under each codec.
type SizeResult struct {
	JSON int `json:"json"`
	CBOR int `json:"cbor"`
}

// Size encodes v with both JSON and CBOR and returns the byte sizes. This is
// useful for deciding whether switching a payload type from JSON to CBOR is
// worthwhile before committing to a format change.
//
// If a codec fails to encode v, that size is reported as -1.
//
//	s := codec.Size(UserCreated{Name: "Alice", Email: "a@b.c"})
//	savings := float64(s.JSON-s.CBOR) / float64(s.JSON) * 100 // e.g. 19
func Size(v any) SizeResult {
	jsonData, err := (JSONCodec{}).Encode(v)
	if err != nil {
		cborData, cborErr := (CBORCodec{}).Encode(v)
		if cborErr != nil {
			return SizeResult{JSON: -1, CBOR: -1}
		}

		return SizeResult{JSON: -1, CBOR: len(cborData)}
	}

	cborData, err := (CBORCodec{}).Encode(v)
	if err != nil {
		return SizeResult{JSON: len(jsonData), CBOR: -1}
	}

	return SizeResult{JSON: len(jsonData), CBOR: len(cborData)}
}
