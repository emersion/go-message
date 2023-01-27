package message

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	mimeqp "mime/quotedprintable"
	"strings"
	"sync"

	"github.com/emersion/go-message/quotedprintable"
	"github.com/emersion/go-textwrapper"
)

var (
	decoders sync.Map // map[string]decoderProviderFn
)

type UnknownEncodingError struct {
	e error
}

func (u UnknownEncodingError) Unwrap() error { return u.e }

func (u UnknownEncodingError) Error() string {
	return "encoding error: " + u.e.Error()
}

// IsUnknownEncoding returns a boolean indicating whether the error is known to
// report that the encoding advertised by the entity is unknown.
func IsUnknownEncoding(err error) bool {
	return errors.As(err, new(UnknownEncodingError))
}

// DecoderProviderFn should return an implementation of io.Reader capable of
// decoding the transport encoding it was registered with, or nil to use the
// module defaults.
type DecoderProviderFn func(r io.Reader) io.Reader

// RegisterTransportDecoder allows custom decoders for a specified transport
// encoding, which can override the module defaults. If there is existing
// custom decoder for a transportEncoding, it is replaced.
func RegisterTransportDecoder(transportEncoding string, f DecoderProviderFn) {
	if transportEncoding == "" || f == nil {
		return
	}
	decoders.Store(strings.ToLower(transportEncoding), f)
}

func encodingReader(enc string, r io.Reader) (io.Reader, error) {
	var dec io.Reader
	enc = strings.ToLower(enc)

	if f, ok := decoders.Load(enc); ok {
		dec = f.(DecoderProviderFn)(r)
		if dec != nil {
			return dec, nil
		}
	}

	switch enc {
	case "quoted-printable":
		dec = quotedprintable.NewReader(r)
	case "base64":
		wrapped := &whitespaceReplacingReader{wrapped: r}
		dec = base64.NewDecoder(base64.StdEncoding, wrapped)
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

func (nopCloser) Close() error {
	return nil
}

func encodingWriter(enc string, w io.Writer) (io.WriteCloser, error) {
	var wc io.WriteCloser
	switch strings.ToLower(enc) {
	case "quoted-printable":
		wc = mimeqp.NewWriter(w)
	case "base64":
		wc = base64.NewEncoder(base64.StdEncoding, textwrapper.NewRFC822(w))
	case "7bit", "8bit":
		wc = nopCloser{textwrapper.New(w, "\r\n", 1000)}
	case "binary", "":
		wc = nopCloser{w}
	default:
		return nil, fmt.Errorf("unhandled encoding %q", enc)
	}
	return wc, nil
}

// whitespaceReplacingReader replaces space and tab characters with a LF so
// base64 bodies with a continuation indent can be decoded by the base64 decoder
// even though it is against the spec.
type whitespaceReplacingReader struct {
	wrapped io.Reader
}

func (r *whitespaceReplacingReader) Read(p []byte) (int, error) {
	n, err := r.wrapped.Read(p)

	for i := 0; i < n; i++ {
		if p[i] == ' ' || p[i] == '\t' {
			p[i] = '\n'
		}
	}

	return n, err
}
