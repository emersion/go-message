package internal

import (
	"bytes"
	"io/ioutil"
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
}

func TestCharsetReader(t *testing.T) {
	for _, test := range testCharsets {
		r, err := CharsetReader(test.charset, bytes.NewReader(test.encoded))
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
