package mail

import (
	"github.com/emersion/go-message"
)

// A TextHeader represents a message text header.
type TextHeader struct {
	message.Header
}

// NewTextHeader creates a new message text header.
func NewTextHeader() TextHeader {
	h := TextHeader{make(message.Header)}
	h.Set("Content-Disposition", "inline")
	h.Set("Content-Transfer-Encoding", "quoted-printable")
	return h
}
