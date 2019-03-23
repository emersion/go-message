package message

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

var from = "Mitsuha Miyamizu <mitsuha.miyamizu@example.com>"
var to = "Taki Tachibana <taki.tachibana@example.org>"
var received2 = "from example.com by example.org"

func newTestHeader() Header2 {
	var h Header2
	h.Add("From", from)
	h.Add("To", to)
	h.Add("Received", "from localhost by example.com")
	h.Add("Received", received2)
	return h
}

func collectHeaderFields(fields HeaderFields) []string {
	var l []string
	for fields.Next() {
		l = append(l, fields.Key() + ": " + fields.Value())
	}
	return l
}

func TestHeader2(t *testing.T) {
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

func TestHeader2_Set(t *testing.T) {
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

func TestHeader2_Del(t *testing.T) {
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

func TestHeader2_Fields_Del(t *testing.T) {
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
	want := []string{"Received: from example.com by example.org"}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("FieldsByKey(\"Received\") reported incorrect values after HeaderFields.Del(): got \n%#v\n but want \n%#v", l, want)
	}
}

func TestHeader2_FieldsByKey_Del(t *testing.T) {
	h := newTestHeader()

	fields := h.FieldsByKey("Received")
	if !fields.Next() {
		t.Fatal("FieldsByKey(\"Received\").Next() = false, want true")
	}
	fields.Del()

	l := collectHeaderFields(h.FieldsByKey("Received"))
	want := []string{"Received: from example.com by example.org"}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("FieldsByKey(\"Received\") reported incorrect values after HeaderFields.Del(): got \n%#v\n but want \n%#v", l, want)
	}
}

const testHeader = `Received: from example.com by example.org
Received: from localhost by example.com
To: Taki Tachibana <taki.tachibana@example.org>
From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>

`

func TestReadHeader(t *testing.T) {
	h, err := readHeader(bufio.NewReader(strings.NewReader(testHeader)))
	if err != nil {
		t.Fatalf("readHeader() returned error %v", err)
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
}
