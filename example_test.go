package messages_test

import (
	"bytes"
	"io"
	"log"
	"net/textproto"
	"strings"

	"github.com/emersion/go-messages"
)

func ExampleReader() {
	// Let's assume r is an io.Reader that contains a message.
	var r io.Reader

	m, err := messages.Read(r)
	if err != nil {
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

			log.Println("A part with type", p.Header.Get("Content-Type"))
		}
	} else {
		log.Println("This is a non-multipart message with type", m.Header.Get("Content-Type"))
	}
}

func ExampleWriter() {
	var b bytes.Buffer

	h := textproto.MIMEHeader{"Content-Type": {"multipart/alternative"}}
	w, err := messages.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	h1 := textproto.MIMEHeader{"Content-Type": {"text/html"}}
	w1, err := w.CreatePart(h1)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w1, "<h1>Hello World!</h1><p>This is an HTML part.</p>")
	w1.Close()

	h2 := textproto.MIMEHeader{"Content-Type": {"text/plain"}}
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

	m, err := messages.Read(r)
	if err != nil {
		log.Fatal(err)
	}

	// We'll add "This message is powered by Go" at the end of each text entity.
	poweredBy := "\n\nThis message is powered by Go."

	var b bytes.Buffer
	w, err := messages.CreateWriter(&b, m.Header)
	if err != nil {
		log.Fatal(err)
	}

	// Define a function that transforms messages.
	var transform func(w *messages.Writer, e *messages.Entity) error
	transform = func(w *messages.Writer, e *messages.Entity) error {
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
