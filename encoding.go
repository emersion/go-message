package message

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"

	"github.com/emersion/go-textwrapper"
)

func encodingReader(enc string, r io.Reader) (io.Reader, error) {
	var dec io.Reader
	switch strings.ToLower(enc) {
	case "quoted-printable":
		dec = quotedprintable.NewReader(r)
	case "base64":
		dec = base64.NewDecoder(base64.StdEncoding, r)
	case "7bit", "8bit", "binary", "":
		dec = r
	default:
		return nil, fmt.Errorf("unhandled encoding %q", enc)
	}
	return dec, nil
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func encodingWriter(enc string, w io.Writer) io.WriteCloser {
	var wc io.WriteCloser
	switch strings.ToLower(enc) {
	case "quoted-printable":
		wc = quotedprintable.NewWriter(w)
	case "base64":
		wc = base64.NewEncoder(base64.StdEncoding, textwrapper.NewRFC822(w))
	case "7bit", "8bit":
		wc = nopCloser{textwrapper.New(w, "\r\n", 1000)}
	default: // "binary"
		wc = nopCloser{w}
	}
	return wc
}
