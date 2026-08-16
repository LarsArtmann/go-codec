package codec

import "fmt"

// envelopeMagic is the marker value that identifies envelope-wrapped data.
// Its presence in the Magic field confirms the data is an envelope, not raw.
const envelopeMagic = "gcdc"

// envelope wraps serialized data with its encoding format tag, making blind
// stores self-describing (like events are with evt.Encoding()).
type envelope struct {
	Magic    string   `json:"$"`   // always "cqrs" — distinguishes envelope from raw data
	Encoding Encoding `json:"enc"` // codec encoding: "json" or "cbor"
	Data     []byte   `json:"dat"` // inner serialized data
}

// WrapEncode serializes v with codec c and wraps the result in a JSON envelope.
// The envelope itself is always JSON-encoded so it can be read without knowing
// the inner codec. Callers should use [UnwrapDecode] on the read path to
// extract the codec and inner data, or fall back to raw decode for old data.
func WrapEncode(v any, c Codec) ([]byte, error) {
	inner, err := c.Encode(v)
	if err != nil {
		return nil, fmt.Errorf("codec: encode for envelope: %w", err)
	}

	env := envelope{Magic: envelopeMagic, Encoding: c.Encoding(), Data: inner}

	out, err := jsonMarshalDet(env)
	if err != nil {
		return nil, fmt.Errorf("codec: marshal envelope: %w", err)
	}

	return out, nil
}

// UnwrapDecode inspects data for an envelope wrapper. If found, it returns the
// stamped codec and inner data bytes. If not found (old unenveloped data), it
// returns the fallback codec and the original data unchanged.
//
// This enables backward-compatible migration: old data decodes via the fallback,
// new data decodes via the stamped codec. Stores can gradually transition
// without a full clear-and-rebuild.
func UnwrapDecode(data []byte, fallback Codec) (Codec, []byte) {
	// Fast path: the envelope is always a JSON object, so it starts with '{'.
	// Any first byte >= cborMinMajorType (0x80) — CBOR arrays (0x80-0x9f),
	// maps (0xa0-0xbf), tags (0xc0-0xdf), simple values — can never begin
	// valid JSON, so the parse below is doomed. Skip it and fall straight
	// through to the fallback codec. Bytes below 0x80 stay on the original
	// path: they are either JSON (envelope or legacy raw JSON) or CBOR
	// scalars/strings, which the parse-fail branch already handles.
	if len(data) > 0 && data[0] >= cborMinMajorType {
		return fallback, data
	}

	var env envelope
	if err := jsonUnmarshal(data, &env); err == nil &&
		env.Magic == envelopeMagic && len(env.Data) > 0 {
		if c, err := ForEncoding(env.Encoding); err == nil {
			return c, env.Data
		}
	}

	return fallback, data
}

// DecodeEnvelopeOrLegacy decodes data that may be envelope-wrapped or raw
// (pre-envelope) into T, keeping the permanent-readability guarantee of
// [UnwrapDecode] while honoring the caller's configured codec:
//
//   - Envelope-wrapped data decodes via its stamped codec (first attempt).
//   - Raw data decodes via the configured codec first.
//   - If that fails and the codec is one of the two standard codecs, exactly
//     one cross-retry with the other (JSON↔CBOR) rescues legacy rows written
//     by a differently-configured store.
//   - Custom codecs get the configured-codec attempt only — their data only
//     decodes with themselves.
//
// Truly undecodable data returns the primary decode error unwrapped, so
// callers can classify it (e.g. as Corruption) with their own context.
func DecodeEnvelopeOrLegacy[T any](data []byte, configured Codec) (T, error) {
	c, inner := UnwrapDecode(data, configured)

	var val T

	err := c.Decode(inner, &val)
	if err == nil {
		return val, nil
	}

	alt, ok := otherStandardCodec(c)
	if !ok {
		return val, err //nolint:wrapcheck // callers classify with their own context
	}

	var retry T

	if altErr := alt.Decode(inner, &retry); altErr == nil {
		return retry, nil
	}

	return val, err //nolint:wrapcheck // callers classify with their own context
}

// otherStandardCodec returns the opposite built-in codec, or false for
// envelope-stamped or custom codecs (their data only decodes with themselves).
func otherStandardCodec(c Codec) (Codec, bool) {
	switch c.(type) {
	case CBORCodec:
		return JSONCodec{}, true
	case JSONCodec:
		return CBORCodec{}, true
	default:
		return nil, false
	}
}
