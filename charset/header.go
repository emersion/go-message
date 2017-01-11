package charset

import (
	"mime"
)

var wordDecoder = &mime.WordDecoder{CharsetReader: Reader}

// DecodeHeader decodes an internationalized header field. If it fails, it
// returns the input string and the error.
func DecodeHeader(s string) (string, error) {
	dec, err := wordDecoder.DecodeHeader(s)
	if err != nil {
		return s, err
	}
	return dec, nil
}

// EncodeHeader encodes an internationalized header field.
func EncodeHeader(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}
