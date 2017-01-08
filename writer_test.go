package message

import (
	"bytes"
	"io"
	"testing"
)

func TestWriter_multipartWithoutCreatePart(t *testing.T) {
	h := make(Header)
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
