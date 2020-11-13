package message

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestQuotedPrintableLongLines(t *testing.T) {
	long := func(ending string) string {
		return strings.Repeat(".", maxQuotedPrintableLineLength-3) + ending
	}

	tests := []struct {
		input, output string
	}{
		{
			"abc",
			"abc",
		},
		{
			long("abcdefghijkloooong"),
			long("abcdefghijkloooong"),
		},
		// Original soft end of line around the wrap position.
		{
			long("=\r\nabcd"),
			long("abcd"),
		},
		{
			long("a=\r\nbcd"),
			long("abcd"),
		},
		{
			long("ab=\r\ncd"),
			long("abcd"),
		},
		{
			long("abc=\r\nd"),
			long("abcd"),
		},
		{
			long("abcd=\r\n"),
			long("abcd"),
		},
		// Encoded = around the wrap position.
		{
			long("=3Dabcd"),
			long("=abcd"),
		},
		{
			long("a=3Dbcd"),
			long("a=bcd"),
		},
		{
			long("ab=3Dcd"),
			long("ab=cd"),
		},
		{
			long("abc=3Dd"),
			long("abc=d"),
		},
		{
			long("abcd=3D"),
			long("abcd="),
		},
		// Encoded é around the wrap position.
		{
			long("=C3=A9abcd"),
			long("éabcd"),
		},
		{
			long("a=C3=A9bcd"),
			long("aébcd"),
		},
		{
			long("ab=C3=A9cd"),
			long("abécd"),
		},
		{
			long("abc=C3=A9d"),
			long("abcéd"),
		},
		{
			long("abcd=C3=A9"),
			long("abcdé"),
		},
		// Very long line with many wraps.
		{
			long(long(long(long(long("abc"))))),
			long(long(long(long(long("abc"))))),
		},
	}
	for idx, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			inputReader := bytes.NewReader([]byte(tc.input + "\r\n"))
			outputReader := newQuotedPrintableLongLinesReader(inputReader)
			output, err := ioutil.ReadAll(outputReader)
			if err != nil {
				t.Error("Expected no error, but got:", err)
			}
			if string(output) != tc.output+"\r\n" {
				t.Errorf("Expected %v but got %v", tc.output, string(output))
			}
		})
	}
}
