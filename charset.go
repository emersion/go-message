package message

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

// CharsetReader returns an io.Reader that converts the provided charset to
// UTF-8.
func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)
	if charset == "utf-8" || charset == "us-ascii" {
		return input, nil
	}
	if enc, ok := charsets[charset]; ok {
		return enc.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("messages: unhandled charset %q", charset)
}
