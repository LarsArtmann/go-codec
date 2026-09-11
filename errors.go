package codec

import errorfamily "github.com/larsartmann/go-error-family"

var (
	// Stable sentinel identities. Declared as the `error` interface so
	// errors.Is call sites match the sentinel guard; every in-package error
	// wraps these with errorfamily.Wrapf using the SAME code, so wrapped
	// errors still match via code+family identity.
	ErrEncodeRawType error = errorfamily.NewRejection(
		"codec.raw_encode_type",
		"raw codec: expected []byte",
	)
	ErrDecodeRawType error = errorfamily.NewRejection(
		"codec.raw_decode_type",
		"raw codec: expected *[]byte target",
	)
	ErrInvalidCOSESign1 error = errorfamily.NewRejection(
		"codec.invalid_cose_sign1",
		"COSE_Sign1 structure has an invalid number of elements",
	)
	ErrInvalidCOSEEncrypt0 error = errorfamily.NewRejection(
		"codec.invalid_cose_encrypt0",
		"COSE_Encrypt0 structure has an invalid number of elements",
	)
	ErrCOSEAlgorithmOverflow error = errorfamily.NewRejection(
		"codec.cose_algorithm_overflow",
		"COSE algorithm value overflows int64",
	)
	ErrCOSEInvalidAlgorithm error = errorfamily.NewRejection(
		"codec.cose_invalid_algorithm",
		"COSE algorithm value is not an integer",
	)
	ErrNormalizeDepthExceeded error = errorfamily.NewRejection(
		"codec.normalize_depth_exceeded",
		"codec: normalizeForJSON recursion depth exceeded",
	)
)
