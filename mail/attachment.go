package mail

import (
	"mime"
	"net/textproto"

	"github.com/emersion/go-message"
)

// An AttachmentHeader represents an attachment's header.
type AttachmentHeader struct {
	textproto.MIMEHeader
}

// NewAttachmentHeader creates a new AttachmentHeader.
func NewAttachmentHeader() AttachmentHeader {
	h := AttachmentHeader{make(textproto.MIMEHeader)}
	h.Set("Content-Disposition", "attachment")
	h.Set("Content-Transfer-Encoding", "base64")
	return h
}

// Filename parses the attachment's filename.
func (h AttachmentHeader) Filename() (string, error) {
	_, params, err := mime.ParseMediaType(h.Get("Content-Disposition"))

	filename, ok := params["filename"]
	if !ok {
		// Using "name" in Content-Type is discouraged
		_, params, err = mime.ParseMediaType(h.Get("Content-Type"))
		filename = params["name"]
	}

	if err != nil {
		return filename, err
	}

	dec := &mime.WordDecoder{CharsetReader: message.CharsetReader}
	decoded, err := dec.DecodeHeader(filename)
	if err == nil {
		filename = decoded
	}
	return filename, err
}

// SetFilename formats the attachment's filename.
func (h AttachmentHeader) SetFilename(filename string) {
	filename = mime.QEncoding.Encode("utf-8", filename)
	dispParams := map[string]string{"filename": filename}
	h.Set("Content-Disposition", mime.FormatMediaType("attachment", dispParams))
}
