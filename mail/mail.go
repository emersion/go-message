// Package mail implements reading and writing mail messages.
//
// This package assumes that a mail message contains one or more text parts and
// zero or more attachment parts. Each text part represents a different version
// of the message content (e.g. a different type, a different language and so
// on).
//
// RFC 5322 defines the Internet Message Format.
package mail

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/martinlindhe/base36"
)

// Generates an RFC 2822-compliant Message-Id based on the informational draft
// "Recommendations for generating Message IDs", for lack of a better
// authoritative source.
func GenerateMessageID() string {
	var (
		now   bytes.Buffer
		nonce []byte = make([]byte, 8)
	)
	binary.Write(&now, binary.BigEndian, time.Now().UnixNano())
	rand.Read(nonce)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return fmt.Sprintf("<%s.%s@%s>",
		base36.EncodeBytes(now.Bytes()),
		base36.EncodeBytes(nonce),
		hostname)
}
