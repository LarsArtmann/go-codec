package codec_test

import (
	"encoding"
	"testing"

	"github.com/larsartmann/go-codec"
	"github.com/onsi/gomega"
)

// fakeBrandedID mirrors the go-branded-id cbid.ID[B, V] wire contract: it
// implements BOTH encoding.TextMarshaler/TextUnmarshaler and
// encoding.BinaryMarshaler/BinaryUnmarshaler over the same raw string value.
// fxamacker encodes BinaryMarshaler types as CBOR byte strings natively, so
// the codec's DecMode must still accept the historical TEXT-string shape via
// UnmarshalText — otherwise every journal written before branded IDs landed
// becomes unreadable the moment payload fields turn branded.
type fakeBrandedID struct{ value string }

func (id fakeBrandedID) MarshalText() ([]byte, error) { return []byte(id.value), nil }

func (id *fakeBrandedID) UnmarshalText(data []byte) error {
	id.value = string(data)

	return nil
}

var (
	_ encoding.TextMarshaler   = fakeBrandedID{}
	_ encoding.TextUnmarshaler = (*fakeBrandedID)(nil)
)

func (id fakeBrandedID) MarshalBinary() ([]byte, error) { return []byte(id.value), nil }

func (id *fakeBrandedID) UnmarshalBinary(data []byte) error {
	id.value = string(data)

	return nil
}

var _ encoding.BinaryMarshaler = fakeBrandedID{}

// fakeLegacyPayload is the pre-branded wire shape: the ID field is a plain
// string, which canonical CBOR encodes as a TEXT string.
type fakeLegacyPayload struct {
	ID   string
	Name string
}

type fakeBrandedPayload struct {
	ID   fakeBrandedID
	Name string
}

func TestCBORCodecsDecodeTextStringIntoTextUnmarshaler(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	legacy := fakeLegacyPayload{ID: "11111111-1111-4111-8111-111111111111", Name: "x.jpg"}

	for name, c := range map[string]codec.Codec{
		"canonical": codec.CBORCodec{},
		"compact":   codec.CBORCompactCodec{},
	} {
		data, err := c.Encode(legacy)
		g.Expect(err).NotTo(gomega.HaveOccurred())

		var decoded fakeBrandedPayload

		err = c.Decode(data, &decoded)
		g.Expect(err).NotTo(gomega.HaveOccurred(), "%s codec must decode a text-string ID via UnmarshalText", name)
		g.Expect(decoded.ID.value).To(gomega.Equal(legacy.ID), name)
		g.Expect(decoded.Name).To(gomega.Equal(legacy.Name), name)
	}
}

func TestCBORCodecsRoundTripBrandedIDByteString(t *testing.T) {
	t.Parallel()

	g := gomega.NewWithT(t)

	original := fakeBrandedPayload{ID: fakeBrandedID{value: "22222222-2222-4222-8222-222222222222"}, Name: "y.jpg"}

	for name, c := range map[string]codec.Codec{
		"canonical": codec.CBORCodec{},
		"compact":   codec.CBORCompactCodec{},
	} {
		data, err := c.Encode(original)
		g.Expect(err).NotTo(gomega.HaveOccurred())

		var decoded fakeBrandedPayload

		err = c.Decode(data, &decoded)
		g.Expect(err).NotTo(gomega.HaveOccurred(), "%s codec must keep the byte-string ID path working", name)
		g.Expect(decoded.ID.value).To(gomega.Equal(original.ID.value), name)
	}
}
