package charset

import (
	"testing"
)

func TestHeader(t *testing.T) {
	s := "¡Hola, señor!"

	enc := EncodeHeader(s)
	dec, err := DecodeHeader(enc)

	if err != nil {
		t.Error("Expected no error while decoding header, got:", err)
	} else if s != dec {
		t.Errorf("Expected decoded string to be %q but got %q", s, dec)
	}
}

func TestDecodeHeader_unknownCharset(t *testing.T) {
	enc := "=?idontexist?q?Hey you?="

	dec, err := DecodeHeader(enc)

	if err == nil {
		t.Error("Expected an error while decoding invalid header")
	}
	if dec != enc {
		t.Errorf("Expected decoded string to fallback to %q but got %q", enc, dec)
	}
}
