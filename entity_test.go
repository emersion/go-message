package message

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func testMakeEntity() *Entity {
	var h Header
	h.Set("Content-Type", "text/plain; charset=US-ASCII")
	h.Set("Content-Transfer-Encoding", "base64")

	r := strings.NewReader("Y2Mgc2F2YQ==")

	e, _ := New(h, r)
	return e
}

func TestNewEntity(t *testing.T) {
	e := testMakeEntity()

	expected := "cc sava"
	if b, err := ioutil.ReadAll(e.Body); err != nil {
		t.Error("Expected no error while reading entity body, got", err)
	} else if s := string(b); s != expected {
		t.Errorf("Expected %q as entity body but got %q", expected, s)
	}
}

func testMakeMultipart() *Entity {
	var h1 Header
	h1.Set("Content-Type", "text/plain")
	r1 := strings.NewReader("Text part")
	e1, _ := New(h1, r1)

	var h2 Header
	h2.Set("Content-Type", "text/html")
	r2 := strings.NewReader("<p>HTML part</p>")
	e2, _ := New(h2, r2)

	var h Header
	h.Set("Content-Type", "multipart/alternative; boundary=IMTHEBOUNDARY")
	e, _ := NewMultipart(h, []*Entity{e1, e2})
	return e
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

const testSingleText = "Content-Type: text/plain\r\n" +
	"\r\n" +
	"Message body"

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

func TestRead_single(t *testing.T) {
	e, err := Read(strings.NewReader(testSingleText))
	if err != nil {
		t.Fatalf("Read() = %v", err)
	}

	b, err := ioutil.ReadAll(e.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll() = %v", err)
	}

	expected := "Message body"
	if string(b) != expected {
		t.Fatalf("Expected body to be %q, got %q", expected, string(b))
	}
}

func TestEntity_WriteTo_decode(t *testing.T) {
	e := testMakeEntity()

	e.Header.SetContentType("text/plain", map[string]string{"charset": "utf-8"})
	e.Header.Del("Content-Transfer-Encoding")

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

func TestEntity_WriteTo_convert(t *testing.T) {
	var h Header
	h.Set("Content-Type", "text/plain; charset=utf-8")
	h.Set("Content-Transfer-Encoding", "base64")
	r := strings.NewReader("Qm9uam91ciDDoCB0b3Vz")
	e, _ := New(h, r)

	e.Header.Set("Content-Transfer-Encoding", "quoted-printable")

	var b bytes.Buffer
	if err := e.WriteTo(&b); err != nil {
		t.Fatal("Expected no error while writing entity, got", err)
	}

	expected := "Content-Transfer-Encoding: quoted-printable\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"Bonjour =C3=A0 tous"

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

func TestNew_unknownTransferEncoding(t *testing.T) {
	var h Header
	h.Set("Content-Transfer-Encoding", "i-dont-exist")

	expected := "hey there"
	r := strings.NewReader(expected)

	e, err := New(h, r)
	if err == nil {
		t.Fatal("New(unknown transfer encoding): expected an error")
	}
	if !isUnknownEncoding(err) {
		t.Fatal("New(unknown transfer encoding): expected an error that verifies isUnknownEncoding")
	}

	if b, err := ioutil.ReadAll(e.Body); err != nil {
		t.Error("Expected no error while reading entity body, got", err)
	} else if s := string(b); s != expected {
		t.Errorf("Expected %q as entity body but got %q", expected, s)
	}
}

func TestNew_unknownCharset(t *testing.T) {
	var h Header
	h.Set("Content-Type", "text/plain; charset=I-DONT-EXIST")

	expected := "hey there"
	r := strings.NewReader(expected)

	e, err := New(h, r)
	if err == nil {
		t.Fatal("New(unknown charset): expected an error")
	}
	if !IsUnknownCharset(err) {
		t.Fatal("New(unknown charset): expected an error that verifies IsUnknownCharset")
	}

	if b, err := ioutil.ReadAll(e.Body); err != nil {
		t.Error("Expected no error while reading entity body, got", err)
	} else if s := string(b); s != expected {
		t.Errorf("Expected %q as entity body but got %q", expected, s)
	}
}

func TestNewEntity_MultipartReader_notMultipart(t *testing.T) {
	e := testMakeEntity()
	mr := e.MultipartReader()
	if mr != nil {
		t.Fatal("(non-multipart).MultipartReader() != nil")
	}
}
