package message

import (
	"bufio"
	"bytes"
	"io"
	"mime/quotedprintable"
)

const (
	// defaultBufSize defines the buffer size used by `mime/quotedprintable`.
	defaultBufSize = 4096

	// maxQuotedPrintableLineLength defines the longest line accepted. Longer
	// lines are splitted by inserting soft breakes.
	// By RFC is allowed line with 76 characters long only. Some messages in
	// the wild do not respect this rule. `mime/quotedprintable` can support
	// up to bufio buffer, which is 4096 bytes. Longer lines than what RFC
	// defines but shorter than bufio buffer are kept without modification
	// to not modify message when not neccessary.
	// Note that soft brake has to continue with line ending, i.e., `Read`
	// operation cannot fill the whole quotedprintable buffer with the line
	// without line ending. That is reported as error too. That's why this
	// lenght is two bytes (to include \r\n) shorter then the buffer.
	maxQuotedPrintableLineLength = defaultBufSize - 3
)

type quotedPrintableLongLinesReader struct {
	r   *bufio.Reader
	buf bytes.Buffer
}

func newQuotedPrintableLongLinesReader(r io.Reader) io.Reader {
	return quotedprintable.NewReader(&quotedPrintableLongLinesReader{
		r: bufio.NewReaderSize(r, maxQuotedPrintableLineLength),
	})
}

func (l *quotedPrintableLongLinesReader) Read(p []byte) (int, error) {
	if l.buf.Len() == 0 {
		line, err := l.r.ReadSlice('\n')
		// ErrBufferFull is returned once the line is not complete.
		if len(line) == 0 && err != nil && err != bufio.ErrBufferFull {
			return 0, err
		}

		if err != bufio.ErrBufferFull {
			l.buf.Write(line)
		} else {
			length := len(line)
			suffix := 0
			if line[length-1] == byte('=') {
				suffix = 1
			} else if line[length-2] == byte('=') {
				suffix = 2
			}

			l.buf.Write(line[:length-suffix])
			l.buf.Write([]byte("=\r\n"))
			if suffix > 0 {
				l.buf.Write(line[length-suffix:])
			}
		}
	}

	return l.buf.Read(p)
}
