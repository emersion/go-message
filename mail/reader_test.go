package mail_test

import (
	"io"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/emersion/go-messages/mail"
)

func ExampleReader() {
	// Let's assume r is an io.Reader that contains a mail.
	var r io.Reader

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Fatal(err)
	}

	// Read each mail's part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		switch h := p.Header.(type) {
		case mail.TextHeader:
			b, _ := ioutil.ReadAll(p.Body)
			log.Println("Got text: %v", string(b))
		case mail.AttachmentHeader:
			filename, _ := h.Filename()
			log.Println("Got attachment: %v", filename)
		}
	}
}

func TestReader(t *testing.T) {
	r := strings.NewReader(mailString)
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Fatal(err)
	}
	defer mr.Close()

	i := 0
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		var expectedBody string
		switch i {
		case 0:
			h, ok := p.Header.(mail.TextHeader)
			if !ok {
				t.Fatalf("Expected a TextHeader, but got a %T", p.Header)
			}

			if h.IsHTML() {
				t.Errorf("Expected a plaintext part, not an HTML part")
			}

			expectedBody = "Who are you?"
		case 1:
			h, ok := p.Header.(mail.AttachmentHeader)
			if !ok {
				t.Fatalf("Expected an AttachmentHeader, but got a %T", p.Header)
			}

			if filename, err := h.Filename(); err != nil {
				t.Error("Expected no error while parsing filename, but got:", err)
			} else if filename != "note.txt" {
				t.Errorf("Expected filename to be %q but got %q", "note.txt", filename)
			}

			expectedBody = "I'm Mitsuha."
		}

		if b, err := ioutil.ReadAll(p.Body); err != nil {
			t.Error("Expected no error while reading part body, but got:", err)
		} else if string(b) != expectedBody {
			t.Errorf("Expected part body to be:\n%v\nbut got:\n%v", expectedBody, string(b))
		}

		i++
	}

	if i != 2 {
		t.Errorf("Expected exactly two parts but got %v", i)
	}
}
