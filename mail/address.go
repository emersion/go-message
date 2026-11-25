package mail

import (
	"mime"
	"net/mail"
	"regexp"
	"strings"

	"github.com/emersion/go-message"
)

// Address represents a single mail address.
// The type alias ensures that a net/mail.Address can be used wherever an
// Address is expected
type Address = mail.Address

// unquotedAtInDisplayName matches an unquoted "@" inside the display name of
// a name-addr. Some mail providers produce such addresses; the stdlib parser
// rejects them because "@" is an RFC 5322 special in an atom. Quote the
// display name before handing it to the parser.
var unquotedAtInDisplayName = regexp.MustCompile(`((?:[^"<>,\\])*?)\s*<`)

func formatAddressList(l []*Address) string {
	formatted := make([]string, len(l))
	for i, a := range l {
		formatted[i] = a.String()
	}
	return strings.Join(formatted, ", ")
}

// ParseAddress parses a single RFC 5322 address, e.g. "Barry Gibbs <bg@example.com>"
// Use this function only if you parse from a string, if you have a Header use
// Header.AddressList instead
func ParseAddress(address string) (*Address, error) {
	parser := mail.AddressParser{
		&mime.WordDecoder{message.CharsetReader},
	}
	return parser.Parse(quoteUnquotedAtInDisplayName(address))
}

// ParseAddressList parses the given string as a list of addresses.
// Use this function only if you parse from a string, if you have a Header use
// Header.AddressList instead
func ParseAddressList(list string) ([]*Address, error) {
	parser := mail.AddressParser{
		&mime.WordDecoder{message.CharsetReader},
	}
	return parser.ParseList(quoteUnquotedAtInDisplayName(list))
}

// quoteUnquotedAtInDisplayName quotes display names containing an unquoted
// "@" so that lenient providers' addresses can still be parsed.
func quoteUnquotedAtInDisplayName(s string) string {
	return unquotedAtInDisplayName.ReplaceAllStringFunc(s, func(match string) string {
		sub := unquotedAtInDisplayName.FindStringSubmatch(match)
		name := strings.TrimSpace(sub[1])
		if !strings.Contains(name, "@") || strings.HasPrefix(name, `"`) {
			return match
		}
		quoted := strings.ReplaceAll(name, `\`, `\\`)
		quoted = strings.ReplaceAll(quoted, `"`, `\"`)
		return `"` + quoted + `" <`
	})
}
