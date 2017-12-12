package mail_test

import (
	"io"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/emersion/go-message/mail"
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

func testReader(t *testing.T, r io.Reader) {
	mr, err := mail.CreateReader(r)
	if err != nil {
		t.Fatalf("mail.CreateReader(r) = %v", err)
	}
	defer mr.Close()

	wantSubject := "Your Name"
	subject, err := mr.Header.Subject()
	if err != nil {
		t.Errorf("mr.Header.Subject() = %v", err)
	} else if subject != wantSubject {
		t.Errorf("mr.Header.Subject() = %v, want %v", subject, wantSubject)
	}

	i := 0
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		var expectedBody string
		switch i {
		case 0:
			h, ok := p.Header.(mail.TextHeader)
			if !ok {
				t.Fatalf("Expected a TextHeader, but got a %T", p.Header)
			}

			if mediaType, _, _ := h.ContentType(); mediaType != "text/plain" {
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

func TestReader(t *testing.T) {
	testReader(t, strings.NewReader(mailString))
}

func TestReader_nonMultipart(t *testing.T) {
	s := "Subject: Your Name\r\n" +
		"\r\n" +
		"Who are you?"

	mr, err := mail.CreateReader(strings.NewReader(s))
	if err != nil {
		t.Fatal("Expected no error while creating reader, got:", err)
	}
	defer mr.Close()

	p, err := mr.NextPart()
	if err != nil {
		t.Fatal("Expected no error while reading part, got:", err)
	}

	if _, ok := p.Header.(mail.TextHeader); !ok {
		t.Fatalf("Expected a TextHeader, but got a %T", p.Header)
	}

	expectedBody := "Who are you?"
	if b, err := ioutil.ReadAll(p.Body); err != nil {
		t.Error("Expected no error while reading part body, but got:", err)
	} else if string(b) != expectedBody {
		t.Errorf("Expected part body to be:\n%v\nbut got:\n%v", expectedBody, string(b))
	}

	if _, err := mr.NextPart(); err != io.EOF {
		t.Fatal("Expected io.EOF while reading part, but got:", err)
	}
}

func TestReader_closeImmediately(t *testing.T) {
	s := "Content-Type: text/plain\r\n" +
		"\r\n" +
		"Who are you?"

	mr, err := mail.CreateReader(strings.NewReader(s))
	if err != nil {
		t.Fatal("Expected no error while creating reader, got:", err)
	}

	mr.Close()

	if _, err := mr.NextPart(); err != io.EOF {
		t.Fatal("Expected io.EOF while reading part, but got:", err)
	}
}

func TestReader_nested(t *testing.T) {
	r := strings.NewReader(nestedMailString)

	mr, err := mail.CreateReader(r)
	if err != nil {
		t.Fatalf("mail.CreateReader(r) = %v", err)
	}
	defer mr.Close()

	i := 0
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		switch i {
		case 0:
			_, ok := p.Header.(mail.TextHeader)
			if !ok {
				t.Fatalf("Expected a TextHeader, but got a %T", p.Header)
			}

			expectedBody := "I forgot."
			if b, err := ioutil.ReadAll(p.Body); err != nil {
				t.Error("Expected no error while reading part body, but got:", err)
			} else if string(b) != expectedBody {
				t.Errorf("Expected part body to be:\n%v\nbut got:\n%v", expectedBody, string(b))
			}
		case 1:
			_, ok := p.Header.(mail.AttachmentHeader)
			if !ok {
				t.Fatalf("Expected an AttachmentHeader, but got a %T", p.Header)
			}

			testReader(t, p.Body)
		}

		i++
	}
}
