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

func TestHeader_MessageID(t *testing.T) {
	tests := []struct {
		raw   string
		msgID string
	}{
		{"", ""},
		{"<123@asdf>", "123@asdf"},
		{
			"  \t <DM6PR09MB253761A38B42C713082A7CE2C60C0@DM6PR09MB2537.namprd09.prod.outlook.com>",
			"DM6PR09MB253761A38B42C713082A7CE2C60C0@DM6PR09MB2537.namprd09.prod.outlook.com",
		},
		{
			`<20200122161125.7enac4n5rsxfnhg7@example.com> (Christopher Wellons's message of "Wed, 22 Jan 2020 11:11:25 -0500")`,
			"20200122161125.7enac4n5rsxfnhg7@example.com",
		},
	}
	for _, test := range tests {
		var h mail.Header
		h.Set("Message-ID", test.raw)
		msgID, err := h.MessageID()
		if err != nil {
			t.Errorf("Failed to parse Message-ID %q: Header.MessageID() = %v", test.raw, err)
		} else if msgID != test.msgID {
			t.Errorf("Failed to parse Message-ID %q: Header.MessageID() = %q, want %q", test.raw, msgID, test.msgID)
		}
	}
}

func TestHeader_MsgIDList(t *testing.T) {
	tests := []struct {
		raw    string
		msgIDs []string
	}{
		{"", nil},
		{"<123@asdf>", []string{"123@asdf"}},
		{
			"  \t <DM6PR09MB253761A38B42C713082A7CE2C60C0@DM6PR09MB2537.namprd09.prod.outlook.com>",
			[]string{"DM6PR09MB253761A38B42C713082A7CE2C60C0@DM6PR09MB2537.namprd09.prod.outlook.com"},
		},
		{
			`<20200122161125.7enac4n5rsxfnhg7@example.com> (Christopher Wellons's message of "Wed, 22 Jan 2020 11:11:25 -0500")`,
			[]string{"20200122161125.7enac4n5rsxfnhg7@example.com"},
		},
		{
			"<87pnfb69f3.fsf@bernat.ch>  \t <20200122161125.7enac4n5rsxfnhg7@nullprogram.com>",
			[]string{"87pnfb69f3.fsf@bernat.ch", "20200122161125.7enac4n5rsxfnhg7@nullprogram.com"},
		},
		{
			"<87pnfb69f3.fsf@bernat.ch> (a comment) \t <20200122161125.7enac4n5rsxfnhg7@nullprogram.com> (another comment)",
			[]string{"87pnfb69f3.fsf@bernat.ch", "20200122161125.7enac4n5rsxfnhg7@nullprogram.com"},
		},
	}
	for _, test := range tests {
		var h mail.Header
		h.Set("In-Reply-To", test.raw)
		msgIDs, err := h.MsgIDList("In-Reply-To")
		if err != nil {
			t.Errorf("Failed to parse In-Reply-To %q: Header.MsgIDList() = %v", test.raw, err)
		} else if !reflect.DeepEqual(msgIDs, test.msgIDs) {
			t.Errorf("Failed to parse In-Reply-To %q: Header.MsgIDList() = %q, want %q", test.raw, msgIDs, test.msgIDs)
		}
	}
}
