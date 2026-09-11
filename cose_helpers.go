package codec

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// decodeBstr decodes a CBOR byte string, accepting nil as an empty byte string.
func decodeBstr(r cbor.RawMessage) ([]byte, error) {
	if isNil(r) {
		return []byte{}, nil
	}

	return decodeCBORRaw[[]byte](r)
}

// decodeOptionalBstr decodes a CBOR byte string or nil value into a byte slice.
// A nil CBOR value returns nil to represent an absent optional field (e.g.,
// detached payload or detached ciphertext).
func decodeOptionalBstr(r cbor.RawMessage) ([]byte, error) {
	if isNil(r) {
		return nil, nil
	}

	return decodeCBORRaw[[]byte](r)
}

// decodeIntMap decodes a CBOR map with integer keys.
func decodeIntMap(r cbor.RawMessage) (map[int64]any, error) {
	if isNil(r) {
		return nil, nil //nolint:nilnil // nil represents absent optional header map
	}

	return decodeCBORRaw[map[int64]any](r)
}

// isNil reports whether r is a CBOR nil value.
func isNil(r cbor.RawMessage) bool {
	return len(r) == 1 && r[0] == 0xf6
}

// decodeCBORRaw decodes r into a fresh T using CBORDecMode. Failures are
// returned unwrapped: the COSE boundary functions classify them with the
// failing message part's stable code (e.g. codec.cose_sign1_protected), so
// this helper must not add a competing layer. Callers handle nil (isNil)
// before calling so this helper is the pure decode tail shared by decodeBstr,
// decodeOptionalBstr, decodeIntMap, and the COSE protected-header decode.
func decodeCBORRaw[T any](r cbor.RawMessage) (T, error) {
	var out T

	if err := CBORDecMode().Unmarshal(r, &out); err != nil {
		return out, err //nolint:wrapcheck // COSE boundary classifies with the part's code
	}

	return out, nil
}

// diagnoseOrError returns the CBOR diagnostic notation of data, or a stable
// "<diagnose failed: ...>" placeholder when the input cannot be diagnosed.
// Shared by COSESign1String and COSEEncrypt0String — both render arbitrary
// COSE messages identically.
func diagnoseOrError(data []byte) string {
	diag, err := cbor.Diagnose(data)
	if err != nil {
		return fmt.Sprintf("<diagnose failed: %v>", err)
	}

	return diag
}

// COSESign1Diagnostic returns a human-readable diagnostic notation of a COSE_Sign1
// message for debugging. It panics only if CBOR diagnosis itself fails, which
// cannot happen for valid CBOR data.
func COSESign1Diagnostic(data []byte) string {
	return diagnoseOrError(data)
}

// COSEEncrypt0Diagnostic returns a human-readable diagnostic notation of a
// COSE_Encrypt0 message for debugging.
func COSEEncrypt0Diagnostic(data []byte) string {
	return diagnoseOrError(data)
}
