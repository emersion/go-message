package message

import (
	"bytes"
	"io"
	"testing"
)

func TestWriter_multipartWithoutCreatePart(t *testing.T) {
	var h Header
	h.Set("Content-Type", "multipart/alternative; boundary=IMTHEBOUNDARY")

	var b bytes.Buffer
	mw, err := CreateWriter(&b, h)
	if err != nil {
		t.Fatal("Expected no error while creating message writer, got:", err)
	}

	io.WriteString(mw, testMultipartBody)
	mw.Close()

	if s := b.String(); s != testMultipartText {
		t.Errorf("Expected output to be \n%s\n but go \n%s", testMultipartText, s)
	}
}

func TestWriter_multipartWithoutBoundary(t *testing.T) {
	var h Header
	h.Set("Content-Type", "multipart/alternative")

	var b bytes.Buffer
	mw, err := CreateWriter(&b, h)
	if err != nil {
		t.Fatal("Expected no error while creating message writer, got:", err)
	}
	mw.Close()

	e, err := Read(&b)
	if err != nil {
		t.Fatal("Expected no error while reading message, got:", err)
	}

	mediaType, mediaParams, err := e.Header.ContentType()
	if err != nil {
		t.Fatal("Expected no error while parsing Content-Type, got:", err)
	} else if mediaType != "multipart/alternative" {
		t.Errorf("Expected media type to be %q, but got %q", "multipart/alternative", mediaType)
	} else if boundary, ok := mediaParams["boundary"]; !ok || boundary == "" {
		t.Error("Expected boundary to be automatically generated")
	}
}
