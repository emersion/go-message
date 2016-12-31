package messages

import (
	"encoding/base64"
	"io"
	"mime/quotedprintable"
	"strings"

	"github.com/emersion/go-textwrapper"
)

func decode(enc string, r io.Reader) io.Reader {
	switch strings.ToLower(enc) {
	case "quoted-printable":
		r = quotedprintable.NewReader(r)
	case "base64":
		r = base64.NewDecoder(base64.StdEncoding, r)
	}
	return r
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func encode(enc string, w io.Writer) io.WriteCloser {
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
