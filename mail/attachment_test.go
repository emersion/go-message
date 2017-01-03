package mail_test

import (
	"testing"

	"github.com/emersion/go-message/mail"
)

func TestAttachmentHeader_Filename_inContentType(t *testing.T) {
	// Note: putting the attachment's filename in Content-Type is discouraged.

	h := mail.NewAttachmentHeader()
	h.Set("Content-Type", "text/plain; name=note.txt")

	if filename, err := h.Filename(); err != nil {
		t.Error("Expected no error while parsing filename, got:", err)
	} else if filename != "note.txt" {
		t.Errorf("Expected filename to be %q but got %q", "note.txt", filename)
	}
}
