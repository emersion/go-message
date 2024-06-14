package message

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func TestUtil(t *testing.T) {
	vietnameseText := "Hôm nay trời đẹp quá!"
	var encodedBuffer bytes.Buffer
	encoder := charmap.Windows1258.NewEncoder()
	writer := transform.NewWriter(&encodedBuffer, encoder)
	if _, err := writer.Write([]byte(vietnameseText)); err != nil {
		fmt.Println("Error encoding:", err)
		return
	}
	if err := writer.Close(); err != nil {
		fmt.Println("Error closing writer:", err)
		return
	}
	encodedBytes := encodedBuffer.Bytes()
	fmt.Println("Encoded bytes:", encodedBytes)
	fmt.Println("Encoded string:", fmt.Sprintf("% x", encodedBytes))
}

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
		enc:     "utf8",
		encoded: "café",
		decoded: "café",
	},
	{
		enc:     "ascii",
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
	{
		enc:     "windows-1252",
		encoded: "Hell\xE9, world!",
		decoded: "Hellé, world!",
	},
	{
		enc:     "cp1252",
		encoded: "Hell\xE9, world!",
		decoded: "Hellé, world!",
	},
	{
		enc:     "iso-2022-jp",
		encoded: "\x1b$B$3$s$K$A$O\x1b(B",
		decoded: "こんにちは", // "hello" in Japanese
	},
	{
		enc:     "iso-8859-14",
		encoded: "\xA1ello, world!",
		decoded: "Ḃello, world!",
	},
	{
		enc:     "ansi_x3.4-1968",
		encoded: "Hello, world!", // This encoding is the same as ascii, but with some revisions
		decoded: "Hello, world!",
	},
	{
		enc:     "iso-8859-2",
		encoded: "Hello, World! \xbf\xf3\xbb\xea",
		decoded: "Hello, World! żóťę",
	},
	{
		enc:     "windows-1251",
		encoded: "Hello, World! \xbf\xf3\xbb\xea",
		decoded: "Hello, World! їу»к",
	},
	{
		enc:     "windows-1256",
		encoded: "Hello, World! \xbf\xf3\xbb\xea",
		decoded: "Hello, World! ؟َ»ê",
	},
	{
		enc:     "koi8-u",
		encoded: "Hello, World! \xbf\xf3\xbb\xea",
		decoded: "Hello, World! ©С╩Й",
	},
	{
		enc:     "ks_c_5601-1987",
		encoded: "\xbeȳ\xe7\xc7ϼ\xbc\xbf\xe4",
		decoded: "안녕하세요", // "hello" in Korean
	},
	{
		enc:     "gbk",
		encoded: "\xc4\xe3\xba\xc3",
		decoded: "你好", // "hello" in Chinese
	},
	{
		enc:     "iso-8859-6",
		encoded: "Hello, World! \xbf\xea",
		decoded: "Hello, World! ؟ي", // Arabic question mark
	},
	{
		enc:     "windows-1257",
		encoded: "Hello, World! \xfe\xe0\xe8\xe6\xeb\xe1\xf0\xf8\xfb",
		decoded: "Hello, World! žąčęėįšųū", // Lithuanian
	},
	{
		enc:     "windows-1250",
		encoded: "\xd8\xedkej, \x9ee t\xec l\xe1ska k \x9eivotu nedovol\xed lh\xe1t.",
		decoded: "Říkej, že tě láska k životu nedovolí lhát.", // Czech
	},
	{
		enc:     "gb2312",
		encoded: "\xc4\xe3\xba\xc3",
		decoded: "你好", // "hello" in Chinese
	},
	{
		enc:     "iso-8859-8-i",
		encoded: "\xf9\xec\xe5\xed, \xee\xe4 \xf9\xec\xe5\xee\xea?",
		decoded: "שלום, מה שלומך?", // Hebrew, "hello, how are you?"
	},
	{
		enc:     "windows-1258",
		encoded: "Ch\xe0o",
		decoded: "Chào", // Vietnamese, "hello"
	},
	{
		enc:     "big5",
		encoded: "\xd2\xf1\xc2k\xbf\xa4",
		decoded: "秭歸縣", // Chinese, "Zigui County"
	},
	{
		enc:     "windows-1255",
		encoded: "\xf9\xec\xe5\xed, \xee\xe4 \xf9\xec\xe5\xee\xea?",
		decoded: "שלום, מה שלומך?", // Hebrew, "hello, how are you?"
	},
	{
		enc:     "windows-1253",
		encoded: "\xca\xe1\xeb\xe7\xec\xdd\xf1\xe1 \xea\xfc\xf3\xec\xe5!",
		decoded: "Καλημέρα κόσμε!", // Greek, "Good morning world!"
	},
	{
		enc:     "iso-8859-9",
		encoded: "Merhaba d\xfcnya!",
		decoded: "Merhaba dünya!", // Turkish, "hello world!"
	},
	{
		enc:     "windows-1254",
		encoded: "Merhaba d\xfcnya!",
		decoded: "Merhaba dünya!", // Turkish, "hello world!"
	},
	{
		enc:     "shift-jis",
		encoded: "\x82\xb1\x82\xf1\x82ɂ\xbf\x82\xcd",
		decoded: "こんにちは", // "hello" in Japanese
	},
	{
		enc:     "utf-16le",
		encoded: "H\x00e\x00l\x00l\x00o\x00,\x00 \x00w\x00o\x00r\x00l\x00d\x00!\x00 \x00=\xd8\x00\xde",
		decoded: "Hello, world! 😀",
	},
	{
		enc:     "iso-8859-5",
		encoded: "\xbf\xe0\xd8\xd2\xd5\xe2, \xdc\xd8\xe0!",
		decoded: "Привет, мир!", // Russian, "hello world!"
	},
	{
		enc:     "iso-8859-7",
		encoded: "\xca\xe1\xeb\xe7\xec\xdd\xf1\xe1, \xea\xfc\xf3\xec\xe5!",
		decoded: "Καλημέρα, κόσμε!", // Greek, "good morning, world!"
	},
	{
		enc:     "iso_8859-1",
		encoded: "caf\xe9",
		decoded: "café",
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
