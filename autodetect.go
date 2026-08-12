package codec

import "fmt"

// cborMinMajorType is the minimum byte value for CBOR major types 4 and 5
// (arrays 0x80-0x9f, maps 0xa0-0xbf). These never start valid JSON.
const cborMinMajorType byte = 0x80

// maxAutoDetectSize bounds the size of trial-decode input. The first-byte
// heuristic above is O(1) and runs on any size, but trial JSON/CBOR decoding
// of oversized payloads is a DoS vector. Payloads larger than this that reach
// the trial-decode fallback return EncodingRaw.
const maxAutoDetectSize = 1 << 20 // 1 MiB

// DetectionReason explains why AutoDetect chose an encoding.
type DetectionReason string

const (
	// DetectionReasonEmpty means the input was empty and defaulted to raw.
	DetectionReasonEmpty DetectionReason = "empty"
	// DetectionReasonCBORMajorType means the first byte identified a CBOR major type.
	DetectionReasonCBORMajorType DetectionReason = "cbor_major_type"
	// DetectionReasonJSONStructure means the first byte was an unambiguous JSON structural start.
	DetectionReasonJSONStructure DetectionReason = "json_structure"
	// DetectionReasonJSONTrialDecode means the first byte was ambiguous, but JSON trial decode succeeded.
	DetectionReasonJSONTrialDecode DetectionReason = "json_trial_decode"
	// DetectionReasonCBORTrialDecode means the first byte was ambiguous and JSON trial decode failed, but CBOR trial decode succeeded.
	DetectionReasonCBORTrialDecode DetectionReason = "cbor_trial_decode"
	// DetectionReasonOversized means the payload exceeded maxAutoDetectSize and was treated as raw.
	DetectionReasonOversized DetectionReason = "oversized"
	// DetectionReasonUnknown means the payload did not match any known format.
	DetectionReasonUnknown DetectionReason = "unknown"
)

// AutoDetectResult carries the encoding chosen by AutoDetect plus the reason.
type AutoDetectResult struct {
	Encoding Encoding
	Reason   DetectionReason
	Detail   string
}

// AutoDetect inspects raw payload bytes and returns the most likely [Encoding].
// It distinguishes JSON from CBOR by examining the structural first byte:
//
//   - JSON objects/arrays/strings start with '{', '[', or '"' (ASCII).
//   - CBOR maps (major type 5) start with 0xa0–0xbf; arrays (major type 4)
//     with 0x80–0x9f. These ranges never overlap with valid JSON leading bytes.
//
// For ambiguous leading bytes (e.g. bare numbers, booleans) the function falls
// back to a trial decode: JSON first, then CBOR.
//
// Empty input returns [EncodingRaw]. Unknown data returns [EncodingRaw].
//
// AutoDetect is a best-effort heuristic for diagnostics and tooling — it is NOT
// a security boundary. Never use it to skip encoding validation; always pair
// detected data with the matching codec's Decode for authoritative parsing.
func AutoDetect(data []byte) Encoding {
	return AutoDetectDebug(data).Encoding
}

// AutoDetectDebug is the verbose form of [AutoDetect]: it returns not only
// the inferred encoding but also the reason and a human-readable detail string.
// Use it when triaging mixed-encoding streams or when logging the decision path.
func AutoDetectDebug(data []byte) AutoDetectResult {
	if len(data) == 0 {
		return AutoDetectResult{
			Encoding: EncodingRaw,
			Reason:   DetectionReasonEmpty,
			Detail:   "empty input defaults to raw",
		}
	}

	first := data[0]

	// CBOR major types 4-7 (0x80-0xff) never start valid JSON.
	if first >= cborMinMajorType {
		return AutoDetectResult{
			Encoding: EncodingCBOR,
			Reason:   DetectionReasonCBORMajorType,
			Detail:   fmt.Sprintf("first byte 0x%02x >= 0x%02x identifies CBOR major type 4-7", first, cborMinMajorType),
		}
	}

	// Unambiguous JSON structural starts.
	switch first {
	case '{', '[', '"':
		return AutoDetectResult{
			Encoding: EncodingJSON,
			Reason:   DetectionReasonJSONStructure,
			Detail:   fmt.Sprintf("first byte %q is an unambiguous JSON structural start", first),
		}
	}

	// JSON keywords and numbers (ASCII letters/digits/signs). These overlap
	// with CBOR major types 0-3, so try JSON decode first.
	if isJSONStart(first) {
		if len(data) > maxAutoDetectSize {
			return AutoDetectResult{
				Encoding: EncodingRaw,
				Reason:   DetectionReasonOversized,
				Detail:   fmt.Sprintf("ambiguous JSON start but payload length %d exceeds maxAutoDetectSize %d", len(data), maxAutoDetectSize),
			}
		}

		var v any

		if err := (JSONCodec{}).Decode(data, &v); err == nil {
			return AutoDetectResult{
				Encoding: EncodingJSON,
				Reason:   DetectionReasonJSONTrialDecode,
				Detail:   fmt.Sprintf("first byte %q is a JSON scalar/keyword start and JSON trial decode succeeded", first),
			}
		}

		return AutoDetectResult{
			Encoding: EncodingCBOR,
			Reason:   DetectionReasonCBORMajorType,
			Detail:   fmt.Sprintf("JSON trial decode failed, treating first byte 0x%02x as CBOR major type 0-3", first),
		}
	}

	// Remaining bytes are either CBOR scalars or unrecognised.
	if len(data) > maxAutoDetectSize {
		return AutoDetectResult{
			Encoding: EncodingRaw,
			Reason:   DetectionReasonOversized,
			Detail:   fmt.Sprintf("payload length %d exceeds maxAutoDetectSize %d", len(data), maxAutoDetectSize),
		}
	}

	var v any

	if err := (CBORCodec{}).Decode(data, &v); err == nil {
		return AutoDetectResult{
			Encoding: EncodingCBOR,
			Reason:   DetectionReasonCBORTrialDecode,
			Detail:   fmt.Sprintf("CBOR trial decode succeeded for first byte 0x%02x", first),
		}
	}

	return AutoDetectResult{
		Encoding: EncodingRaw,
		Reason:   DetectionReasonUnknown,
		Detail:   fmt.Sprintf("first byte 0x%02x does not match any known format", first),
	}
}

// isJSONStart reports whether b is a valid first byte for a JSON value
// (per RFC 8259): object, array, string, number, true, false, null.
func isJSONStart(b byte) bool {
	switch b {
	case '{', '[', '"', '-', 't', 'f', 'n':
		return true
	}

	return b >= '0' && b <= '9'
}
