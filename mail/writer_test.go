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
	h := mail.NewHeader()
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)

	// Create a new mail writer
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateText()
	if err != nil {
		log.Fatal(err)
	}
	th := mail.NewTextHeader()
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()
	tw.Close()

	// Create an attachment
	ah := mail.NewAttachmentHeader()
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

	h := mail.NewHeader()
	h.SetSubject("Your Name")
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateText()
	if err != nil {
		t.Fatal(err)
	}
	th := mail.NewTextHeader()
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()
	tw.Close()

	// Create an attachment
	ah := mail.NewAttachmentHeader()
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

func TestWriter_singleText(t *testing.T) {
	var b bytes.Buffer

	h := mail.NewHeader()
	h.SetSubject("Your Name")
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	// Create a text part
	th := mail.NewTextHeader()
	th.Set("Content-Type", "text/plain")
	w, err := mw.CreateSingleText(th)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()

	// Create an attachment
	ah := mail.NewAttachmentHeader()
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
