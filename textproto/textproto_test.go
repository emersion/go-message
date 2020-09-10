package textproto

import (
	"fmt"

)

func ExampleHeader() {
	var h Header
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
