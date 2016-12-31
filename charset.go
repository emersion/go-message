package messages

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var charsets = map[string]encoding.Encoding{
	"iso-8859-1":   charmap.ISO8859_1,
	"windows-1252": charmap.Windows1252,
}

// charsetReader returns an io.Reader that converts the provided charset to
// UTF-8.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	if strings.EqualFold("utf-8", charset) {
		return input, nil
	}
	if enc, ok := charsets[strings.ToLower(charset)]; ok {
		return enc.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("message: unhandled charset %q", charset)
}
