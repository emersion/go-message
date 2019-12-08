package mail_test

import (
	"testing"

	"github.com/emersion/go-message/mail"
)

func TestAttachmentHeader_Filename(t *testing.T) {
	var h mail.AttachmentHeader
	h.Set("Content-Disposition", "attachment; filename=note.txt")

	if filename, err := h.Filename(); err != nil {
		t.Error("Expected no error while parsing filename, got:", err)
	} else if filename != "note.txt" {
		t.Errorf("Expected filename to be %q but got %q", "note.txt", filename)
	}
}

func TestAttachmentHeader_Filename_inContentType(t *testing.T) {
	// Note: putting the attachment's filename in Content-Type is discouraged.

	var h mail.AttachmentHeader
	h.Set("Content-Type", "text/plain; name=note.txt")

	if filename, err := h.Filename(); err != nil {
		t.Error("Expected no error while parsing filename, got:", err)
	} else if filename != "note.txt" {
		t.Errorf("Expected filename to be %q but got %q", "note.txt", filename)
	}
}

func TestAttachmentHeader_Filename_none(t *testing.T) {
	var h mail.AttachmentHeader
	if filename, err := h.Filename(); err != nil {
		t.Error("Expected no error while parsing filename, got:", err)
	} else if filename != "" {
		t.Errorf("Expected filename to be %q but got %q", "", filename)
	}
}

func TestAttachmentHeader_Filename_encoded(t *testing.T) {
	var h mail.AttachmentHeader
	h.Set("Content-Disposition", "attachment; filename=\"=?UTF-8?Q?Opis_przedmiotu_zam=c3=b3wienia_-_za=c5=82=c4=85cznik_nr_1?= =?UTF-8?Q?=2epdf?=\"")

	if filename, err := h.Filename(); err != nil {
		t.Error("Expected no error while parsing filename, got:", err)
	} else if filename != "Opis przedmiotu zamówienia - załącznik nr 1.pdf" {
		t.Errorf("Expected filename to be %q but got %q", "", filename)
	}
}
