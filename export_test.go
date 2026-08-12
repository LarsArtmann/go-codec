package codec

// export_test.go exposes internal helpers to the external test package
// (package codec_test) so that tests can verify internal behavior through
// the public API surface rather than white-box access.

//nolint:gochecknoglobals // test-only exports, standard Go pattern
var (
	CanonicalEncMode = canonicalEncMode
	CanonicalDecMode = canonicalDecMode
	JSONUnmarshal    = jsonUnmarshal
	EnvelopeMagic    = envelopeMagic
)

type (
	RawJSONValue = rawJSONValue
	Envelope     = envelope
)
