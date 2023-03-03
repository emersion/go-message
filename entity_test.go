package message

import (
	"bytes"
	"errors"
	"golang.org/x/text/encoding/ianaindex"
	"io"
	"io/ioutil"
	"math"
	"reflect"
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

func testMakeMultipartVaryingCharset() *Entity {
	var h1 Header
	h1.Set("Content-Type", "text/plain; charset=windows-1250")
	r1 := bytes.NewReader([]byte{0x8c, 0x8d, 0x8f, 0x9c, 0x9d, 0x9f, 0xbc, 0xbe})
	e1, _ := New(h1, r1)

	var h2 Header
	h2.Set("Content-Type", "text/html; charset=iso-8859-1")
	r2 := bytes.NewReader([]byte{0x63, 0x61, 0x66, 0xE9})
	e2, _ := New(h2, r2)

	var h Header
	h.Set("Content-Type", "multipart/alternative; boundary=IMTHEBOUNDARY")
	e, _ := NewMultipart(h, []*Entity{e1, e2})
	return e
}

const testMultipartHeader = "Mime-Version: 1.0\r\n" +
	"Content-Type: multipart/alternative; boundary=IMTHEBOUNDARY\r\n\r\n"

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

var testMultipartVaryingCharsetBody = "--IMTHEBOUNDARY\r\n" +
	"Content-Type: text/plain; charset=windows-1250\r\n" +
	"\r\n" +
	string([]byte{0x8c, 0x8d, 0x8f, 0x9c, 0x9d, 0x9f, 0xbc, 0xbe}) + "\r\n" +
	"--IMTHEBOUNDARY\r\n" +
	"Content-Type: text/html; charset=iso-8859-1\r\n" +
	"\r\n" +
	string([]byte{0x63, 0x61, 0x66, 0xE9}) + "\r\n" +
	"--IMTHEBOUNDARY--\r\n"

var testMultipartVaryingCharsetText = testMultipartHeader + testMultipartVaryingCharsetBody

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

func TestRead_tooBig(t *testing.T) {
	raw := "Subject: " + strings.Repeat("A", 4096*1024) + "\r\n" +
		"\r\n" +
		"This header is too big.\r\n"
	_, err := Read(strings.NewReader(raw))
	if err != errHeaderTooBig {
		t.Fatalf("Read() = %q, want %q", err, errHeaderTooBig)
	}
}

func TestReadOptions_withDefaults(t *testing.T) {
	// verify that .withDefaults() doesn't mutate original values
	original := &ReadOptions{MaxHeaderBytes: -123}
	modified := original.withDefaults() // should set MaxHeaderBytes to math.MaxInt64

	if original.MaxHeaderBytes == modified.MaxHeaderBytes {
		t.Error("Expected ReadOptions.withDefaults() to not mutate the original value")
	}
}

func TestReadWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		original *ReadOptions
		want     *ReadOptions
		wantErr  bool
	}{
		{
			name:     "default value",
			original: &ReadOptions{},
			want:     &ReadOptions{MaxHeaderBytes: defaultMaxHeaderBytes},
			wantErr:  true,
		},
		{
			name:     "infinite header value",
			original: &ReadOptions{MaxHeaderBytes: -1},
			want:     &ReadOptions{MaxHeaderBytes: math.MaxInt64},
			wantErr:  false,
		},
		{
			name:     "infinite header value any negative",
			original: &ReadOptions{MaxHeaderBytes: -1234},
			want:     &ReadOptions{MaxHeaderBytes: math.MaxInt64},
			wantErr:  false,
		},
		{
			name:     "custom header value",
			original: &ReadOptions{MaxHeaderBytes: 128},
			want:     &ReadOptions{MaxHeaderBytes: 128},
			wantErr:  true,
		},
	}
	for _, test := range tests {

		raw := "Subject: " + strings.Repeat("A", 4096*1024) + "\r\n" +
			"\r\n" +
			"This header is very big, but we should allow it via options.\r\n"

		t.Run(test.name, func(t *testing.T) {

			// First validate the options will be set as expected, or there is no
			// point checking the ReadWithOptions func.
			got := test.original.withDefaults()
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ReadOptions.withDefaults() =\n%#v\nbut want:\n%#v", got, test.want)
			}

			_, err := ReadWithOptions(strings.NewReader(raw), test.original)
			gotErr := err != nil

			if gotErr != test.wantErr {
				t.Errorf("ReadWithOptions() = %t but want: %t", gotErr, test.wantErr)
			}
		})
	}
}

func TestReadWithOptions_nilDefault(t *testing.T) {
	raw := "Subject: Something\r\n"
	var opts *ReadOptions
	opts = nil
	_, err := ReadWithOptions(strings.NewReader(raw), opts)
	if err != nil {
		t.Fatalf("ReadWithOptions() = %v", err)
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

	expected := "Mime-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
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

	expected := "Mime-Version: 1.0\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
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

var testCharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
	enc, err := ianaindex.MIME.Encoding(charset)
	if err != nil {
		return nil, err
	}
	return enc.NewDecoder().Reader(input), nil
}

var testCharsetWriter = func(charset string, writer io.Writer) (io.Writer, error) {
	enc, err := ianaindex.MIME.Encoding(charset)
	if err != nil {
		return nil, err
	}
	return enc.NewEncoder().Writer(writer), nil
}

// Returns a func that should be called at the end of the test
func testSetupCharsetReaderWriter() func() {
	oldCharsetReaderValue := CharsetReader
	oldCharsetWriterValue := CharsetWriter
	CharsetReader = testCharsetReader
	CharsetWriter = testCharsetWriter
	return func() {
		CharsetReader = oldCharsetReaderValue
		CharsetWriter = oldCharsetWriterValue
	}
}

// Test going from windows-1252 base64 to windows-1252 quoted-printable
func TestEntity_WriteTo_charset_convert_transfer(t *testing.T) {
	resetCharset := testSetupCharsetReaderWriter()
	defer resetCharset()

	var h Header
	h.Set("Content-Type", "text/plain; charset=windows-1252")
	h.Set("Content-Transfer-Encoding", "base64")
	// "quoted é €"
	// 71 75 6F 74 65 64 20 E9 20 80
	r := strings.NewReader("cXVvdGVkIOkggA==")
	e, _ := New(h, r)

	e.Header.Set("Content-Transfer-Encoding", "quoted-printable")

	var b bytes.Buffer
	if err := e.WriteTo(&b); err != nil {
		t.Fatal("Expected no error while writing entity, got", err)
	}

	expected := "Mime-Version: 1.0\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
		"Content-Type: text/plain; charset=windows-1252\r\n" +
		"\r\n" +
		"quoted =E9 =80"

	if s := b.String(); s != expected {
		t.Errorf("Expected written entity to be:\n%s\nbut got:\n%s", expected, s)
	}
}

// Test going from windows-1252 base64 to utf-8 quoted-printable
func TestEntity_WriteTo_convert_charset_transfer(t *testing.T) {
	resetCharset := testSetupCharsetReaderWriter()
	defer resetCharset()

	var h Header
	h.Set("Content-Type", "text/plain; charset=windows-1252")
	h.Set("Content-Transfer-Encoding", "base64")
	// "quoted é €"
	// 71 75 6F 74 65 64 20 E9 20 80
	r := strings.NewReader("cXVvdGVkIOkggA==")
	e, _ := New(h, r)

	h.Set("Content-Transfer-Encoding", "quoted-printable")
	e.Header.Set("Content-Type", "text/plain; charset=utf-8")

	var b bytes.Buffer
	if err := e.WriteTo(&b); err != nil {
		t.Fatal("Expected no error while writing entity, got", err)
	}

	expected := "Mime-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		"quoted =C3=A9 =E2=82=AC"

	if s := b.String(); s != expected {
		t.Errorf("Expected written entity to be:\n%s\nbut got:\n%s", expected, s)
	}
}

// Test going from utf-8 to windows-1252 with unsupported runes
func TestEntity_WriteTo_invalid_charset(t *testing.T) {
	resetCharset := testSetupCharsetReaderWriter()
	defer resetCharset()

	var h Header
	h.Set("Content-Type", "text/plain; charset=utf-8")
	r := strings.NewReader("non-ascii çhars © ζ Ψ ⊆ ‰")
	e, _ := New(h, r)

	h.Set("Content-Transfer-Encoding", "quoted-printable")
	e.Header.Set("Content-Type", "text/plain; charset=windows-1252")

	var b bytes.Buffer
	err := e.WriteTo(&b)
	if err == nil {
		t.Fatal("New(encoding unsupported rune): expected an error")
	}
	if IsUnknownEncoding(err) {
		t.Fatal("New(encoding unsupported rune): expected an error that does not verify IsUnknownEncoding")
	}
	if !strings.Contains(err.Error(), "rune not supported") {
		t.Fatal("New(encoding unsupported rune): expected 'rune not supported by encoding' error, got", err)
	}
}

func TestEntity_WriteTo_multipart_charset(t *testing.T) {
	resetCharset := testSetupCharsetReaderWriter()
	defer resetCharset()
	e := testMakeMultipartVaryingCharset()

	var b bytes.Buffer
	if err := e.WriteTo(&b); err != nil {
		t.Fatal("Expected no error while writing entity, got", err)
	}

	if s := b.String(); s != testMultipartVaryingCharsetText {
		t.Errorf("Expected written entity to be:\n%s\nbut got:\n%s", testMultipartVaryingCharsetText, s)
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
	if !IsUnknownEncoding(err) {
		t.Fatal("New(unknown transfer encoding): expected an error that verifies IsUnknownEncoding")
	}
	if !errors.As(err, &UnknownEncodingError{}) {
		t.Fatal("New(unknown transfer encoding): expected an error that verifies errors.As(err, &EncodingError{})")
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

// Checks that we are compatible both with lines longer than 72 octets and
// FWS indented lines - per RFC-2045 whitespace should be ignored.
func TestNew_paddedBase64(t *testing.T) {

	testPartRaw := "Content-Type: text/plain; name=\"test.txt\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"Content-ID: <1234567@example.com>\r\n" +
		"Content-Disposition: attachment; filename=\"text.txt\"\r\n" +
		"\r\n" +
		"TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdCwgc2VkIGRvIGVpdXNtb2QgdGVtc\r\n" +
		" G9yIGluY2lkaWR1bnQgdXQgbGFib3JlIGV0IGRvbG9yZSBtYWduYSBhbGlxdWEuIFV0IGVuaW0gYWQgbWluaW0gdmVuaWFtLCBxd\r\n" +
		" WlzIG5vc3RydWQgZXhlcmNpdGF0aW9uIHVsbGFtY28gbGFib3JpcyBuaXNpIHV0IGFsaXF1aXAgZXggZWEgY29tbW9kbyBjb25zZ\r\n" +
		" XF1YXQuIER1aXMgYXV0ZSBpcnVyZSBkb2xvciBpbiByZXByZWhlbmRlcml0IGluIHZvbHVwdGF0ZSB2ZWxpdCBlc3NlIGNpbGx1b\r\n" +
		" SBkb2xvcmUgZXUgZnVnaWF0IG51bGxhIHBhcmlhdHVyLiBFeGNlcHRldXIgc2ludCBvY2NhZWNhdCBjdXBpZGF0YXQgbm9uIHByb\r\n" +
		" 2lkZW50LCBzdW50IGluIGN1bHBhIHF1aSBvZmZpY2lhIGRlc2VydW50IG1vbGxpdCBhbmltIGlkIGVzdCBsYWJvcnVtLg=="

	expected := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed" +
		" do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut e" +
		"nim ad minim veniam, quis nostrud exercitation ullamco laboris nisi " +
		"ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehe" +
		"nderit in voluptate velit esse cillum dolore eu fugiat nulla pariatu" +
		"r. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui" +
		" officia deserunt mollit anim id est laborum."

	e, err := Read(strings.NewReader(testPartRaw))
	if err != nil {
		t.Fatal("New(padded Base64): expected no error, got", err)
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

type testWalkPart struct {
	path      []int
	mediaType string
	body      string
	err       error
}

func walkCollect(e *Entity) ([]testWalkPart, error) {
	var l []testWalkPart
	err := e.Walk(func(path []int, part *Entity, err error) error {
		var body string
		if part.MultipartReader() == nil {
			b, err := ioutil.ReadAll(part.Body)
			if err != nil {
				return err
			}
			body = string(b)
		}
		mediaType, _, _ := part.Header.ContentType()
		l = append(l, testWalkPart{
			path:      path,
			mediaType: mediaType,
			body:      body,
			err:       err,
		})
		return nil
	})
	return l, err
}

func TestWalk_single(t *testing.T) {
	e, err := Read(strings.NewReader(testSingleText))
	if err != nil {
		t.Fatalf("Read() = %v", err)
	}

	want := []testWalkPart{{
		path:      nil,
		mediaType: "text/plain",
		body:      "Message body",
	}}

	got, err := walkCollect(e)
	if err != nil {
		t.Fatalf("Entity.Walk() = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Entity.Walk() =\n%#v\nbut want:\n%#v", got, want)
	}
}

func TestWalk_multipart(t *testing.T) {
	e := testMakeMultipart()

	want := []testWalkPart{
		{
			path:      nil,
			mediaType: "multipart/alternative",
		},
		{
			path:      []int{0},
			mediaType: "text/plain",
			body:      "Text part",
		},
		{
			path:      []int{1},
			mediaType: "text/html",
			body:      "<p>HTML part</p>",
		},
	}

	got, err := walkCollect(e)
	if err != nil {
		t.Fatalf("Entity.Walk() = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Entity.Walk() =\n%#v\nbut want:\n%#v", got, want)
	}
}
