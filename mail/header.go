package mail

import (
	"net/mail"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/charset"
)

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// A Header is a mail header.
type Header struct {
	message.Header
}

// NewHeader creates a new mail header.
func NewHeader() Header {
	return Header{make(message.Header)}
}

// AddressList parses the named header field as a list of addresses.
func (h Header) AddressList(key string) ([]*Address, error) {
	return parseAddressList(h.Get(key))
}

// SetAddressList formats the named header to the provided list of addresses.
func (h Header) SetAddressList(key string, addrs []*Address) {
	h.Set(key, formatAddressList(addrs))
}

// Date parses the Date header field.
func (h Header) Date() (time.Time, error) {
	return mail.Header(h.Header).Date()
}

// SetDate formats the Date header field.
func (h Header) SetDate(t time.Time) {
	h.Set("Date", t.Format(dateLayout))
}

// Subject parses the Subject header field. If there is an error, the raw field
// value is returned alongside the error.
func (h Header) Subject() (string, error) {
	return charset.DecodeHeader(h.Get("Subject"))
}

// SetSubject formats the Subject header field.
func (h Header) SetSubject(s string) {
	h.Set("Subject", charset.EncodeHeader(s))
}
