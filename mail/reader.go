package mail

import (
	"container/list"
	"io"
	"mime"
	"net/textproto"
	"strings"

	"github.com/emersion/go-messages"
)

// A PartHeader is a mail part header. It contains convenience functions to get
// and set header fields.
type PartHeader interface {
	// Add adds the key, value pair to the header.
	Add(key, value string)
	// Del deletes the values associated with key.
	Del(key string)
	// Get gets the first value associated with the given key. If there are no
	// values associated with the key, Get returns "".
	Get(key string) string
	// Set sets the header entries associated with key to the single element
	// value. It replaces any existing values associated with key.
	Set(key, value string)
}

// A Part is either a mail text or an attachment. Header is either a TextHeader
// or an AttachmentHeader.
type Part struct {
	Header PartHeader
	Body   io.Reader
}

// A Reader reads mail parts.
type Reader struct {
	Header Header

	e       *messages.Entity
	readers *list.List
}

// NewReader creates a new mail reader.
func NewReader(e *messages.Entity) *Reader {
	mr := e.MultipartReader()
	if mr == nil {
		// Artificially create a multipart entity
		h := make(textproto.MIMEHeader)
		h.Set("Content-Type", "multipart/mixed")
		mr = messages.NewMultipart(h, []*messages.Entity{e}).MultipartReader()
	}

	l := list.New()
	l.PushBack(mr)

	return &Reader{Header{e.Header}, e, l}
}

// CreateReader reads a mail header from r and returns a new mail reader.
func CreateReader(r io.Reader) (*Reader, error) {
	e, err := messages.Read(r)
	if err != nil {
		return nil, err
	}

	return NewReader(e), nil
}

// NextPart returns the next mail part. If there is no more part, io.EOF is
// returned as error.
func (r *Reader) NextPart() (*Part, error) {
	for r.readers.Len() > 0 {
		e := r.readers.Back()
		mr := e.Value.(messages.MultipartReader)

		p, err := mr.NextPart()
		if err == io.EOF {
			// This whole multipart entity has been read, continue with the next one
			r.readers.Remove(e)
			continue
		} else if err != nil {
			return nil, err
		}

		if pmr := p.MultipartReader(); pmr != nil {
			// This is a multipart part, read it
			r.readers.PushBack(pmr)
		} else {
			// This is a non-multipart part, return a mail part
			mp := &Part{Body: p.Body}
			disp, _, _ := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
			if strings.HasPrefix(p.Header.Get("Content-Type"), "text/") && disp != "attachment" {
				mp.Header = TextHeader{p.Header}
			} else {
				mp.Header = AttachmentHeader{p.Header}
			}
			return mp, nil
		}
	}

	return nil, io.EOF
}

// Close finishes the reader.
func (r *Reader) Close() error {
	for r.readers.Len() > 0 {
		e := r.readers.Back()
		mr := e.Value.(messages.MultipartReader)

		if err := mr.Close(); err != nil {
			return err
		}

		r.readers.Remove(e)
	}

	return nil
}
