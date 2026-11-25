package mail_test

import (
	"net/mail"
	"reflect"
	"testing"

	gomail "github.com/emersion/go-message/mail"
)

func TestParseAddressList(t *testing.T) {
	want := []*mail.Address{{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"},
		{"Han Solo", "hanibunny@example.org"},
	}
	input := "Mitsuha Miyamizu <mitsuha.miyamizu@example.org>,Han Solo <hanibunny@example.org>"
	if got, err := mail.ParseAddressList(input); err != nil {
		t.Error("Expected no error while parsing address list got:", err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected address list to be %v, but got %v", want, got)
	}
}

func TestParseAddress(t *testing.T) {
	want := &mail.Address{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"}
	input := "Mitsuha Miyamizu <mitsuha.miyamizu@example.org>"

	if got, err := mail.ParseAddress(input); err != nil {
		t.Error("Expected no error while parsing address got:", err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected address to be %v, but got %v", want, got)
	}
}

func TestParseAddressUnquotedAtInDisplayName(t *testing.T) {
	want := &mail.Address{"name@yahoo.com", "name@yahoo.com"}

	if got, err := gomail.ParseAddress("name@yahoo.com <name@yahoo.com>"); err != nil {
		t.Error("Expected no error while parsing address got:", err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected address to be %v, but got %v", want, got)
	}
}

func TestParseAddressListUnquotedAtInDisplayName(t *testing.T) {
	want := []*mail.Address{
		{"name@yahoo.com", "name@yahoo.com"},
		{"other@example.com", "other@example.com"},
	}
	input := "name@yahoo.com <name@yahoo.com>, other@example.com <other@example.com>"

	if got, err := gomail.ParseAddressList(input); err != nil {
		t.Error("Expected no error while parsing address list got:", err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected address list to be %v, but got %v", want, got)
	}
}

func TestParseAddressAlreadyQuotedAtInDisplayName(t *testing.T) {
	want := &mail.Address{"name@yahoo.com", "name@yahoo.com"}

	if got, err := gomail.ParseAddress(`"name@yahoo.com" <name@yahoo.com>`); err != nil {
		t.Error("Expected no error while parsing address got:", err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected address to be %v, but got %v", want, got)
	}
}
