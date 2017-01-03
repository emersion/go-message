package mail

import (
	"io"
	"mime"
	"net/textproto"
)

// A TextHeader represents a message text header.
type TextHeader struct {
	textproto.MIMEHeader
}

// NewTextHeader creates a new message text header.
func NewTextHeader() TextHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", "inline")
	return TextHeader{h}
}

// IsHTML checks if this text is formatted using HTML.
func (h TextHeader) IsHTML() bool {
	mediaType, _, _ := mime.ParseMediaType(h.Get("Content-Type"))
	return mediaType == "text/html"
}

// A Text represents a message's text.
type Text struct {
	Header TextHeader
	Body   io.Reader
}
