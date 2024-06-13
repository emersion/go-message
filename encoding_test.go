package message

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	{
		enc:     "iso-8859-1",
		encoded: "caf" + string([]byte{0xe9}) + " ",
		decoded: "café ",
	},
}

func TestDecode(t *testing.T) {
	for _, test := range testEncodings {
		r, err := encodingReader(test.enc, strings.NewReader(test.encoded))
		if err != nil {
			t.Errorf("Expected no error when creating decoder for encoding %q, but got: %v", test.enc, err)
		} else if b, err := io.ReadAll(r); err != nil {
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
		wc, err := encodingWriter(test.enc, &b)
		assert.NoError(t, err)
		io.WriteString(wc, test.decoded)
		wc.Close()
		if s := b.String(); s != test.encoded {
			t.Errorf("Expected encoded text to be %q but got %q", test.encoded, s)
		}
	}
}

var lineWrapperTests = []struct {
	name    string
	in, out string
}{
	{
		name: "empty",
		in:   "",
		out:  "",
	},
	{
		name: "oneLine",
		in:   "ab",
		out:  "ab",
	},
	{
		name: "oneLineMax",
		in:   "abc",
		out:  "abc",
	},
	{
		name: "twoLines",
		in:   "abcde",
		out:  "abc\r\nde",
	},
	{
		name: "twoLinesMax",
		in:   "abcdef",
		out:  "abc\r\ndef",
	},
	{
		name: "threeLines",
		in:   "abcdefhi",
		out:  "abc\r\ndef\r\nhi",
	},
	{
		name: "wrappedMiss",
		in:   "abcd\nef",
		out:  "abc\r\nd\r\nef",
	},
	{
		name: "wrappedLF",
		in:   "abc\ndef\nhij",
		out:  "abc\r\ndef\r\nhij",
	},
	{
		name: "wrappedCRLF",
		in:   "abc\r\ndef\r\nhij",
		out:  "abc\r\ndef\r\nhij",
	},
	{
		name: "trailingCRLF",
		in:   "a\r\n",
		out:  "a\r\n",
	},
	{
		name: "cr",
		in:   "\r\r\r\r\r",
		out:  "\r\r\r\r\n\r",
	},
}

func TestLineWrapper(t *testing.T) {
	for _, tc := range lineWrapperTests {
		t.Run(tc.name, func(t *testing.T) {
			var sb strings.Builder
			w := &lineWrapper{w: &sb, maxLineLen: 3}
			if _, err := io.WriteString(w, tc.in); err != nil {
				t.Fatalf("WriteString() = %v", err)
			}
			if s := sb.String(); s != tc.out {
				t.Errorf("got %q, want %q", s, tc.out)
			}
		})

		t.Run(tc.name+"/bytePerByte", func(t *testing.T) {
			var sb strings.Builder
			w := &lineWrapper{w: &sb, maxLineLen: 3}
			if err := writeStringBytePerByte(w, tc.in); err != nil {
				t.Fatalf("writeStringBytePerByte() = %v", err)
			}
			if s := sb.String(); s != tc.out {
				t.Errorf("got %q, want %q", s, tc.out)
			}
		})
	}
}

func writeStringBytePerByte(w io.Writer, s string) error {
	for i := 0; i < len(s); i++ {
		if _, err := w.Write([]byte{s[i]}); err != nil {
			return err
		}
	}
	return nil
}
