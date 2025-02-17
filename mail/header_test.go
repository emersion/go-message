package mail_test

import (
	"bufio"
	"bytes"
	netmail "net/mail"
	"net/url"
	"reflect"
	"strings"
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

func TestHeader_Date_CFWS(t *testing.T) {
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

func TestHeader_Date_empty(t *testing.T) {
	var h mail.Header
	date, err := h.Date()
	if err != nil {
		t.Errorf("Date() = %v", err)
	} else if !date.IsZero() {
		t.Errorf("Date() = %v, want time.Time{}", date)
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
		{
			"<123@[2001:db8:85a3:8d3:1319:8a2e:370:7348]>",
			"123@[2001:db8:85a3:8d3:1319:8a2e:370:7348]",
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

func TestHeader_GenerateMessageID(t *testing.T) {
	var h mail.Header
	if err := h.GenerateMessageID(); err != nil {
		t.Fatalf("Header.GenerateMessageID() = %v", err)
	}
	if _, err := h.MessageID(); err != nil {
		t.Errorf("Failed to parse generated Message-Id: Header.MessageID() = %v", err)
	}
}

func TestHeader_SetMsgIDList(t *testing.T) {
	tests := []struct {
		raw    string
		msgIDs []string
	}{
		{"", nil},
		{"<123@asdf>", []string{"123@asdf"}},
		{"<123@asdf> <456@asdf>", []string{"123@asdf", "456@asdf"}},
	}
	for _, test := range tests {
		var h mail.Header
		h.SetMsgIDList("In-Reply-To", test.msgIDs)
		raw := h.Get("In-Reply-To")
		if raw != test.raw {
			t.Errorf("Failed to format In-Reply-To %q: Header.Get() = %q, want %q", test.msgIDs, raw, test.raw)
		}
	}
}

func TestHeader_CanUseNetMailAddress(t *testing.T) {
	netfrom := []*netmail.Address{{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"}}
	mailfrom := []*mail.Address{{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"}}

	//sanity check that they types are identical
	if !reflect.DeepEqual(netfrom, mailfrom) {
		t.Error("[]*net/mail.Address differs from []*mail.Address")
	}

	//roundtrip
	var h mail.Header
	h.SetAddressList("From", netfrom)
	if got, err := h.AddressList("From"); err != nil {
		t.Error("Expected no error while parsing header address list, got:", err)
	} else if !reflect.DeepEqual(got, netfrom) {
		t.Errorf("Expected header address list to be %v, but got %v", netfrom, got)
	}
}

func TestHeader_EmptyAddressList(t *testing.T) {
	tests := []struct {
		key   string
		list  []*mail.Address
		unset bool
	}{
		{"cc", nil, false},
		{"to", []*mail.Address{}, false},
		{"cc", []*mail.Address{{"Mitsuha Miyamizu", "mitsuha.miyamizu@example.org"}}, true},
	}

	for _, test := range tests {
		var h mail.Header
		h.SetAddressList(test.key, test.list)
		if test.unset {
			h.SetAddressList(test.key, nil)
		}
		buf := bytes.NewBuffer(nil)
		w, err := mail.CreateSingleInlineWriter(buf, h)
		if err != nil {
			t.Error("Expected no error while creating inline writer, got:", err)
		}
		if err := w.Close(); err != nil {
			t.Error("Expected no error while closing inline writer, got:", err)
		}
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			line := strings.ToLower(scanner.Text())
			if strings.HasPrefix(line, test.key) {
				t.Error("Expected no address list header field, but got:", scanner.Text())
			}
		}
	}
}

func TestHeader_ListCommandURLList(t *testing.T) {
	tests := []struct {
		header string
		raw    string
		urls   []*url.URL
		xfail  bool
	}{
		{
			header: "List-Help",
			raw:    "<mailto:hello@example.com",
			xfail:  true,
		},
		// These tests might seem repetitive, but they are the examples given at
		// https://www.rfc-editor.org/rfc/rfc2369.
		{
			header: "List-Help",
			raw:    "<mailto:list@host.com?subject=help> (List Instructions)",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list@host.com", RawQuery: "subject=help"},
			},
		},
		{
			header: "List-Help",
			raw:    "<mailto:list-manager@host.com?body=info>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list-manager@host.com", RawQuery: "body=info"},
			},
		},
		{
			header: "List-Help",
			raw:    "<mailto:list-info@host.com> (Info about the list)",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list-info@host.com"},
			},
		},
		{
			header: "List-Help",
			raw:    "<http://www.host.com/list/>, <mailto:list-info@host.com>",
			urls: []*url.URL{
				{Scheme: "http", Host: "www.host.com", Path: "/list/"},
				{Scheme: "mailto", Opaque: "list-info@host.com"},
			},
		},
		{
			header: "List-Help",
			raw:    "<ftp://ftp.host.com/list.txt> (FTP),\r\n\t<mailto:list@host.com?subject=help>",
			urls: []*url.URL{
				{Scheme: "ftp", Host: "ftp.host.com", Path: "/list.txt"},
				{Scheme: "mailto", Opaque: "list@host.com", RawQuery: "subject=help"},
			},
		},
		{
			header: "List-Unsubscribe",
			raw:    "<mailto:list@host.com?subject=unsubscribe>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list@host.com", RawQuery: "subject=unsubscribe"},
			},
		},
		{
			header: "List-Unsubscribe",
			raw:    "(Use this command to get off the list)\r\n\t<mailto:list-manager@host.com?body=unsubscribe%20list>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list-manager@host.com", RawQuery: "body=unsubscribe%20list"},
			},
		},
		{
			header: "List-Unsubscribe",
			raw:    "<mailto:list-off@host.com>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list-off@host.com"},
			},
		},
		{
			header: "List-Unsubscribe",
			raw:    "<http://www.host.com/list.cgi?cmd=unsub&lst=list>,\r\n\t<mailto:list-request@host.com?subject=unsubscribe>",
			urls: []*url.URL{
				{Scheme: "http", Host: "www.host.com", Path: "/list.cgi", RawQuery: "cmd=unsub&lst=list"},
				{Scheme: "mailto", Opaque: "list-request@host.com", RawQuery: "subject=unsubscribe"},
			},
		},
		{
			header: "List-Subscribe",
			raw:    "<mailto:list@host.com?subject=subscribe>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list@host.com", RawQuery: "subject=subscribe"},
			},
		},
		{
			header: "List-Subscribe",
			raw:    "(Use this command to join the list)\r\n\t<mailto:list-manager@host.com?body=subscribe%20list>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list-manager@host.com", RawQuery: "body=subscribe%20list"},
			},
		},
		{
			header: "List-Unsubscribe",
			raw:    "<mailto:list-on@host.com>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list-on@host.com"},
			},
		},
		{
			header: "List-Subscribe",
			raw:    "<http://www.host.com/list.cgi?cmd=sub&lst=list>,\r\n\t<mailto:list-manager@host.com?subject=subscribe>",
			urls: []*url.URL{
				{Scheme: "http", Host: "www.host.com", Path: "/list.cgi", RawQuery: "cmd=sub&lst=list"},
				{Scheme: "mailto", Opaque: "list-manager@host.com", RawQuery: "subject=subscribe"},
			},
		},
		{
			header: "List-Post",
			raw:    "<mailto:list@host.com>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list@host.com"},
			},
		},
		{
			header: "List-Post",
			raw:    "<mailto:moderator@host.com> (Postings are Moderated)",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "moderator@host.com"},
			},
		},
		{
			header: "List-Post",
			raw:    "<mailto:moderator@host.com?subject=list%20posting>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "moderator@host.com", RawQuery: "subject=list%20posting"},
			},
		},
		{
			header: "List-Post",
			raw:    "NO (posting not allowed on this list)",
			urls:   []*url.URL{nil},
		},
		{
			header: "List-Owner",
			raw:    "<mailto:listmom@host.com> (Contact Person for Help)",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "listmom@host.com"},
			},
		},
		{
			header: "List-Owner",
			raw:    "<mailto:grant@foo.bar> (Grant Neufeld)",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "grant@foo.bar"},
			},
		},
		{
			header: "List-Owner",
			raw:    "<mailto:josh@foo.bar?Subject=list>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "josh@foo.bar", RawQuery: "Subject=list"},
			},
		},
		{
			header: "List-Archive",
			raw:    "<mailto:archive@host.com?subject=index%20list>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "archive@host.com", RawQuery: "subject=index%20list"},
			},
		},
		{
			header: "List-Archive",
			raw:    "<ftp://ftp.host.com/pub/list/archive/>",
			urls: []*url.URL{
				{Scheme: "ftp", Host: "ftp.host.com", Path: "/pub/list/archive/"},
			},
		},
		{
			header: "List-Archive",
			raw:    "<http://www.host.com/list/archive/> (Web Archive)",
			urls: []*url.URL{
				{Scheme: "http", Host: "www.host.com", Path: "/list/archive/"},
			},
		},
	}

	for _, test := range tests {
		var h mail.Header
		h.Set(test.header, test.raw)

		urls, err := h.ListCommandURLList(test.header)
		if err != nil && !test.xfail {
			t.Errorf("Failed to parse %s %q: Header.ListCommandURLList() = %v", test.header, test.raw, err)
		} else if !reflect.DeepEqual(urls, test.urls) {
			t.Errorf("Failed to parse %s %q: Header.ListCommandURLList() = %q, want %q", test.header, test.raw, urls, test.urls)
		}
	}
}

func TestHeader_SetListCommandURLList(t *testing.T) {
	tests := []struct {
		raw  string
		urls []*url.URL
	}{
		{
			raw: "<mailto:list@example.com>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list@example.com"},
			},
		},
		{
			raw: "<mailto:list@example.com>, <https://example.com:8080>",
			urls: []*url.URL{
				{Scheme: "mailto", Opaque: "list@example.com"},
				{Scheme: "https", Host: "example.com:8080"},
			},
		},
	}
	for _, test := range tests {
		var h mail.Header
		h.SetListCommandURLList("List-Post", test.urls)
		raw := h.Get("List-Post")
		if raw != test.raw {
			t.Errorf("Failed to format List-Post %q: Header.Get() = %q, want %q", test.urls, raw, test.raw)
		}
	}
}
