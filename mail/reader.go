package mail

import (
	"container/list"
	"io"
	"mime"
	"net/textproto"
	"strings"

	"github.com/emersion/go-messages"
)

// A Text represents a message's text.
type Text struct {
	Header textproto.MIMEHeader
	Body   io.Reader
}

// IsHTML checks if Body is formatted using HTML.
func (t *Text) IsHTML() bool {
	mediaType, _, _ := mime.ParseMediaType(t.Header.Get("Content-Type"))
	return mediaType == "text/html"
}

// A Reader reads mail parts.
type Reader struct {
	Header  Header

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

	return &Reader{Header(e.Header), e, l}
}

// CreateReader reads a mail header from r and returns a new mail reader.
func CreateReader(r io.Reader) (*Reader, error) {
	e, err := messages.Read(r)
	if err != nil {
		return nil, err
	}

	return NewReader(e), nil
}

// NextPart returns the next mail part, which can be either a Text or an
// Attachment. If there is no more part, io.EOF is returned as error.
func (r *Reader) NextPart() (interface{}, error) {
	for r.readers.Len() > 0 {
		e := r.readers.Back()
		mr := e.Value.(messages.MultipartReader)

		p, err := mr.NextPart()
		if err == io.EOF {
			continue
		} else if err != nil {
			return nil, err
		}

		if pmr := p.MultipartReader(); pmr != nil {
			r.readers.PushBack(pmr)
		} else {
			disp, _, _ := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
			if strings.HasPrefix(p.Header.Get("Content-Type"), "text/") && disp != "attachment" {
				return &Text{p.Header, p.Body}, nil
			} else {
				return &Attachment{p.Header, p.Body}, nil
			}
		}

		r.readers.Remove(e)
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
