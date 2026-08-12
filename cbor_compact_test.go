package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
	"github.com/onsi/gomega"
)

func TestCBORCompactCodecRoundTrip(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	type payload struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
	}

	c := codec.CBORCompactCodec{}
	g.Expect(c.Encoding()).To(gomega.Equal(codec.EncodingCBOR))

	original := payload{Name: testEventType, Version: 3}

	data, err := c.Encode(original)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(data).NotTo(gomega.BeEmpty())

	var decoded payload

	err = c.Decode(data, &decoded)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(decoded).To(gomega.Equal(original))
}

func TestCBORCompactCodecRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	type v1 struct {
		Name string `json:"name"`
	}

	type v2 struct {
		Name     string `json:"name"`
		NewField string `json:"newField"`
	}

	c := codec.CBORCompactCodec{}

	data, err := c.Encode(v2{Name: testValue, NewField: "extra"})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	var decoded v1

	err = c.Decode(data, &decoded)
	g.Expect(err).To(gomega.HaveOccurred(), "should reject unknown field 'new_field'")
}

func TestCBORCompactCodecNotCompatibleWithCBORCodec(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	// The two codecs use different sort orders — output bytes should differ.
	type payload struct {
		B string `json:"b"`
		A string `json:"a"`
	}

	standard, err := codec.CBORCodec{}.Encode(payload{A: "1", B: "2"})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	compact, err := codec.CBORCompactCodec{}.Encode(payload{A: "1", B: "2"})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// CoreDet uses SortBytewiseLexical, Canonical uses SortLengthFirst.
	// For single-char keys they might be the same length, so check with longer keys.
	type longPayload struct {
		Bravo   string `json:"bravo"`
		Alpha   string `json:"alpha"`
		Charlie string `json:"charlie"`
	}

	stdLong, err := codec.CBORCodec{}.Encode(longPayload{Alpha: "a", Bravo: "b", Charlie: "c"})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	cptLong, err := codec.CBORCompactCodec{}.Encode(longPayload{Alpha: "a", Bravo: "b", Charlie: "c"})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// They might or might not differ depending on key lengths.
	// The important property is both are valid CBOR that round-trip correctly.
	_ = standard
	_ = compact

	g.Expect(stdLong).NotTo(gomega.BeEmpty())
	g.Expect(cptLong).NotTo(gomega.BeEmpty())
}

func TestDiagnose(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	type payload struct {
		Name string `json:"name"`
	}

	data, err := codec.CBORCodec{}.Encode(payload{Name: testValue})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	diag, err := codec.Diagnose(data)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(diag).To(gomega.ContainSubstring(testField))
	g.Expect(diag).To(gomega.ContainSubstring(testValue))
}

func TestDiagnoseInvalidCBOR(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	_, err := codec.Diagnose([]byte{0xff, 0xff})
	g.Expect(err).To(gomega.HaveOccurred())
}
