package mail_test

import (
	"regexp"
	"testing"

	"github.com/emersion/go-message/mail"
)

const mailString = "Subject: Your Name\r\n" +
	"Content-Type: multipart/mixed; boundary=message-boundary\r\n" +
	"\r\n" +
	"--message-boundary\r\n" +
	"Content-Type: multipart/alternative; boundary=text-boundary\r\n" +
	"\r\n" +
	"--text-boundary\r\n" +
	"Content-Type: text/plain\r\n" +
	"Content-Disposition: inline\r\n" +
	"\r\n" +
	"Who are you?\r\n" +
	"--text-boundary--\r\n" +
	"--message-boundary\r\n" +
	"Content-Type: text/plain\r\n" +
	"Content-Disposition: attachment; filename=note.txt\r\n" +
	"\r\n" +
	"I'm Mitsuha.\r\n" +
	"--message-boundary--\r\n"

const nestedMailString = "Subject: Fwd: Your Name\r\n" +
	"Content-Type: multipart/mixed; boundary=outer-message-boundary\r\n" +
	"\r\n" +
	"--outer-message-boundary\r\n" +
	"Content-Type: text/plain\r\n" +
	"Content-Disposition: inline\r\n" +
	"\r\n" +
	"I forgot.\r\n" +
	"--outer-message-boundary\r\n" +
	"Content-Type: message/rfc822\r\n" +
	"Content-Disposition: attachment; filename=attached-message.eml\r\n" +
	"\r\n" +
	mailString +
	"--outer-message-boundary--\r\n"

func TestGenerateMessageID(t *testing.T) {
	msgId := mail.GenerateMessageID()
	regex := regexp.MustCompile(`^<.*@.*>$`)
	if !regex.MatchString(msgId) {
		t.Error("Generated message ID does not meet RFC requirement")
	}
}
