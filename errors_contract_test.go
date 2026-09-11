package codec_test

import (
	"errors"
	"testing"

	cbor "github.com/fxamacker/cbor/v2"
	"github.com/larsartmann/go-codec"
	errorfamily "github.com/larsartmann/go-error-family"
	. "github.com/onsi/gomega"
)

// These tests lock the codec error contract: every returned error carries a
// stable machine-readable code, a behavioral family, and — where the caller
// supplied the offending value — structured context. Sentinel identity is
// preserved through wrapping (errors.Is), and inner classifications are never
// double-wrapped (WrapOnce at orchestration boundaries).

func TestForEncoding_ErrorContract(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	_, err := codec.ForEncoding("bogus")
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, codec.ErrUnknownEncoding)).To(BeTrue())

	e, ok := errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(BeTrue())
	g.Expect(e.Code()).To(Equal("codec.unknown_encoding"))
	g.Expect(e.ErrorContext()["encoding"]).To(Equal("bogus"))
}

func TestUnmarshalCOSESign1_ErrorContract(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	// Not CBOR at all → corruption with the unmarshal code.
	_, err := codec.UnmarshalCOSESign1([]byte("not cbor"))
	g.Expect(err).To(HaveOccurred())

	e, ok := errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(BeTrue())
	g.Expect(e.Code()).To(Equal("codec.cose_sign1_unmarshal"))
	g.Expect(e.ErrorFamily()).To(Equal(errorfamily.Corruption))

	// Wrong element count → sentinel identity AND the detail code.
	tooShort, merr := cbor.Marshal([]any{"only-one"})
	g.Expect(merr).NotTo(HaveOccurred())

	_, err = codec.UnmarshalCOSESign1(tooShort)
	g.Expect(errors.Is(err, codec.ErrInvalidCOSESign1)).To(BeTrue())

	e, ok = errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(BeTrue())
	g.Expect(e.Code()).To(Equal("codec.cose_sign1_element_count"))

	// Structurally valid array, unreadable protected header → part code.
	badProtected, merr := cbor.Marshal([]any{"oops", map[int64]any{}, []byte(nil), []byte("sig")})
	g.Expect(merr).NotTo(HaveOccurred())

	_, err = codec.UnmarshalCOSESign1(badProtected)
	g.Expect(err).To(HaveOccurred())

	e, ok = errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(BeTrue())
	g.Expect(e.Code()).To(Equal("codec.cose_sign1_protected"))
}

func TestUnmarshalCOSEEncrypt0_ErrorContract(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	tooShort, merr := cbor.Marshal([]any{"only-one"})
	g.Expect(merr).NotTo(HaveOccurred())

	_, err := codec.UnmarshalCOSEEncrypt0(tooShort)
	g.Expect(errors.Is(err, codec.ErrInvalidCOSEEncrypt0)).To(BeTrue())

	e, ok := errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(BeTrue())
	g.Expect(e.Code()).To(Equal("codec.cose_encrypt0_element_count"))
}

func TestWrapEncode_InnerCodePreserved(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	// RawCodec rejects the value with a classified inner error; the envelope
	// boundary must propagate it unchanged (WrapOnce), keeping the inner
	// sentinel identity and code instead of stacking a second one.
	_, err := codec.WrapEncode("not bytes", codec.RawCodec{})
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, codec.ErrEncodeRawType)).To(BeTrue())

	e, ok := errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(BeTrue())
	g.Expect(e.Code()).To(Equal("codec.raw_encode_type"))
	g.Expect(err.Error()).NotTo(ContainSubstring("envelope_encode"))
}

func TestTranscodeToJSON_ErrorContract(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	_, err := codec.TranscodeToJSON([]byte("garbage"), codec.EncodingCBOR)
	g.Expect(err).To(HaveOccurred())

	e, ok := errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(BeTrue())
	g.Expect(e.Code()).To(Equal("codec.transcode_decode"))
	g.Expect(e.ErrorFamily()).To(Equal(errorfamily.Corruption))
}

func TestErrorSentinels_CarryStableCodes(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	for _, tc := range []struct {
		sentinel error
		code     string
	}{
		{codec.ErrUnknownEncoding, "codec.unknown_encoding"},
		{codec.ErrEncodeRawType, "codec.raw_encode_type"},
		{codec.ErrDecodeRawType, "codec.raw_decode_type"},
		{codec.ErrInvalidCOSESign1, "codec.invalid_cose_sign1"},
		{codec.ErrInvalidCOSEEncrypt0, "codec.invalid_cose_encrypt0"},
		{codec.ErrCOSEAlgorithmOverflow, "codec.cose_algorithm_overflow"},
		{codec.ErrCOSEInvalidAlgorithm, "codec.cose_invalid_algorithm"},
		{codec.ErrNormalizeDepthExceeded, "codec.normalize_depth_exceeded"},
	} {
		g.Expect(tc.sentinel).To(HaveOccurred())
		g.Expect(errorfamily.Code(tc.sentinel)).To(Equal(tc.code))
	}
}
