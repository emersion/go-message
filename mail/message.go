package mail

import (
	"io"
	"net/mail"
	"net/textproto"
	"time"
)

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// A Message represents a parsed mail message.
type Message struct {
	Header      textproto.MIMEHeader
	Text        io.Reader
	HTML        io.Reader
	Attachments []*Attachment
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
