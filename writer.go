package message

import (
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"sort"
	"strings"
)

// From https://golang.org/src/mime/multipart/writer.go?s=2140:2215#L76
func writeHeader(w io.Writer, header Header) error {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			if _, err := io.WriteString(w, formatHeaderField(k, v)); err != nil {
				return err
			}
		}
	}
	_, err := io.WriteString(w, "\r\n")
	return err
}

// A Writer formats entities.
type Writer struct {
	w  io.Writer
	c  io.Closer
	mw *multipart.Writer
}

// newWriter creates a new Writer writing to w with the provided header. Nothing
// is written to w when it is called. header is modified in-place.
func newWriter(w io.Writer, header Header) *Writer {
	ww := &Writer{w: w}

	mediaType, mediaParams, _ := header.ContentType()
	if strings.HasPrefix(mediaType, "multipart/") {
		ww.mw = multipart.NewWriter(ww.w)

		// Do not set ww's io.Closer for now: if this is a multipart entity but
		// CreatePart is not used (only Write is used), then the final boundary
		// is expected to be written by the user too. In this case, ww.Close
		// shouldn't write the final boundary.

		if mediaParams["boundary"] != "" {
			ww.mw.SetBoundary(mediaParams["boundary"])
		} else {
			mediaParams["boundary"] = ww.mw.Boundary()
			header.SetContentType(mediaType, mediaParams)
		}

		header.Del("Content-Transfer-Encoding")
	} else {
		wc := encodingWriter(header.Get("Content-Transfer-Encoding"), ww.w)
		ww.w = wc
		ww.c = wc
	}

	return ww
}

// CreateWriter creates a new Writer writing to w. If header contains an
// encoding, data written to the Writer will automatically be encoded with it.
func CreateWriter(w io.Writer, header Header) (*Writer, error) {
	ww := newWriter(w, header)
	if err := writeHeader(w, header); err != nil {
		return nil, err
	}
	return ww, nil
}

// Write implements io.Writer.
func (w *Writer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

// Close implements io.Closer.
func (w *Writer) Close() error {
	if w.c != nil {
		return w.c.Close()
	}
	return nil
}

// CreatePart returns a Writer to a new part in this multipart entity. If this
// entity is not multipart, it fails. The body of the part should be written to
// the returned io.WriteCloser.
func (w *Writer) CreatePart(header Header) (*Writer, error) {
	if w.mw == nil {
		return nil, errors.New("cannot create a part in a non-multipart message")
	}

	if w.c == nil {
		// We know that the user calls CreatePart so Close should write the final
		// boundary
		w.c = w.mw
	}

	// cw -> ww -> pw -> w.mw -> w.w

	ww := &struct{ io.Writer }{nil}
	cw := newWriter(ww, header)
	pw, err := w.mw.CreatePart(textproto.MIMEHeader(header))
	if err != nil {
		return nil, err
	}

	ww.Writer = pw
	return cw, nil
}
