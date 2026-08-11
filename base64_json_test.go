package codec

import (
	"bytes"
	"errors"
	"testing"
)

func TestDecodeBase64String_URLSafe(t *testing.T) {
	t.Parallel()

	// "-" and "_" are URL-safe base64 characters (would fail StdEncoding)
	raw := []byte{0xff, 0x00, 0xfe}
	encoded := "_wD-" // base64.URLEncoding

	got, err := DecodeBase64String(encoded)
	if err != nil {
		t.Fatalf("DecodeBase64String: %v", err)
	}

	if !bytes.Equal(got, raw) {
		t.Errorf("got %v, want %v", got, raw)
	}
}

func TestDecodeBase64String_StandardFallback(t *testing.T) {
	t.Parallel()

	// "+" and "/" are standard base64 characters (not URL-safe)
	raw := []byte{0xfb, 0xff, 0xbf}
	encoded := "+/+/" // base64.StdEncoding (has +)

	got, err := DecodeBase64String(encoded)
	if err != nil {
		t.Fatalf("DecodeBase64String: %v", err)
	}

	if !bytes.Equal(got, raw) {
		t.Errorf("got %v, want %v", got, raw)
	}
}

func TestDecodeBase64String_Invalid(t *testing.T) {
	t.Parallel()

	_, err := DecodeBase64String("!!!not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestMarshalBase64JSON(t *testing.T) {
	t.Parallel()

	raw := []byte{0x01, 0x02, 0x03}
	got, err := MarshalBase64JSON(raw)
	if err != nil {
		t.Fatalf("MarshalBase64JSON: %v", err)
	}

	// Should be a quoted base64 URL-safe string
	if string(got) != `"AQID"` {
		t.Errorf("got %s, want %q", got, `"AQID"`)
	}
}

func TestMarshalBase64JSONWithModule(t *testing.T) {
	t.Parallel()

	raw := []byte{0x01, 0x02}
	got, err := MarshalBase64JSONWithModule(raw, "test", "data")
	if err != nil {
		t.Fatalf("MarshalBase64JSONWithModule: %v", err)
	}

	if string(got) != `"AQI="` {
		t.Errorf("got %s, want %q", got, `"AQI="`)
	}
}

func TestUnmarshalBase64JSON(t *testing.T) {
	t.Parallel()

	decoded, err := UnmarshalBase64JSON([]byte(`"AQID"`), "test", "data")
	if err != nil {
		t.Fatalf("UnmarshalBase64JSON: %v", err)
	}

	want := []byte{0x01, 0x02, 0x03}
	if !bytes.Equal(decoded, want) {
		t.Errorf("got %v, want %v", decoded, want)
	}
}

func TestUnmarshalBase64JSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := UnmarshalBase64JSON([]byte(`not-json`), "test", "data")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestUnmarshalBase64JSON_InvalidBase64(t *testing.T) {
	t.Parallel()

	_, err := UnmarshalBase64JSON([]byte(`"!!!invalid!!!"`), "test", "data")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestAssignBase64JSON(t *testing.T) {
	t.Parallel()

	var target []byte
	err := AssignBase64JSON([]byte(`"AQID"`), "test", "data", &target)
	if err != nil {
		t.Fatalf("AssignBase64JSON: %v", err)
	}

	want := []byte{0x01, 0x02, 0x03}
	if !bytes.Equal(target, want) {
		t.Errorf("got %v, want %v", target, want)
	}
}

func TestWrapCOSEMarshal_NilError(t *testing.T) {
	t.Parallel()

	_, err := WrapCOSEMarshal(nil, errors.New("marshal failed"), "test", "msg")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWrapCOSEMarshal_Success(t *testing.T) {
	t.Parallel()

	data := []byte{0x01}
	got, err := WrapCOSEMarshal(data, nil, "test", "msg")
	if err != nil {
		t.Fatalf("WrapCOSEMarshal: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Errorf("got %v, want %v", got, data)
	}
}

func TestPrepareCOSESetup(t *testing.T) {
	t.Parallel()

	type config struct {
		KID []byte
	}

	kid := []byte{0x01}

	opts := []func(*config){
		func(c *config) { c.KID = kid },
	}

	protected, err := PrepareCOSESetup(&config{}, opts, COSEAlgAES256GCM)
	if err != nil {
		t.Fatalf("PrepareCOSESetup: %v", err)
	}

	if len(protected) == 0 {
		t.Fatal("expected non-empty protected header")
	}

	// The protected header should decode to {1: 3} (alg=AES256GCM)
	headers, err := UnmarshalCOSEProtectedHeader(protected)
	if err != nil {
		t.Fatalf("UnmarshalCOSEProtectedHeader: %v", err)
	}

	if alg, ok := headers[COSEHeaderAlg]; !ok {
		t.Fatal("missing alg in protected header")
	} else {
		normalized, err := NormalizeCOSEAlgorithm(alg)
		if err != nil {
			t.Fatalf("NormalizeCOSEAlgorithm: %v", err)
		}
		if normalized != COSEAlgAES256GCM {
			t.Errorf("alg = %d, want %d", normalized, COSEAlgAES256GCM)
		}
	}
}
