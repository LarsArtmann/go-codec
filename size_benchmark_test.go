package codec_test

import (
	"testing"

	"github.com/larsartmann/go-codec"
)

// BenchmarkSize measures the Size helper, which encodes the value with both
// JSON and CBOR to report their byte sizes.
func BenchmarkSize(b *testing.B) {
	type UserCreated struct {
		Name  string
		Email string
		Time  int64
	}

	v := UserCreated{Name: testName, Email: testEmail, Time: 1700000000}

	b.ReportAllocs()

	for b.Loop() {
		_ = codec.Size(v)
	}
}
