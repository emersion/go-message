package mail

import (
	"net/mail"
	"net/textproto"
	"time"
)

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// A Header is a mail header.
type Header textproto.MIMEHeader

// Add adds the key, value pair to the header. It appends to any existing values
// associated with key.
func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

// Del deletes the values associated with key.
func (h Header) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

// Get gets the first value associated with the given key. If there are no
// values associated with the key, Get returns "".
func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

// Set sets the header entries associated with key to the single element value.
// It replaces any existing values associated with key.
func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

// AddressList parses the named header field as a list of addresses.
func (h Header) AddressList(key string) ([]*mail.Address, error) {
	return mail.Header(h).AddressList(key)
}

// SetAddressList formats the named header to the provided list of addresses.
func (h Header) SetAddressList(key string, addrs []*mail.Address) {
	h.Set(key, formatAddressList(addrs))
}

// Date parses the Date header field.
func (h Header) Date() (time.Time, error) {
	return mail.Header(h).Date()
}

// SetDate formats the Date header field.
func (h Header) SetDate(t time.Time) {
	h.Set("Date", t.Format(dateLayout))
}
