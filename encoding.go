package message

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"
	"bytes"

	"log"
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

func (nopCloser) Close() error {
	return nil
}

func encodingWriter(enc string, w io.Writer) (io.WriteCloser, error) {
	var wc io.WriteCloser
	switch strings.ToLower(enc) {
	case "quoted-printable":
		wc = quotedprintable.NewWriter(w)
	case "base64":
		wc = base64.NewEncoder(base64.StdEncoding, &wrapWriter{w: w, max: 76})
	case "7bit", "8bit":
		wc = nopCloser{&wrapWriter{w: w, max: 1000}}
	case "binary", "":
		wc = nopCloser{w}
	default:
		return nil, fmt.Errorf("unhandled encoding %q", enc)
	}
	return wc, nil
}

// wrapWriter is an io.Writer that wraps long text lines to a specified length.
type wrapWriter struct {
	w   io.Writer
	max int // including CRLF

	n  int  // current line length
	cr bool // previous byte was \r
	crlf bool // previous bytes were \r\n
}

func (w *wrapWriter) Write(b []byte) (int, error) {
	N := 0
	for len(b) > 0 {
		i := bytes.IndexByte(b, '\n')

		to := i + 1
		if i < 0 || to > w.max - w.n + 2 {
			to = w.max - w.n + 2
			if to > len(b) {
				to = len(b)
			} else if b[to-2] == '\n' {
				to--
			} else if b[to-1] != '\n' || b[to-2] != '\r' {
				to -= 2
			}
		}

		n, err := w.writeChunk(b[:to])
		N += n
		if err != nil {
			return N, err
		}

		b = b[to:]
	}

	return N, nil
}

func (w *wrapWriter) writeChunk(b []byte) (int, error) {
	lf := bytes.HasSuffix(b, []byte{'\n'})
	crlf := (w.cr && bytes.HasPrefix(b, []byte{'\n'})) || bytes.HasSuffix(b, []byte{'\r', '\n'})

	log.Printf("%q crlf=%v n=%v", string(b), w.crlf, w.n)
	if !w.crlf && w.n >= w.max {
		// If the previous line didn't end with a CRLF, write one
		if _, err := w.w.Write([]byte{'\r', '\n'}); err != nil {
			return 0, err
		}
		w.n = 0
	}

	var (
		n   int
		err error
	)
	if lf && !crlf {
		// Need to convert lone LF to CRLF
		n, err = w.w.Write(b[:len(b)-1])
		if err != nil {
			return n, err
		}
		if _, err := w.w.Write([]byte{'\r', '\n'}); err != nil {
			return n, err
		}
		n++
		w.crlf = true
	} else {
		n, err = w.w.Write(b)
		if err != nil {
			return n, err
		}
		w.crlf = crlf
	}

	w.cr = bytes.HasSuffix(b, []byte{'\r'})
	if lf || crlf {
		w.n = 0
	} else {
		w.n += n
	}
	return n, nil
}
