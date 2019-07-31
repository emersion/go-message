package message

import (
	"fmt"
	"io"
	"mime"
	"strings"
)

type unknownCharsetError struct {
	error
}

// IsUnknownCharset returns a boolean indicating whether the error is known to
// report that the charset advertised by the entity is unknown.
func IsUnknownCharset(err error) bool {
	_, ok := err.(unknownCharsetError)
	return ok
}

// CharsetReader, if non-nil, defines a function to generate charset-conversion
// readers, converting from the provided charset into UTF-8. Charsets are always
// lower-case. utf-8 and us-ascii charsets are handled by default. One of the
// the CharsetReader's result values must be non-nil.
//
// Importing github.com/emersion/go-message/charset will set CharsetReader to
// a function that handles most common charsets. Alternatively, CharsetReader
// can be set to e.g. golang.org/x/net/html/charset.NewReaderLabel.
var CharsetReader func(charset string, input io.Reader) (io.Reader, error)

// charsetReader calls CharsetReader if non-nil.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)
	if charset == "utf-8" || charset == "us-ascii" {
		return input, nil
	}
	if CharsetReader != nil {
		return CharsetReader(charset, input)
	}
	return input, fmt.Errorf("message: unhandled charset %q", charset)
}

// decodeHeader decodes an internationalized header field. If it fails, it
// returns the input string and the error.
func decodeHeader(s string) (string, error) {
	wordDecoder := mime.WordDecoder{CharsetReader: charsetReader}
	dec, err := wordDecoder.DecodeHeader(s)
	if err != nil {
		return s, err
	}
	return dec, nil
}

func encodeHeader(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}
