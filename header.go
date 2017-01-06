package message

import (
	"mime"
	"net/textproto"

	"github.com/emersion/go-message/internal"
)

// A Header represents the key-value pairs in a message header.
type Header map[string][]string

// Add adds the key, value pair to the header. It appends to any existing values
// associated with key.
func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

// Set sets the header entries associated with key to the single element value.
// It replaces any existing values associated with key.
func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

// Get gets the first value associated with the given key. If there are no
// values associated with the key, Get returns "".
func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

// Del deletes the values associated with key.
func (h Header) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

// ContentType parses the Content-Type header field.
func (h Header) ContentType() (t string, params map[string]string, err error) {
	return mime.ParseMediaType(h.Get("Content-Type"))
}

// SetContentType formats the Content-Type header field.
func (h Header) SetContentType(t string, params map[string]string) {
	h.Set("Content-Type", mime.FormatMediaType(t, params))
}

// ContentDescription parses the Content-Description header field.
func (h Header) ContentDescription() (string, error) {
	return internal.DecodeHeader(h.Get("Content-Description"))
}

// SetContentDescription parses the Content-Description header field.
func (h Header) SetContentDescription(desc string) {
	h.Set("Content-Description", internal.EncodeHeader(desc))
}
