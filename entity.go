package message

import (
	"bufio"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/emersion/go-message/charset"
)

// An Entity is either a message or a one of the parts in the body of a
// multipart entity.
type Entity struct {
	Header Header    // The entity's header.
	Body   io.Reader // The decoded entity's body.

	mediaType   string
	mediaParams map[string]string
}

// NewEntity makes a new Entity with the provided header and body. The entity's
// encoding and charset are automatically decoded to UTF-8.
func NewEntity(header Header, body io.Reader) *Entity {
	body = encodingReader(header.Get("Content-Transfer-Encoding"), body)

	mediaType, mediaParams, _ := header.ContentType()
	if ch, ok := mediaParams["charset"]; ok {
		if converted, err := charset.Reader(ch, body); err == nil {
			body = converted
		}
	}

	return &Entity{
		Header:      header,
		Body:        body,
		mediaType:   mediaType,
		mediaParams: mediaParams,
	}
}

// NewMultipart makes a new multipart Entity with the provided header and parts.
// The Content-Type header must begin with "multipart/".
func NewMultipart(header Header, parts []*Entity) *Entity {
	r := &multipartBody{
		header: header,
		parts:  parts,
	}

	return NewEntity(header, r)
}

// Read reads a message from r. The message's encoding and charset are
// automatically decoded to UTF-8.
func Read(r io.Reader) (*Entity, error) {
	br := bufio.NewReader(r)
	h, err := textproto.NewReader(br).ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	return NewEntity(Header(h), br), nil
}

// MultipartReader returns a MultipartReader that reads parts from this entity's
// body. If this entity is not multipart, it returns nil.
func (e *Entity) MultipartReader() MultipartReader {
	if !strings.HasPrefix(e.mediaType, "multipart/") {
		return nil
	}
	if mb, ok := e.Body.(*multipartBody); ok {
		return mb
	}
	return &multipartReader{multipart.NewReader(e.Body, e.mediaParams["boundary"])}
}

// writeBodyTo writes this entity's body to w (without the header).
func (e *Entity) writeBodyTo(w *Writer) error {
	var err error
	if mb, ok := e.Body.(*multipartBody); ok {
		err = mb.writeBodyTo(w)
	} else {
		_, err = io.Copy(w, e.Body)
	}
	return err
}

// WriteTo writes this entity's header and body to w.
func (e *Entity) WriteTo(w io.Writer) error {
	ew, err := CreateWriter(w, e.Header)
	if err != nil {
		return err
	}
	defer ew.Close()

	return e.writeBodyTo(ew)
}
