package mail

import (
	"io"

	"github.com/emersion/go-message"
)

// A Writer writes a mail message. A mail message contains one or more text
// parts and zero or more attachments.
type Writer struct {
	mw *message.Writer
}

// CreateWriter writes a mail header to w and creates a new Writer.
func CreateWriter(w io.Writer, header Header) (*Writer, error) {
	header.Set("Content-Type", "multipart/mixed")

	mw, err := message.CreateWriter(w, header.Header)
	if err != nil {
		return nil, err
	}

	return &Writer{mw}, nil
}

// CreateText creates a TextWriter. One or more parts representing the same text
// in different formats can be written to a TextWriter.
func (w *Writer) CreateText() (*TextWriter, error) {
	h := make(message.Header)
	h.Set("Content-Type", "multipart/alternative")

	mw, err := w.mw.CreatePart(h)
	if err != nil {
		return nil, err
	}
	return &TextWriter{mw}, nil
}

// CreatePart creates a new single text part with the provided header. The body
// of the part should be written to the returned io.WriteCloser. Only one single
// text part should be written, use CreateText if you want multiple text parts.
func (w *Writer) CreateSingleText(header TextHeader) (io.WriteCloser, error) {
	return w.mw.CreatePart(header.Header)
}

// CreateAttachment creates a new attachment with the provided header. The body
// of the part should be written to the returned io.WriteCloser.
func (w *Writer) CreateAttachment(header AttachmentHeader) (io.WriteCloser, error) {
	return w.mw.CreatePart(header.Header)
}

// Close finishes the Writer.
func (w *Writer) Close() error {
	return w.mw.Close()
}

// TextWriter writes a mail message's text.
type TextWriter struct {
	mw *message.Writer
}

// CreatePart creates a new text part with the provided header. The body of the
// part should be written to the returned io.WriteCloser.
func (w *TextWriter) CreatePart(header TextHeader) (io.WriteCloser, error) {
	return w.mw.CreatePart(header.Header)
}

// Close finishes the TextWriter.
func (w *TextWriter) Close() error {
	return w.mw.Close()
}
