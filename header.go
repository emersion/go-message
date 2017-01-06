package message

import (
	"mime"
	"net/textproto"

	"github.com/emersion/go-message/internal"
)

func parseHeaderWithParams(s string) (f string, params map[string]string, err error) {
	f, params, err = mime.ParseMediaType(s)
	if err != nil {
		return
	}
	for k, v := range params {
		params[k], _ = internal.DecodeHeader(v)
	}
	return
}

func formatHeaderWithParams(f string, params map[string]string) string {
	encParams := make(map[string]string)
	for k, v := range params {
		encParams[k] = internal.EncodeHeader(v)
	}
	return mime.FormatMediaType(f, encParams)
}

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
	return parseHeaderWithParams(h.Get("Content-Type"))
}

// SetContentType formats the Content-Type header field.
func (h Header) SetContentType(t string, params map[string]string) {
	h.Set("Content-Type", formatHeaderWithParams(t, params))
}

// ContentDescription parses the Content-Description header field.
func (h Header) ContentDescription() (string, error) {
	return internal.DecodeHeader(h.Get("Content-Description"))
}

// SetContentDescription parses the Content-Description header field.
func (h Header) SetContentDescription(desc string) {
	h.Set("Content-Description", internal.EncodeHeader(desc))
}

// ContentDisposition parses the Content-Disposition header field, as defined in
// RFC 2183.
func (h Header) ContentDisposition() (disp string, params map[string]string, err error) {
	return parseHeaderWithParams(h.Get("Content-Disposition"))
}

// SetContentDisposition formats the Content-Disposition header field, as
// defined in RFC 2183.
func (h Header) SetContentDisposition(disp string, params map[string]string) {
	h.Set("Content-Disposition", formatHeaderWithParams(disp, params))
}
