package mail_test

const mailString = "Content-Type: multipart/mixed; boundary=message-boundary\r\n" +
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
