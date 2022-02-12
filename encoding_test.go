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

var wrapWriterTests = []struct{
	name   string
	writes []string
	out    string
}{
	{
		name:   "singleLine",
		writes: []string{"Hey"},
		out:    "Hey",
	},
	{
		name: "twoLines",
		writes: []string{"Hey\r\nYo"},
		out: "Hey\r\nYo",
	},
	{
		name: "finalCRLF",
		writes: []string{"Hey\r\n"},
		out: "Hey\r\n",
	},
	{
		name: "singleLineSplit",
		writes: []string{"He", "y"},
		out: "Hey",
	},
	{
		name: "twoLinesSplit",
		writes: []string{"He", "y", "\r\nY", "o"},
		out: "Hey\r\nYo",
	},
	{
		name: "longLine",
		writes: []string{"How are you today?"},
		out: "How are yo\r\nu today?",
	},
	{
		name: "longLineSplit",
		writes: []string{"How are ", "you today?"},
		out: "How are yo\r\nu today?",
	},
	{
		name: "lf",
		writes: []string{"Hey\nYo"},
		out: "Hey\r\nYo",
	},
	{
		name: "max",
		writes: []string{"Hey there!"},
		out: "Hey there!",
	},
	{
		name: "maxCRLF",
		writes: []string{"Hey there!\r\n"},
		out: "Hey there!\r\n",
	},
	{
		name: "maxLF",
		writes: []string{"Hey there!\n"},
		out: "Hey there!\r\n",
	},
	{
		name: "maxMinusOne",
		writes: []string{"Hey there"},
		out: "Hey there",
	},
	{
		name: "maxMinusOneCRLF",
		writes: []string{"Hey there\r\n"},
		out: "Hey there\r\n",
	},
	{
		name: "maxMinusOneLF",
		writes: []string{"Hey there\n"},
		out: "Hey there\r\n",
	},
	{
		name: "maxPlusOne",
		writes: []string{"Hey there!!"},
		out: "Hey there!\r\n!",
	},
	{
		name: "maxPlusTwo",
		writes: []string{"Hey there!!!"},
		out: "Hey there!\r\n!!",
	},
	{
		name: "maxSplit",
		writes: []string{"Hey ", "there!", "\r", "\n"},
		out: "Hey there!\r\n",
	},
}

func TestWrapWriter(t *testing.T) {
	for _, test := range wrapWriterTests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			ww := wrapWriter{
				w:   &buf,
				max: 10,
			}

			for _, s := range test.writes {
				if _, err := ww.Write([]byte(s)); err != nil {
					t.Fatalf("wrapWriter.Write() = %v", err)
				}
			}

			if buf.String() != test.out {
				t.Errorf("got %q, want %q", buf.String(), test.out)
			}
		})
	}
}
