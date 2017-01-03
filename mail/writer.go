package mail

import (
	"io"
	"net/textproto"

	"github.com/emersion/go-message"
)

type Writer struct {
	mw *message.Writer
}

func CreateWriter(w io.Writer, header Header) (*Writer, error) {
	header.Set("Content-Type", "multipart/mixed")

	mw, err := message.CreateWriter(w, header.MIMEHeader)
	if err != nil {
		return nil, err
	}

	return &Writer{mw}, nil
}

func (w *Writer) CreateText() (*TextWriter, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "multipart/alternative")

	mw, err := w.mw.CreatePart(h)
	if err != nil {
		return nil, err
	}
	return &TextWriter{mw}, nil
}

func (w *Writer) CreateAttachment(header AttachmentHeader) (io.WriteCloser, error) {
	return w.mw.CreatePart(header.MIMEHeader)
}

func (w *Writer) Close() error {
	return w.mw.Close()
}

type TextWriter struct {
	mw *message.Writer
}

func (w *TextWriter) CreatePart(header TextHeader) (io.WriteCloser, error) {
	return w.mw.CreatePart(header.MIMEHeader)
}

func (w *TextWriter) Close() error {
	return w.mw.Close()
}
