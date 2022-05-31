package message

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

var testEncodings = []struct {
	enc     string
	encoded string
	decoded string
}{
	{
		enc:     "binary",
		encoded: "café",
		decoded: "café",
	},
	{
		enc:     "8bit",
		encoded: "café",
		decoded: "café",
	},
	{
		enc:     "7bit",
		encoded: "hi there",
		decoded: "hi there",
	},
	{
		enc:     "quoted-printable",
		encoded: "caf=C3=A9",
		decoded: "café",
	},
	{
		enc:     "base64",
		encoded: "Y2Fmw6k=",
		decoded: "café",
	},
	{ // wild emails seen with a single = terminator and no crlf
		enc:     "quoted-printable",
		encoded: "=",
		decoded: "",
	},
}

func TestDecode(t *testing.T) {
	for _, test := range testEncodings {
		r, err := encodingReader(test.enc, strings.NewReader(test.encoded))
		if err != nil {
			t.Errorf("Expected no error when creating decoder for encoding %q, but got: %v", test.enc, err)
		} else if b, err := ioutil.ReadAll(r); err != nil {
			t.Errorf("Expected no error when reading encoding %q, but got: %v", test.enc, err)
		} else if s := string(b); s != test.decoded {
			t.Errorf("Expected decoded text to be %q but got %q", test.decoded, s)
		}
	}
}

func TestDecode_error(t *testing.T) {
	_, err := encodingReader("idontexist", nil)
	if err == nil {
		t.Errorf("Expected an error when creating decoder for invalid encoding")
	}
}

func TestEncode(t *testing.T) {
	for _, test := range testEncodings {
		var b bytes.Buffer
		wc, _ := encodingWriter(test.enc, &b)
		io.WriteString(wc, test.decoded)
		wc.Close()
		if s := b.String(); s != test.encoded {
			t.Errorf("Expected encoded text to be %q but got %q", test.encoded, s)
		}
	}
}
