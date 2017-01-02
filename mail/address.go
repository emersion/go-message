package mail

import (
	"net/mail"
	"strings"
)

func formatAddressList(l []*mail.Address) string {
	formatted := make([]string, len(l))
	for i, a := range l {
		formatted[i] = a.String()
	}
	return strings.Join(formatted, ", ")
}
