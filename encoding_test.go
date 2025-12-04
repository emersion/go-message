package message

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"mime/quotedprintable"
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

func TestLenientQPReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid_hex",
			input:    "=62",
			expected: "b",
		},
		{
			name:     "valid_soft_break",
			input:    "foo=\r\nbar",
			expected: "foobar",
		},
		{
			name:     "malformed_soft_break_spaces",
			input:    "foo=  \r\nbar", // Spaces after = are stripped
			expected: "foobar",
		},
		{
			name:     "malformed_soft_break_tabs",
			input:    "foo=\t\t\r\nbar",
			expected: "foobar",
		},
		{
			name:  "spaces_before_soft_break",
			input: "foo  =\r\nbar", // Leading spaces before soft break should be preserved logic?
			// Wait, my logic preserves spaces if they are before '=' and '=' is start of soft break.
			// input: "foo  =\r\n"
			// ' ' -> peek -> ' ' -> peek -> '='. '=' is soft break. Encode spaces.
			// encoded: "foo=20=20=\r\n"
			// decoded: "foo  " + "bar"
			expected: "foo  bar",
		},
		{
			name:     "malformed_literal_space",
			input:    "equation = 5", // = followed by space but no newline
			expected: "equation = 5", // = processed as =3D
		},
		{
			name:     "malformed_literal_eof",
			input:    "end=", // = at EOF
			expected: "end=",
		},
		{
			name:  "trailing_whitespace_eol",
			input: "end  \r\n", // Trailing spaces at EOL
			// My logic: If ' ' peek finds \r or \n -> do NOT encode (let Go strip).
			// So decoded "end\r\n" (stripped).
			expected: "end\r\n", // Go QP decoder usually strips this?
			// quotedprintable.NewReader returns decoded content.
			// If input "end  \r\n". Decoded is "end\r\n" or "end  \r\n"?
			// Go spec: "Trailing whitespace is removed from each line."
			// So "end\r\n".
		},
		{
			name:     "interleaved_valid_and_invalid",
			input:    "foo=  \r\n=62ar", // malformed soft break, then valid hex
			expected: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			// Manually construct lenientQPReader as it's private
			lqp := &lenientQPReader{wrapped: bufio.NewReader(r)}
			// Wrap with standard QP decoder
			dec := quotedprintable.NewReader(lqp)

			gotBytes, err := io.ReadAll(dec)
			if err != nil {
				t.Fatalf("ReadAll failed: %v", err)
			}
			got := string(gotBytes)

			if got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
}
