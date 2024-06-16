package mail

import (
	"github.com/emersion/go-message"
)

// parseFilename parses the filename from the header.
func parseFilename(h message.Header) (string, error) {
	_, params, err := h.ContentDisposition()

	filename, ok := params["filename"]
	if !ok {
		// Using "name" in Content-Type is discouraged
		_, params, err = h.ContentType()
		filename = params["name"]
	}

	return filename, err
}

// An AttachmentHeader represents an attachment's header.
type AttachmentHeader struct {
	message.Header
}

// Filename parses the attachment's filename.
func (h *AttachmentHeader) Filename() (string, error) {
	return parseFilename(h.Header)
}

// SetFilename formats the attachment's filename.
func (h *AttachmentHeader) SetFilename(filename string) {
	dispParams := map[string]string{"filename": filename}
	h.SetContentDisposition("attachment", dispParams)
}

// An InlineAttachmentHeader represents an inlined attachment's header.
type InlineAttachmentHeader struct {
	message.Header
}

// Filename parses the attachment's filename.
func (h *InlineAttachmentHeader) Filename() (string, error) {
	return parseFilename(h.Header)
}

// SetFilename formats the attachment's filename.
func (h *InlineAttachmentHeader) SetFilename(filename string) {
	dispParams := map[string]string{"filename": filename}
	h.SetContentDisposition("inline", dispParams)
}
