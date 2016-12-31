package messages_test

import (
	"bytes"
	"io"
	"log"
	"net/textproto"

	"github.com/emersion/go-messages"
)

func ExampleReader() {
	// Let's assume r is an io.Reader that contains a message.
	var r io.Reader

	m, err := messages.Read(r)
	if err != nil {
		log.Fatal(err)
	}

	if pr := m.PartsReader(); pr != nil {
		// This is a multipart message
		log.Println("This is a multipart message containing:")
		for {
			p, err := pr.NextPart()
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
