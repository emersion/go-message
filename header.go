package message

import (
	"mime"
	"net/textproto"
	"strings"

	"github.com/emersion/go-message/charset"
)

const maxHeaderLen = 76

func parseHeaderWithParams(s string) (f string, params map[string]string, err error) {
	f, params, err = mime.ParseMediaType(s)
	if err != nil {
		return s, nil, err
	}
	for k, v := range params {
		params[k], _ = charset.DecodeHeader(v)
	}
	return
}

func formatHeaderWithParams(f string, params map[string]string) string {
	encParams := make(map[string]string)
	for k, v := range params {
		encParams[k] = charset.EncodeHeader(v)
	}
	return mime.FormatMediaType(f, encParams)
}

// formatHeaderField formats a header field, ensuring each line is no longer
// than 76 characters. It tries to fold lines at whitespace characters if
// possible. If the header contains a word longer than this limit, it will be
// split.
func formatHeaderField(k, v string) string {
	s := k + ": "

	first := true
	for len(v) > 0 {
		maxlen := maxHeaderLen
		if first {
			maxlen -= len(s)
		}

		// We'll need to fold before i
		foldBefore := maxlen + 1
		foldAt := len(v)

		var folding string
		if foldBefore > len(v) {
			// We reached the end of the string
			if v[len(v)-1] != '\n' {
				// If there isn't already a trailing CRLF, insert one
				folding = "\r\n"
			}
		} else {
			// Find the closest whitespace before i
			foldAt = strings.LastIndexAny(v[:foldBefore], " \t\n")
			if foldAt == 0 {
				// The whitespace we found was the previous folding WSP
				foldAt = foldBefore - 1
			} else if foldAt < 0 {
				// We didn't find any whitespace, we have to insert one
				foldAt = foldBefore - 2
			}

			switch v[foldAt] {
			case ' ', '\t':
				if v[foldAt-1] != '\n' {
					folding = "\r\n" // The next char will be a WSP, don't need to insert one
				}
			case '\n':
				folding = "" // There is already a CRLF, nothing to do
			default:
				folding = "\r\n " // Another char, we need to insert CRLF + WSP
			}
		}

		s += v[:foldAt] + folding
		v = v[foldAt:]
		first = false
	}

	return s
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
	return charset.DecodeHeader(h.Get("Content-Description"))
}

// SetContentDescription parses the Content-Description header field.
func (h Header) SetContentDescription(desc string) {
	h.Set("Content-Description", charset.EncodeHeader(desc))
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
