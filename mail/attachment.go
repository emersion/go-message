package mail

import (
	"io"
	"mime"
	"net/textproto"

	"github.com/emersion/go-messages"
)

// An Attachment represents a mail attachment.
type Attachment struct {
	Header textproto.MIMEHeader
	Body   io.Reader
}

// Filename parses the attachment filename.
func (a *Attachment) Filename() (string, error) {
	_, params, err := mime.ParseMediaType(a.Header.Get("Content-Disposition"))

	filename, ok := params["filename"]
	if !ok {
		// Using "name" in Content-Type is discouraged
		_, params, err = mime.ParseMediaType(a.Header.Get("Content-Type"))
		filename = params["name"]
	}

	if err != nil {
		return filename, err
	}

	dec := &mime.WordDecoder{CharsetReader: messages.CharsetReader}
	decoded, err := dec.DecodeHeader(filename)
	if err == nil {
		filename = decoded
	}
	return filename, err
}

// SetFilename formats the attachment filename.
func (a *Attachment) SetFilename(filename string) {
	filename = mime.QEncoding.Encode("utf-8", filename)
	dispParams := map[string]string{"filename": filename}
	a.Header.Set("Content-Disposition", mime.FormatMediaType("attachment", dispParams))
}
