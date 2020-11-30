// Package charset provides functions to decode and encode charsets.
//
// It imports all supported charsets, which adds about 1MiB to binaries size.
// Importing the package automatically sets message.CharsetReader.
package charset

import (
	"fmt"
	"io"
	"strings"

	"github.com/emersion/go-message"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/ianaindex"
)

// Quirks table for charsets not handled by ianaindex
//
// A nil entry disables the charset.
//
// For aliases, see
// https://www.iana.org/assignments/character-sets/character-sets.xhtml
var charsets = map[string]encoding.Encoding{
	"ansi_x3.110-1983": charmap.ISO8859_1, // see RFC 1345 page 62, mostly superset of ISO 8859-1
}

func init() {
	message.CharsetReader = Reader
}

// Reader returns an io.Reader that converts the provided charset to UTF-8.
func Reader(charset string, input io.Reader) (io.Reader, error) {
	var err error
	enc, ok := charsets[strings.ToLower(charset)]
	if ok && enc == nil {
		return nil, fmt.Errorf("charset %q: charset is disabled", charset)
	} else if !ok {
		enc, err = ianaindex.MIME.Encoding(charset)
	}
	if enc == nil {
		enc, err = ianaindex.MIME.Encoding("cs" + charset)
	}
	if enc == nil {
		enc, err = htmlindex.Get(charset)
	}
	if err != nil {
		return nil, fmt.Errorf("charset %q: %v", charset, err)
	}
	// See https://github.com/golang/go/issues/19421
	if enc == nil {
		return nil, fmt.Errorf("charset %q: unsupported charset", charset)
	}
	return enc.NewDecoder().Reader(input), nil
}

// RegisterEncoding registers an encoding. This is intended to be called from
// the init function in packages that want to support additional charsets.
func RegisterEncoding(name string, enc encoding.Encoding) {
	charsets[name] = enc
}
