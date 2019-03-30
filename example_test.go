package message_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/emersion/go-message"
)

func ExampleHeader2() {
	var h message.Header2
	h.Add("From", "<root@nsa.gov>")
	h.Add("To", "<root@gchq.gov.uk>")
	h.Set("Subject", "Tonight's dinner")

	fmt.Println("From: ", h.Get("From"))
	fmt.Println("Has Received: ", h.Has("Received"))

	fmt.Println("Header fields:")
	fields := h.Fields()
	for fields.Next() {
		fmt.Println("  ", fields.Key())
	}
}

func ExampleRead() {
	// Let's assume r is an io.Reader that contains a message.
	var r io.Reader

	m, err := message.Read(r)
	if message.IsUnknownEncoding(err) {
		// This error is not fatal
		log.Println("Unknown encoding:", err)
	} else if err != nil {
		log.Fatal(err)
	}

	if mr := m.MultipartReader(); mr != nil {
		// This is a multipart message
		log.Println("This is a multipart message containing:")
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			t, _, _ := p.Header.ContentType()
			log.Println("A part with type", t)
		}
	} else {
		t, _, _ := m.Header.ContentType()
		log.Println("This is a non-multipart message with type", t)
	}
}

func ExampleWriter() {
	var b bytes.Buffer

	h := make(message.Header)
	h.SetContentType("multipart/alternative", nil)
	w, err := message.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	h1 := make(message.Header)
	h1.SetContentType("text/html", nil)
	w1, err := w.CreatePart(h1)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w1, "<h1>Hello World!</h1><p>This is an HTML part.</p>")
	w1.Close()

	h2 := make(message.Header)
	h1.SetContentType("text/plain", nil)
	w2, err := w.CreatePart(h2)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w2, "Hello World!\n\nThis is a text part.")
	w2.Close()

	w.Close()

	log.Println(b.String())
}

func Example_transform() {
	// Let's assume r is an io.Reader that contains a message.
	var r io.Reader

	m, err := message.Read(r)
	if message.IsUnknownEncoding(err) {
		log.Println("Unknown encoding:", err)
	} else if err != nil {
		log.Fatal(err)
	}

	// We'll add "This message is powered by Go" at the end of each text entity.
	poweredBy := "\n\nThis message is powered by Go."

	var b bytes.Buffer
	w, err := message.CreateWriter(&b, m.Header)
	if err != nil {
		log.Fatal(err)
	}

	// Define a function that transforms message.
	var transform func(w *message.Writer, e *message.Entity) error
	transform = func(w *message.Writer, e *message.Entity) error {
		if mr := e.MultipartReader(); mr != nil {
			// This is a multipart entity, transform each of its parts
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}

				pw, err := w.CreatePart(p.Header)
				if err != nil {
					return err
				}

				if err := transform(pw, p); err != nil {
					return err
				}

				pw.Close()
			}
			return nil
		} else {
			body := e.Body
			if strings.HasPrefix(m.Header.Get("Content-Type"), "text/") {
				body = io.MultiReader(body, strings.NewReader(poweredBy))
			}
			_, err := io.Copy(w, body)
			return err
		}
	}

	if err := transform(w, m); err != nil {
		log.Fatal(err)
	}
	w.Close()

	log.Println(b.String())
}
