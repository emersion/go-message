package mail

import (
	"io"
	"strings"

	"github.com/emersion/go-message"
)

func initInlineContentTransferEncoding(h *message.Header) {
	if !h.Has("Content-Transfer-Encoding") {
		t, _, _ := h.ContentType()
		if strings.HasPrefix(t, "text/") {
			h.Set("Content-Transfer-Encoding", "quoted-printable")
		} else {
			h.Set("Content-Transfer-Encoding", "base64")
		}
	}
}

func initInlineHeader(h *InlineHeader) {
	h.Set("Content-Disposition", "inline")
	initInlineContentTransferEncoding(&h.Header)
}

func initAttachmentHeader(h *AttachmentHeader) {
	disp, _, _ := h.ContentDisposition()
	if disp != "attachment" {
		h.Set("Content-Disposition", "attachment")
	}
	if !h.Has("Content-Transfer-Encoding") {
		h.Set("Content-Transfer-Encoding", "base64")
	}
}

func initInlineAttachmentHeader(h *InlineAttachmentHeader) {
	disp, _, _ := h.ContentDisposition()
	if disp != "inline" {
		h.Set("Content-Disposition", "inline")
	}
	if !h.Has("Content-Transfer-Encoding") {
		h.Set("Content-Transfer-Encoding", "base64")
	}
}

// A Writer writes a mail message. A mail message contains one or more text
// parts and zero or more attachments.
type Writer struct {
	mw *message.Writer
}

// CreateWriter writes a mail header to w and creates a new Writer.
func CreateWriter(w io.Writer, header Header) (*Writer, error) {
	header = header.Copy() // don't modify the caller's view
	header.Set("Content-Type", "multipart/mixed")

	mw, err := message.CreateWriter(w, header.Header)
	if err != nil {
		return nil, err
	}

	return &Writer{mw}, nil
}

// CreateAlternativeWriter writes a mail header to w. The mail will contain an
// inline part, allowing to represent the same text in different formats.
// Attachments cannot be included.
func CreateAlternativeWriter(w io.Writer, header Header) (*AlternativeWriter, error) {
	header = header.Copy() // don't modify the caller's view
	header.Set("Content-Type", "multipart/alternative")

	mw, err := message.CreateWriter(w, header.Header)
	if err != nil {
		return nil, err
	}

	return &AlternativeWriter{mw}, nil
}

// CreateInlineWriter writes a mail header to w. The mail will contain an
// inline part, allowing to represent the same text in different formats.
// Attachments cannot be included.
// Deprecated: Use CreateAlternativeWriter and the returned AlternativeWriter instead.
func CreateInlineWriter(w io.Writer, header Header) (*InlineWriter, error) {
	header = header.Copy() // don't modify the caller's view
	header.Set("Content-Type", "multipart/alternative")

	mw, err := message.CreateWriter(w, header.Header)
	if err != nil {
		return nil, err
	}

	return &InlineWriter{mw}, nil
}

// CreateRelatedWriter writes a mail header to w. The mail will contain an
// inline part and any inline attachments. Non-inline attachments cannot be added.
func CreateRelatedWriter(w io.Writer, header Header) (*RelatedWriter, error) {
	header = header.Copy() // don't modify the caller's view
	header.Set("Content-Type", "multipart/related")

	mw, err := message.CreateWriter(w, header.Header)
	if err != nil {
		return nil, err
	}

	return &RelatedWriter{mw}, nil
}

// CreateSingleInlineWriter writes a mail header to w. The mail will contain a
// single inline part. The body of the part should be written to the returned
// io.WriteCloser. Only one single inline part should be written, use
// CreateWriter if you want multiple parts.
func CreateSingleInlineWriter(w io.Writer, header Header) (io.WriteCloser, error) {
	header = header.Copy() // don't modify the caller's view
	initInlineContentTransferEncoding(&header.Header)
	return message.CreateWriter(w, header.Header)
}

// ------------------------------------------------------------------------------------------------------------------ //

// CreateAlternative creates an AlternativeWriter. One or more parts representing the same
// text in different formats can be written to an AlternativeWriter.
func (w *Writer) CreateAlternative() (*AlternativeWriter, error) {
	var h message.Header
	h.Set("Content-Type", "multipart/alternative")

	mw, err := w.mw.CreatePart(h)
	if err != nil {
		return nil, err
	}
	return &AlternativeWriter{mw}, nil
}

// CreateInline creates a InlineWriter. One or more parts representing the same
// text in different formats can be written to a InlineWriter.
// Deprecated: Use CreateAlternative() and the AlternativeWriter instead.
func (w *Writer) CreateInline() (*InlineWriter, error) {
	var h message.Header
	h.Set("Content-Type", "multipart/alternative")

	mw, err := w.mw.CreatePart(h)
	if err != nil {
		return nil, err
	}
	return &InlineWriter{mw}, nil
}

// CreateRelated created a RelatedWriter. Inline attachments and one or more
// parts representing the same text in different format can be written to a
// RelatedWriter.
func (w *Writer) CreateRelated() (*RelatedWriter, error) {
	var h message.Header
	h.Set("Content-Type", "multipart/related")

	mw, err := w.mw.CreatePart(h)
	if err != nil {
		return nil, err
	}
	return &RelatedWriter{mw}, nil
}

// CreateSingleInline creates a new single text part with the provided header.
// The body of the part should be written to the returned io.WriteCloser. Only
// one single text part should be written, use CreateAlternative if you want
// multiple text parts.
func (w *Writer) CreateSingleInline(h InlineHeader) (io.WriteCloser, error) {
	h = InlineHeader{h.Header.Copy()} // don't modify the caller's view
	initInlineHeader(&h)
	return w.mw.CreatePart(h.Header)
}

// CreateAttachment creates a new attachment with the provided header. The body
// of the part should be written to the returned io.WriteCloser.
func (w *Writer) CreateAttachment(h AttachmentHeader) (io.WriteCloser, error) {
	h = AttachmentHeader{h.Header.Copy()} // don't modify the caller's view
	initAttachmentHeader(&h)
	return w.mw.CreatePart(h.Header)
}

// Close finishes the Writer.
func (w *Writer) Close() error {
	return w.mw.Close()
}

// AlternativeWriter writes a mail message's text.
type AlternativeWriter struct {
	mw *message.Writer
}

// CreatePart creates a new text part with the provided header. The body of the
// part should be written to the returned io.WriteCloser.
func (w *AlternativeWriter) CreatePart(h InlineHeader) (io.WriteCloser, error) {
	h = InlineHeader{h.Header.Copy()} // don't modify the caller's view
	initInlineHeader(&h)
	return w.mw.CreatePart(h.Header)
}

// Close finishes the AlternativeWriter.
func (w *AlternativeWriter) Close() error {
	return w.mw.Close()
}

// InlineWriter writes a mail message's text.
// Deprecated: Use AlternativeWriter instead.
type InlineWriter struct {
	mw *message.Writer
}

// CreatePart creates a new text part with the provided header. The body of the
// part should be written to the returned io.WriteCloser.
// Deprecated: Use AlternativeWriter and its CreatePart instead.
func (w *InlineWriter) CreatePart(h InlineHeader) (io.WriteCloser, error) {
	h = InlineHeader{h.Header.Copy()} // don't modify the caller's view
	initInlineHeader(&h)
	return w.mw.CreatePart(h.Header)
}

// Close finishes the InlineWriter.
func (w *InlineWriter) Close() error {
	return w.mw.Close()
}

// RelatedWriter write a mail message with inline attachments and text parts.
type RelatedWriter struct {
	mw *message.Writer
}

// CreateAlternative creates an AlternativeWriter. One or more parts representing the same
// text in different formats can be written to the AlternativeWriter.
func (w *RelatedWriter) CreateAlternative() (*AlternativeWriter, error) {
	var h message.Header
	h.Set("Content-Type", "multipart/alternative")

	mw, err := w.mw.CreatePart(h)
	if err != nil {
		return nil, err
	}
	return &AlternativeWriter{mw}, nil
}

// CreateSingleInline creates a new single text part with the provided header.
// The body of the part should be written to the returned io.WriteCloser. Only
// one single text part should be written, use CreateAlternative if you want multiple
// text parts.
func (w *RelatedWriter) CreateSingleInline(h InlineHeader) (io.WriteCloser, error) {
	h = InlineHeader{h.Header.Copy()} // don't modify the caller's view
	initInlineHeader(&h)
	return w.mw.CreatePart(h.Header)
}

// CreateInlineAttachment creates a new inline attachment with the provided header.
// The body of the part should be written to the returned io.WriteCloser.
func (w *RelatedWriter) CreateInlineAttachment(h InlineAttachmentHeader) (io.WriteCloser, error) {
	h = InlineAttachmentHeader{h.Header.Copy()} // don't modify the caller's view
	initInlineAttachmentHeader(&h)
	return w.mw.CreatePart(h.Header)
}

// Close finishes the RelatedWriter.
func (w *RelatedWriter) Close() error {
	return w.mw.Close()
}
