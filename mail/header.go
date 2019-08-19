package mail

import (
	"net/mail"
	"regexp"
	"time"

	"github.com/emersion/go-message"
)

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// TODO: this is a blunt way to strip any trailing CFWS (comment). A sharper
// one would strip multiple CFWS, and only if really valid according to
// RFC5322.
var commentRE = regexp.MustCompile(`[ \t]+\(.*\)$`)

// A Header is a mail header.
type Header struct {
	message.Header
}

// AddressList parses the named header field as a list of addresses. If the
// header is missing, it returns nil.
func (h *Header) AddressList(key string) ([]*Address, error) {
	v := h.Get(key)
	if v == "" {
		return nil, nil
	}
	return parseAddressList(v)
}

// SetAddressList formats the named header to the provided list of addresses.
func (h *Header) SetAddressList(key string, addrs []*Address) {
	h.Set(key, formatAddressList(addrs))
}

// Date parses the Date header field.
func (h *Header) Date() (time.Time, error) {
	//TODO: remove this once https://go-review.googlesource.com/c/go/+/117596/
	// is merged
	date := commentRE.ReplaceAllString(h.Get("Date"), "")
	return mail.ParseDate(date)
}

// SetDate formats the Date header field.
func (h *Header) SetDate(t time.Time) {
	h.Set("Date", t.Format(dateLayout))
}

// Subject parses the Subject header field. If there is an error, the raw field
// value is returned alongside the error.
func (h *Header) Subject() (string, error) {
	return h.Text("Subject")
}

// SetSubject formats the Subject header field.
func (h *Header) SetSubject(s string) {
	h.SetText("Subject", s)
}
