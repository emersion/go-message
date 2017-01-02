package messages

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/textproto"
	"strings"
	"testing"
)

func testMakeEntity() *Entity {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "text/plain; charset=US-ASCII")
	h.Set("Content-Transfer-Encoding", "base64")

	r := strings.NewReader("Y2Mgc2F2YQ==")

	return NewEntity(h, r)
}

func TestNewEntity(t *testing.T) {
	e := testMakeEntity()

	if e.Header.Get("Content-Transfer-Encoding") != "" {
		t.Error("Expected Content-Transfer-Encoding to be unset")
	}
	if e.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Expected Content-Type charset to be utf-8, got %s", e.Header.Get("Content-Type"))
	}

	expected := "cc sava"
	if b, err := ioutil.ReadAll(e.Body); err != nil {
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
	h.Set("Content-Type", "multipart/alternative; boundary=IMTHEBOUNDARY")
	return NewMultipart(h, []*Entity{e1, e2})
}

const testMultipartHeader = "Content-Type: multipart/alternative; boundary=IMTHEBOUNDARY\r\n" +
	"\r\n"

const testMultipartBody = "--IMTHEBOUNDARY\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" +
	"Text part\r\n" +
	"--IMTHEBOUNDARY\r\n" +
	"Content-Type: text/html\r\n" +
	"\r\n" +
	"<p>HTML part</p>\r\n" +
	"--IMTHEBOUNDARY--\r\n"

var testMultipartText = testMultipartHeader + testMultipartBody

func testMultipart(t *testing.T, e *Entity) {
	mr := e.MultipartReader()
	if mr == nil {
		t.Fatalf("Expected MultipartReader not to return nil")
	}
	defer mr.Close()

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
		if b, err := ioutil.ReadAll(p.Body); err != nil {
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

func TestNewMultipart(t *testing.T) {
	testMultipart(t, testMakeMultipart())
}

func TestNewMultipart_read(t *testing.T) {
	e := testMakeMultipart()

	if b, err := ioutil.ReadAll(e.Body); err != nil {
		t.Error("Expected no error while reading multipart body, got", err)
	} else if s := string(b); s != testMultipartBody {
		t.Errorf("Expected %q as multipart body but got %q", testMultipartBody, s)
	}
}

func TestRead_multipart(t *testing.T) {
	e, err := Read(strings.NewReader(testMultipartText))
	if err != nil {
		t.Fatal("Expected no error while reading multipart, got", err)
	}

	testMultipart(t, e)
}

func TestEntity_WriteTo(t *testing.T) {
	e := testMakeEntity()

	var b bytes.Buffer
	if err := e.WriteTo(&b); err != nil {
		t.Fatal("Expected no error while writing entity, got", err)
	}

	expected := "Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"cc sava"

	if s := b.String(); s != expected {
		t.Errorf("Expected written entity to be:\n%s\nbut got:\n%s", expected, s)
	}
}

func TestEntity_WriteTo_multipart(t *testing.T) {
	e := testMakeMultipart()

	var b bytes.Buffer
	if err := e.WriteTo(&b); err != nil {
		t.Fatal("Expected no error while writing entity, got", err)
	}

	if s := b.String(); s != testMultipartText {
		t.Errorf("Expected written entity to be:\n%s\nbut got:\n%s", testMultipartText, s)
	}
}
