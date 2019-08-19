package mail_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/emersion/go-message/mail"
)

func TestHeader(t *testing.T) {
	date := time.Unix(1466253744, 0)
	from := []*mail.Address{{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"}}
	subject := "Café"

	var h mail.Header
	h.SetAddressList("From", from)
	h.SetDate(date)
	h.SetSubject(subject)

	if got, err := h.Date(); err != nil {
		t.Error("Expected no error while parsing header date, got:", err)
	} else if !got.Equal(date) {
		t.Errorf("Expected header date to be %v, but got %v", date, got)
	}

	if got, err := h.AddressList("From"); err != nil {
		t.Error("Expected no error while parsing header address list, got:", err)
	} else if !reflect.DeepEqual(got, from) {
		t.Errorf("Expected header address list to be %v, but got %v", from, got)
	}

	if got, err := h.AddressList("Cc"); err != nil {
		t.Error("Expected no error while parsing missing header address list, got:", err)
	} else if got != nil {
		t.Errorf("Expected missing header address list to be %v, but got %v", nil, got)
	}

	if got, err := h.Subject(); err != nil {
		t.Error("Expected no error while parsing header subject, got:", err)
	} else if got != subject {
		t.Errorf("Expected header subject to be %v, but got %v", subject, got)
	}
}

func TestCFWSDates(t *testing.T) {
	tc := []string{
		"Mon, 22 Jul 2019 13:57:29 -0500 (GMT-05:00)",
		"Mon, 22 Jul 2019 13:57:29 -0500",
		"Mon, 2 Jan 06 15:04:05 MST (Some random stuff)",
		"Mon, 2 Jan 06 15:04:05 MST",
	}
	var h mail.Header
	for _, tt := range tc {
		h.Set("Date", tt)
		_, err := h.Date()
		if err != nil {
			t.Errorf("Failed to parse time %q: %v", tt, err)
		}
	}
}
