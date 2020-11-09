package mail

import (
	"mime"
	"net/mail"
	"strings"

	"github.com/emersion/go-message"
)

// Address represents a single mail address.
// The type alias ensures that a net/mail.Address can be used wherever an
// Address is expected
type Address = mail.Address

func parseAddressList(s string) ([]*Address, error) {
	parser := mail.AddressParser{
		&mime.WordDecoder{message.CharsetReader},
	}
	return parser.ParseList(s)
}

func formatAddressList(l []*Address) string {
	formatted := make([]string, len(l))
	for i, a := range l {
		formatted[i] = a.String()
	}
	return strings.Join(formatted, ", ")
}
