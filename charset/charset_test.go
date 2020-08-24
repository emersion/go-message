package charset

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

var testCharsets = []struct {
	charset string
	encoded []byte
	decoded string
}{
	{
		charset: "us-ascii",
		encoded: []byte("yuudachi"),
		decoded: "yuudachi",
	},
	{
		charset: "utf-8",
		encoded: []byte("café"),
		decoded: "café",
	},
	{
		charset: "utf8",
		encoded: []byte("café"),
		decoded: "café",
	},
	{
		charset: "windows-1250",
		encoded: []byte{0x8c, 0x8d, 0x8f, 0x9c, 0x9d, 0x9f, 0xbc, 0xbe},
		decoded: "ŚŤŹśťźĽľ",
	},
	{
		charset: "windows-1252",
		encoded: []byte{0x63, 0x61, 0x66, 0xE9, 0x20, 0x80},
		decoded: "café €",
	},
	{
		charset: "iso-8859-1",
		encoded: []byte{0x63, 0x61, 0x66, 0xE9},
		decoded: "café",
	},
	{
		charset: "idontexist",
		encoded: []byte{42},
	},
	{
		charset: "gb2312",
		encoded: []byte{178, 226, 202, 212},
		decoded: "测试",
	},
	{
		charset: "iso8859-2",
		encoded: []byte{0x63, 0x61, 0x66, 0xE9, 0x20, 0xfb},
		decoded: "café ű",
	},
}

func TestCharsetReader(t *testing.T) {
	for _, test := range testCharsets {
		r, err := Reader(test.charset, bytes.NewReader(test.encoded))
		if test.decoded == "" {
			if err == nil {
				t.Errorf("Expected an error when creating reader for charset %q", test.charset)
			}
		}
		if test.decoded != "" {
			if err != nil {
				t.Errorf("Expected no error when creating reader for charset %q, but got: %v", test.charset, err)
			} else if b, err := ioutil.ReadAll(r); err != nil {
				t.Errorf("Expected no error when reading charset %q, but got: %v", test.charset, err)
			} else if s := string(b); s != test.decoded {
				t.Errorf("Expected decoded text to be %q but got %q", test.decoded, s)
			}
		}
	}
}

func TestDisabledCharsetReader(t *testing.T) {
	_, err := Reader("hz-gb-2312", strings.NewReader("Some dummy text"))
	if err == nil {
		t.Errorf("%v encoding is disabled and should give an error", "hz-gb-2312")
		return
	}
	if !strings.HasSuffix(err.Error(), "charset is disabled") {
		t.Errorf("expected error to end in 'unsupported charset', got %v",
			err.Error())
	}
}
