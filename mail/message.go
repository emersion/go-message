package mail

import (
	"bytes"
	"io"
	"net/mail"
	"net/textproto"
	"time"

	"github.com/emersion/go-messages"
)

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// A Message represents a parsed mail message.
type Message struct {
	Header      textproto.MIMEHeader
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

// ReadMessage reads a message from r and buffers its text and attachments.
func ReadMessage(r io.Reader) (*Message, error) {
	e, err := messages.Read(r)
	if err != nil {
		return nil, err
	}

	m := &Message{Header: e.Header}
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

// AddressList parses the named header field as a list of addresses.
func (m *Message) AddressList(key string) ([]*mail.Address, error) {
	return mail.Header(m.Header).AddressList(key)
}

// SetAddressList formats the named header to the provided list of addresses.
func (m *Message) SetAddressList(key string, addrs []*mail.Address) {
	m.Header.Set(key, formatAddressList(addrs))
}

// Date parses the Date header field.
func (m *Message) Date() (time.Time, error) {
	return mail.Header(m.Header).Date()
}

// SetDate formats the Date header field.
func (m *Message) SetDate(t time.Time) {
	m.Header.Set("Date", t.Format(dateLayout))
}
