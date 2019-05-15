package textproto

import (
	"bufio"
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

var from = "Mitsuha Miyamizu <mitsuha.miyamizu@example.com>"
var to = "Taki Tachibana <taki.tachibana@example.org>"
var received2 = "from example.com by example.org"

func newTestHeader() Header {
	var h Header
	h.Add("From", from)
	h.Add("To", to)
	h.Add("Received", "from localhost by example.com")
	h.Add("Received", received2)
	return h
}

func collectHeaderFields(fields HeaderFields) []string {
	var l []string
	for fields.Next() {
		l = append(l, fields.Key()+": "+fields.Value())
	}
	return l
}

func TestHeader(t *testing.T) {
	h := newTestHeader()

	if got := h.Get("From"); got != from {
		t.Errorf("Get(\"From\") = %#v, want %#v", got, from)
	}
	if got := h.Get("Received"); got != received2 {
		t.Errorf("Get(\"Received\") = %#v, want %#v", got, received2)
	}
	if got := h.Get("X-I-Dont-Exist"); got != "" {
		t.Errorf("Get(non-existing) = %#v, want \"\"", got)
	}

	if !h.Has("From") {
		t.Errorf("Has(\"From\") = false, want true")
	}
	if h.Has("X-I-Dont-Exist") {
		t.Errorf("Has(non-existing) = true, want false")
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Received: from example.com by example.org",
		"Received: from localhost by example.com",
		"To: Taki Tachibana <taki.tachibana@example.org>",
		"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	l = collectHeaderFields(h.FieldsByKey("Received"))
	want = []string{
		"Received: from example.com by example.org",
		"Received: from localhost by example.com",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("FieldsByKey(\"Received\") reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	if h.FieldsByKey("X-I-Dont-Exist").Next() {
		t.Errorf("FieldsByKey(non-existing).Next() returned true, want false")
	}
}

func TestHeader_Set(t *testing.T) {
	h := newTestHeader()

	h.Set("From", to)
	if got := h.Get("From"); got != to {
		t.Errorf("Get(\"From\") = %#v after Set(), want %#v", got, to)
	}
	l := collectHeaderFields(h.FieldsByKey("From"))
	want := []string{"From: Taki Tachibana <taki.tachibana@example.org>"}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("FieldsByKey(\"From\") reported incorrect values after Set(): got \n%#v\n but want \n%#v", l, want)
	}
}

func TestHeader_Del(t *testing.T) {
	h := newTestHeader()

	h.Del("Received")
	if h.Has("Received") {
		t.Errorf("Has(\"Received\") = true after Del(), want false")
	}
	l := collectHeaderFields(h.FieldsByKey("Received"))
	var want []string = nil
	if !reflect.DeepEqual(l, want) {
		t.Errorf("FieldsByKey(\"Received\") reported incorrect values after Del(): got \n%#v\n but want \n%#v", l, want)
	}
}

func TestHeader_Fields_Del_multiple(t *testing.T) {
	h := newTestHeader()

	ok := false
	fields := h.Fields()
	for fields.Next() {
		if fields.Key() == "Received" {
			fields.Del()
			ok = true
			break
		}
	}
	if !ok {
		t.Fatal("Fields() didn't yield \"Received\"")
	}

	l := collectHeaderFields(h.FieldsByKey("Received"))
	want := []string{"Received: from localhost by example.com"}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("FieldsByKey(\"Received\") reported incorrect values after HeaderFields.Del(): got \n%#v\n but want \n%#v", l, want)
	}
}

func TestHeader_Fields_Del_single(t *testing.T) {
	h := newTestHeader()

	ok := false
	fields := h.Fields()
	for fields.Next() {
		if fields.Key() == "To" {
			fields.Del()
			ok = true
			break
		}
	}
	if !ok {
		t.Fatal("Fields() didn't yield \"To\"")
	}

	if h.FieldsByKey("To").Next() {
		t.Errorf("FieldsByKey(\"To\") returned a non-empty set")
	}
}

func TestHeader_Fields_Del_all(t *testing.T) {
	h := newTestHeader()

	fields := h.Fields()
	for fields.Next() {
		fields.Del()
	}

	if h.Fields().Next() {
		t.Errorf("Fields() returned a non-empty set")
	}
}

func TestHeader_FieldsByKey_Del(t *testing.T) {
	h := newTestHeader()

	fields := h.FieldsByKey("Received")
	if !fields.Next() {
		t.Fatal("FieldsByKey(\"Received\").Next() = false, want true")
	}
	fields.Del()

	l := collectHeaderFields(h.FieldsByKey("Received"))
	want := []string{"Received: from localhost by example.com"}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("FieldsByKey(\"Received\") reported incorrect values after HeaderFields.Del(): got \n%#v\n but want \n%#v", l, want)
	}
}

const testHeader = "Received: from example.com by example.org\r\n" +
	"Received: from localhost by example.com\r\n" +
	"To: Taki Tachibana <taki.tachibana@example.org>\r\n" +
	"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>\r\n\r\n"

func TestReadHeader(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(testHeader))
	h, err := ReadHeader(r)
	if err != nil {
		t.Fatalf("readHeader() returned error: %v", err)
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Received: from example.com by example.org",
		"Received: from localhost by example.com",
		"To: Taki Tachibana <taki.tachibana@example.org>",
		"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	b := make([]byte, 1)
	if _, err := r.Read(b); err != io.EOF {
		t.Errorf("Read() didn't return EOF: %v", err)
	}
}

const testLFHeader = `From: contact@example.org
To: contact@example.org
Subject: A little message, just for you
Date: Wed, 11 May 2016 14:31:59 +0000
Message-ID: <0000000@localhost/>
Content-Type: text/plain
`

// TODO: make this test pass
func TestReadHeader_lf(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(testLFHeader))
	h, err := ReadHeader(r)
	if err != nil {
		t.Fatalf("readHeader() returned error: %v", err)
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"From: contact@example.org",
		"To: contact@example.org",
		"Subject: A little message, just for you",
		"Date: Wed, 11 May 2016 14:31:59 +0000",
		"Message-ID: <0000000@localhost/>",
		"Content-Type: text/plain",
	}
	if !reflect.DeepEqual(l, want) {
		t.Logf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	b := make([]byte, 1)
	if _, err := r.Read(b); err != io.EOF {
		t.Logf("Read() didn't return EOF: %v", err)
	}
}

func TestWriteHeader(t *testing.T) {
	h := newTestHeader()

	var b bytes.Buffer
	if err := WriteHeader(&b, h); err != nil {
		t.Fatalf("writeHeader() returned error: %v", err)
	}

	if b.String() != testHeader {
		t.Errorf("writeHeader() wrote invalid data: got \n%v\n but want \n%v", b.String(), testHeader)
	}
}

// RFC says key shouldn't have trailing spaces, but those appear in the wild, so
// we need to handle them.
const testHeaderWithWhitespace = "Subject \t : \t Hey \r\n" +
	" \t there\r\n" +
	"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>\r\n\r\n"

func TestHeaderWithWhitespace(t *testing.T) {
	h, err := ReadHeader(bufio.NewReader(strings.NewReader(testHeaderWithWhitespace)))
	if err != nil {
		t.Fatalf("readHeader() returned error: %v", err)
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Subject: Hey there",
		"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	var b bytes.Buffer
	if err := WriteHeader(&b, h); err != nil {
		t.Fatalf("writeHeader() returned error: %v", err)
	}

	if b.String() != testHeaderWithWhitespace {
		t.Errorf("writeHeader() wrote invalid data: got \n%v\n but want \n%v", b.String(), testHeaderWithWhitespace)
	}
}

var formatHeaderFieldTests = []struct {
	k, v      string
	formatted string
}{
	{
		k:         "From",
		v:         "Mitsuha Miyamizu <mitsuha.miyamizu@example.org>",
		formatted: "From: Mitsuha Miyamizu <mitsuha.miyamizu@example.org>\r\n",
	},
	{
		k:         "Subject",
		v:         "This is a very long subject, much longer than just the 76 characters limit that applies to message header fields",
		formatted: "Subject: This is a very long subject, much longer than just the 76\r\n characters limit that applies to message header fields\r\n",
	},
	{
		k:         "Subject",
		v:         "This is        yet          \t  another    subject          \t                   with many         whitespace      characters",
		formatted: "Subject: This is        yet          \t  another    subject          \t       \r\n            with many         whitespace      characters\r\n",
	},
	{
		k:         "Subject",
		v:         "=?utf-8?q?=E2=80=9CDeveloper_reads_customer_requested_change.=E2=80=9D=0A?= =?utf-8?q?=0ACaravaggio=0A=0AOil_on...?=",
		formatted: "Subject: =?utf-8?q?=E2=80=9CDeveloper_reads_customer_requested_change.\r\n =E2=80=9D=0A?= =?utf-8?q?=0ACaravaggio=0A=0AOil_on...?=\r\n",
	},
	{
		k:         "Subject",
		v:         "=?utf-8?q?=E2=80=9CShort subject=E2=80=9D=0A?= =?utf-8?q?=0AAuthor=0A=0AOil_on...?=",
		formatted: "Subject: =?utf-8?q?=E2=80=9CShort subject=E2=80=9D=0A?= =?utf-8?q?\r\n =0AAuthor=0A=0AOil_on...?=\r\n",
	},
	{
		k:         "Subject",
		v:         "=?utf-8?q?=E2=80=9CVery long subject very long subject very long subject very long subject=E2=80=9D=0A?= =?utf-8?q?=0ALong second part of subject long second part of subject long second part of subject long subject=0A=0AOil_on...?=",
		formatted: "Subject: =?utf-8?q?=E2=80=9CVery long subject very long subject very long\r\n subject very long subject=E2=80=9D=0A?= =?utf-8?q?=0ALong second part of\r\n subject long second part of subject long second part of subject long\r\n subject=0A=0AOil_on...?=\r\n",
	},
	{
		k:         "DKIM-Signature",
		v:         "v=1;\r\n h=From:To:Reply-To:Subject:Message-ID:References:In-Reply-To:MIME-Version;\r\n d=example.org\r\n",
		formatted: "Dkim-Signature: v=1;\r\n h=From:To:Reply-To:Subject:Message-ID:References:In-Reply-To:MIME-Version;\r\n d=example.org\r\n",
	},
	{
		k:         "DKIM-Signature",
		v:         "v=1; h=From; d=example.org; b=AuUoFEfDxTDkHlLXSZEpZj79LICEps6eda7W3deTVFOk4yAUoqOB4nujc7YopdG5dWLSdNg6xNAZpOPr+kHxt1IrE+NahM6L/LbvaHutKVdkLLkpVaVVQPzeRDI009SO2Il5Lu7rDNH6mZckBdrIx0orEtZV4bmp/YzhwvcubU4=\r\n",
		formatted: "Dkim-Signature: v=1; h=From; d=example.org;\r\n b=AuUoFEfDxTDkHlLXSZEpZj79LICEps6eda7W3deTVFOk4yAUoqOB4nujc7YopdG5dWLSdNg6x\r\n NAZpOPr+kHxt1IrE+NahM6L/LbvaHutKVdkLLkpVaVVQPzeRDI009SO2Il5Lu7rDNH6mZckBdrI\r\n x0orEtZV4bmp/YzhwvcubU4=\r\n",
	},
	{
		k:         "Bcc",
		v:         "",
		formatted: "Bcc: \r\n",
	},
	{
		k:         "Bcc",
		v:         " ",
		formatted: "Bcc:  \r\n",
	},
}

func TestWriteHeader_continued(t *testing.T) {
	for _, test := range formatHeaderFieldTests {
		var h Header
		h.Add(test.k, test.v)

		var b bytes.Buffer
		if err := WriteHeader(&b, h); err != nil {
			t.Fatalf("writeHeader() returned error: %v", err)
		}
		if b.String() != test.formatted+"\r\n" {
			t.Errorf("Expected formatted header to be \n%v\n but got \n%v", test.formatted+"\r\n", b.String())
		}
	}
}
