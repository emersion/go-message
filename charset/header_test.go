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
