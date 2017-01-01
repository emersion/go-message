package messages

import (
	"io"
	"io/ioutil"
	"net/textproto"
	"strings"
	"testing"
)

func TestNewEntity(t *testing.T) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "text/plain; charset=US-ASCII")
	h.Set("Content-Transfer-Encoding", "base64")

	r := strings.NewReader("Y2Mgc2F2YQ==")

	e := NewEntity(h, r)

	if e.Header.Get("Content-Transfer-Encoding") != "" {
		t.Error("Expected Content-Transfer-Encoding to be unset")
	}
	if e.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Expected Content-Type charset to be utf-8, got %s", e.Header.Get("Content-Type"))
	}

	expected := "cc sava"
	if b, err := ioutil.ReadAll(e); err != nil {
		t.Error("Expected no error while reading entity body, got", err)
	} else if s := string(b); s != expected {
		t.Errorf("Expected %q as entity body but got %q", expected, s)
	}
}

func testMakeMultipart() *Entity {
	h1 := make(textproto.MIMEHeader)
	h1.Set("Content-Type", "text/plain")
	r1 := strings.NewReader("Text part")
	e1 := NewEntity(h1, r1)

	h2 := make(textproto.MIMEHeader)
	h2.Set("Content-Type", "text/html")
	r2 := strings.NewReader("<p>HTML part</p>")
	e2 := NewEntity(h2, r2)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "multipart/alternative")
	return NewMultipart(h, []*Entity{e1, e2})
}

func TestNewMultipart(t *testing.T) {
	mr := testMakeMultipart().MultipartReader()

	i := 0
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal("Expected no error while reading multipart entity, got", err)
		}

		var expectedType string
		var expectedBody string
		switch i {
		case 0:
			expectedType = "text/plain"
			expectedBody = "Text part"
		case 1:
			expectedType = "text/html"
			expectedBody = "<p>HTML part</p>"
		}

		if mediaType := p.Header.Get("Content-Type"); mediaType != expectedType {
			t.Errorf("Expected part Content-Type to be %q, got %q", expectedType, mediaType)
		}
		if b, err := ioutil.ReadAll(p); err != nil {
			t.Error("Expected no error while reading part body, got", err)
		} else if s := string(b); s != expectedBody {
			t.Errorf("Expected %q as part body but got %q", expectedBody, s)
		}

		i++
	}

	if i != 2 {
		t.Fatalf("Expected multipart entity to contain exactly 2 parts, got %v", i)
	}
}
