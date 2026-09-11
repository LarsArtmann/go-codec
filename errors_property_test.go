package codec_test

import (
	"errors"
	"math"
	"testing"

	"github.com/larsartmann/go-codec"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/onsi/gomega"
	"pgregory.net/rapid"
)

// Property: every error that escapes the CLASSIFIED API surface is a
// *errorfamily.Error carrying a codec.-prefixed code and one of the four
// behavioral families. This mechanically catches future unclassified wraps
// (fmt.Errorf regressions) at the API boundary.
//
// Scope is the wrapping surface registered in docs/error-codes.md. The thin
// codec wrappers (CBORCodec/JSONCodec Encode/Decode) intentionally return raw
// library errors for caller classification — see docs/adr/0001-error-taxonomy.md.

func isValidFamily(f errorfamily.Family) bool {
	switch f {
	case errorfamily.Rejection, errorfamily.Corruption, errorfamily.Infrastructure,
		errorfamily.Orchestration, errorfamily.Conflict, errorfamily.Transient:
		return true
	default:
		return false
	}
}

func assertClassified(g *gomega.WithT, op string, err error) {
	if err == nil {
		return
	}

	e, ok := errors.AsType[*errorfamily.Error](err)
	g.Expect(ok).To(gomega.BeTrue(), "%s returned an unclassified error: %v", op, err)

	code := e.Code()
	g.Expect(code).To(gomega.HavePrefix("codec."), "%s returned code %q", op, code)
	g.Expect(isValidFamily(e.ErrorFamily())).To(gomega.BeTrue(),
		"%s returned unknown family %v for code %q", op, e.ErrorFamily(), code)
}

// genUnencodable produces values that CBOR/JSON cannot encode, so the encode
// paths reliably fail and exercise their wrap sites.
func genUnencodable(t *rapid.T) any {
	switch rapid.IntRange(0, 4).Draw(t, "kind") {
	case 0:
		return make(chan int)
	case 1:
		return func() {}
	case 2:
		return complex(float64(rapid.Int().Draw(t, "re")), float64(rapid.Int().Draw(t, "im")))
	case 3:
		return &struct{ Ch chan struct{} }{}

	default:
		return map[string]any{"ch": make(chan int)}
	}
}

func genNonBytes(t *rapid.T) any {
	switch rapid.IntRange(0, 3).Draw(t, "kind") {
	case 0:
		return rapid.Int().Draw(t, "int")
	case 1:
		return rapid.StringN(0, 20, 50).Draw(t, "string")
	case 2:
		return rapid.Float64().Draw(t, "float")

	default:
		return &struct{ ID string }{ID: rapid.StringN(0, 10, 20).Draw(t, "id")}
	}
}

func genWrongTarget(t *rapid.T) any {
	switch rapid.IntRange(0, 3).Draw(t, "kind") {
	case 0:
		return new(int)
	case 1:
		return new(string)
	case 2:
		return new(float64)

	default:
		return new(struct{ A, B int })
	}
}

func genNonAlgorithm(t *rapid.T) any {
	switch rapid.IntRange(0, 3).Draw(t, "kind") {
	case 0:
		return rapid.StringN(0, 10, 20).Draw(t, "string")
	case 1:
		return rapid.Float64().Draw(t, "float")
	case 2:
		return true

	default:
		return []any{rapid.Int().Draw(t, "elem")}
	}
}

func TestProperty_ClassifiedErrorsCarryStableCodes(t *testing.T) {
	t.Parallel()

	ops := []struct {
		name string
		call func(t *rapid.T) error
	}{
		{"ForEncoding", func(t *rapid.T) error {
			enc := codec.Encoding(rapid.StringN(1, 8, 16).
				Filter(func(s string) bool { return s != "json" && s != "cbor" && s != "raw" }).
				Draw(t, "encoding"))
			_, err := codec.ForEncoding(enc)

			return err
		}},
		{"RawCodec.Encode", func(t *rapid.T) error {
			_, err := codec.RawCodec{}.Encode(genNonBytes(t))

			return err
		}},
		{"RawCodec.Decode", func(t *rapid.T) error {
			data := rapid.SliceOfN(rapid.Byte(), 0, 32).Draw(t, "data")

			return codec.RawCodec{}.Decode(data, genWrongTarget(t))
		}},
		{"NormalizeCOSEAlgorithm.invalid", func(t *rapid.T) error {
			_, err := codec.NormalizeCOSEAlgorithm(genNonAlgorithm(t))

			return err
		}},
		{"NormalizeCOSEAlgorithm.overflow", func(t *rapid.T) error {
			_, err := codec.NormalizeCOSEAlgorithm(rapid.Uint64Range(
				uint64(math.MaxInt64)+1, math.MaxUint64).Draw(t, "overflow"))

			return err
		}},
		{"DecodeBase64String", func(t *rapid.T) error {
			_, err := codec.DecodeBase64String(rapid.StringN(1, 24, 48).Draw(t, "b64"))

			return err
		}},
		{"UnmarshalCOSESign1", func(t *rapid.T) error {
			data := rapid.SliceOfN(rapid.Byte(), 0, 48).Draw(t, "data")
			_, err := codec.UnmarshalCOSESign1(data)

			return err
		}},
		{"UnmarshalCOSEEncrypt0", func(t *rapid.T) error {
			data := rapid.SliceOfN(rapid.Byte(), 0, 48).Draw(t, "data")
			_, err := codec.UnmarshalCOSEEncrypt0(data)

			return err
		}},
		{"UnmarshalCOSEProtectedHeader", func(t *rapid.T) error {
			data := rapid.SliceOfN(rapid.Byte(), 0, 48).Draw(t, "data")
			_, err := codec.UnmarshalCOSEProtectedHeader(data)

			return err
		}},
		{"TranscodeToJSON", func(t *rapid.T) error {
			data := rapid.SliceOfN(rapid.Byte(), 0, 48).Draw(t, "data")
			_, err := codec.TranscodeToJSON(data, codec.EncodingCBOR)

			return err
		}},
		{"WrapEncode.raw", func(t *rapid.T) error {
			_, err := codec.WrapEncode(genNonBytes(t), codec.RawCodec{})

			return err
		}},
		{"WrapEncode.cbor", func(t *rapid.T) error {
			_, err := codec.WrapEncode(genUnencodable(t), codec.CBORCodec{})

			return err
		}},
		{"EncodePooled", func(t *rapid.T) error {
			return codec.EncodePooled(codec.CBORCodec{}, genUnencodable(t), func([]byte) error {
				return nil
			})
		}},
	}

	rapid.Check(t, func(t *rapid.T) {
		g := gomega.NewWithT(t)
		op := rapid.IntRange(0, len(ops)-1).Draw(t, "op")
		assertClassified(g, ops[op].name, ops[op].call(t))
	})
}
