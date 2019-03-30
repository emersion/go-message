package textproto_test

import (
	"fmt"

	"github.com/emersion/go-message/textproto"
)

func ExampleHeader() {
	var h textproto.Header
	h.Add("From", "<root@nsa.gov>")
	h.Add("To", "<root@gchq.gov.uk>")
	h.Set("Subject", "Tonight's dinner")

	fmt.Println("From: ", h.Get("From"))
	fmt.Println("Has Received: ", h.Has("Received"))

	fmt.Println("Header fields:")
	fields := h.Fields()
	for fields.Next() {
		fmt.Println("  ", fields.Key())
	}
}
