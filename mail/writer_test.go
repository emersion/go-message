package mail_test

import (
	"bytes"
	"io"
	"log"
	"testing"
	"time"

	"github.com/emersion/go-message/mail"
)

func ExampleWriter() {
	var b bytes.Buffer

	from := []*mail.Address{{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"}}
	to := []*mail.Address{{"Taki Tachibana", "taki.tachibana@example.org"}}

	// Create our mail header
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)

	// Create a new mail writer
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateAlternative()
	if err != nil {
		log.Fatal(err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()
	tw.Close()

	// Create an attachment
	var ah mail.AttachmentHeader
	ah.Set("Content-Type", "image/jpeg")
	ah.SetFilename("picture.jpg")
	w, err = mw.CreateAttachment(ah)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: write a JPEG file to w
	w.Close()

	mw.Close()

	log.Println(b.String())
}

func TestWriter(t *testing.T) {
	var b bytes.Buffer

	var h mail.Header
	h.SetSubject("Your Name")
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateAlternative()
	if err != nil {
		t.Fatal(err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()
	tw.Close()

	// Create an attachment
	var ah mail.AttachmentHeader
	ah.Set("Content-Type", "text/plain")
	ah.SetFilename("note.txt")
	w, err = mw.CreateAttachment(ah)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "I'm Mitsuha.")
	w.Close()

	mw.Close()

	testReader(t, &b)
}

func TestWriter_singleInline(t *testing.T) {
	var b bytes.Buffer

	var h mail.Header
	h.SetSubject("Your Name")
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	// Create a text part
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := mw.CreateSingleInline(th)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()

	// Create an attachment
	var ah mail.AttachmentHeader
	ah.Set("Content-Type", "text/plain")
	ah.SetFilename("note.txt")
	w, err = mw.CreateAttachment(ah)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "I'm Mitsuha.")
	w.Close()

	mw.Close()

	t.Logf("Formatted message: \n%v", b.String())

	testReader(t, &b)
}
