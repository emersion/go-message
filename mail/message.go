package mail

import (
	"bytes"
	"io"

	"github.com/emersion/go-messages"
)

// A Message represents a parsed mail message.
type Message struct {
	Header      Header
	Text        io.Reader
	HTML        io.Reader
	Attachments []*Attachment
}

func buffer(r io.Reader) (io.Reader, error) {
	b := new(bytes.Buffer)
	if _, err := io.Copy(b, r); err != nil {
		return nil, err
	}
	return b, nil
}

// ReadMessage reads a message from r and buffers its text and attachments. If
// possible, use a Reader instead.
func ReadMessage(r io.Reader) (*Message, error) {
	e, err := messages.Read(r)
	if err != nil {
		return nil, err
	}

	m := &Message{Header: Header(e.Header)}
	mr := NewReader(e)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return m, err
		}

		// Buffer p.Body because it gets discarded when NextPart is called
		switch p := p.(type) {
		case *Text:
			p.Body, err = buffer(p.Body)
			if p.IsHTML() {
				m.HTML = p.Body
			} else {
				m.Text = p.Body
			}
		case *Attachment:
			p.Body, err = buffer(p.Body)
			m.Attachments = append(m.Attachments, p)
		}
	}

	return m, nil
}
