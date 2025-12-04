package message

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"
)

type UnknownEncodingError struct {
	e error
}

func (u UnknownEncodingError) Unwrap() error { return u.e }

func (u UnknownEncodingError) Error() string {
	return "encoding error: " + u.e.Error()
}

// IsUnknownEncoding returns a boolean indicating whether the error is known to
// report that the encoding advertised by the entity is unknown.
func IsUnknownEncoding(err error) bool {
	return errors.As(err, new(UnknownEncodingError))
}

func encodingReader(enc string, r io.Reader) (io.Reader, error) {
	var dec io.Reader
	switch strings.ToLower(enc) {
	case "quoted-printable":
		// Wrap with lenient QP reader to handle malformed QP encoding
		// (e.g., "= " followed by whitespace which some encoders produce)
		wrapped := &lenientQPReader{wrapped: bufio.NewReader(r)}
		dec = quotedprintable.NewReader(wrapped)
	case "base64":
		wrapped := &whitespaceReplacingReader{wrapped: r}
		dec = base64.NewDecoder(base64.StdEncoding, wrapped)
	case "7bit", "8bit", "binary", "":
		dec = r
	default:
		return nil, fmt.Errorf("unhandled encoding %q", enc)
	}
	return dec, nil
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

func encodingWriter(enc string, w io.Writer) (io.WriteCloser, error) {
	var wc io.WriteCloser
	switch strings.ToLower(enc) {
	case "quoted-printable":
		wc = quotedprintable.NewWriter(w)
	case "base64":
		wc = base64.NewEncoder(base64.StdEncoding, &lineWrapper{w: w, maxLineLen: 76})
	case "7bit", "8bit":
		wc = nopCloser{&lineWrapper{w: w, maxLineLen: 998}}
	case "binary", "":
		wc = nopCloser{w}
	default:
		return nil, fmt.Errorf("unhandled encoding %q", enc)
	}
	return wc, nil
}

// whitespaceReplacingReader replaces space and tab characters with a LF so
// base64 bodies with a continuation indent can be decoded by the base64 decoder
// even though it is against the spec.
type whitespaceReplacingReader struct {
	wrapped io.Reader
}

func (r *whitespaceReplacingReader) Read(p []byte) (int, error) {
	n, err := r.wrapped.Read(p)

	for i := 0; i < n; i++ {
		if p[i] == ' ' || p[i] == '\t' {
			p[i] = '\n'
		}
	}

	return n, err
}

// lenientQPReader fixes malformed quoted-printable encoding while preserving valid sequences.
// It handles the following cases:
// 1. "=[spaces]\r\n" -> Soft line break (malformed). Strips spaces to "=\r\n".
// 2. "=[spaces][other]" -> Malformed literal. Encodes to "=3D[spaces][other]".
// 3. "=62" -> Valid QP. Preserved.
// 4. "=\r\n" -> Valid soft break. Preserved.
// 5. "[spaces]\r\n" -> Trailing whitespace. Encodes to "=20" or "=09".
type lenientQPReader struct {
	wrapped *bufio.Reader
	pending []byte // data waiting to be read (e.g. from expansion)
}

func (r *lenientQPReader) Read(p []byte) (int, error) {
	if len(r.pending) > 0 {
		n := copy(p, r.pending)
		r.pending = r.pending[n:]
		return n, nil
	}

	i := 0
	for i < len(p) {
		// Peek 1 byte
		bPrefix, err := r.wrapped.Peek(1)
		if err != nil {
			if i > 0 {
				return i, nil
			}
			return 0, err
		}

		b := bPrefix[0]

		if b == '=' {
			// Look ahead to categorize
			peekSize := 2
			spacesCount := 0
			foundEnd := false
			var targetChar byte
			isSoftBreak := false
			isLiteral := false

			for {
				peaked, err := r.wrapped.Peek(peekSize)
				if len(peaked) < peekSize {
					if err != nil && err != io.EOF {
						// Error peeking, treat as literal to be safe
						isLiteral = true
						foundEnd = true
						break
					}
					// EOF reached
					isLiteral = true
					foundEnd = true
					break
				}

				c := peaked[peekSize-1]
				if c == ' ' || c == '\t' {
					spacesCount++
					peekSize++
					continue
				}

				targetChar = c
				foundEnd = true
				break
			}

			if !foundEnd {
				// Buffer likely full or scan too long
				isLiteral = true
			} else {
				if targetChar == '\r' || targetChar == '\n' {
					if spacesCount > 0 {
						// Malformed soft break: =[spaces]\r\n
						isSoftBreak = true
					}
					// Else valid soft break: =\r\n (fall through to default copy)
				} else {
					// Not a soft break
					// Check if valid Hex if no spaces (e.g. =62)
					if spacesCount == 0 && isHex(targetChar) {
						// Need one more char for full hex check
						peaked3, _ := r.wrapped.Peek(3)
						if len(peaked3) >= 3 && isHex(peaked3[2]) {
							// Valid QP hex: =XX. Preserved.
						} else {
							// =X? or =X[EOF]. Malformed.
							isLiteral = true
						}
					} else {
						// =[spaces]X or =[bad char]. Malformed literal.
						isLiteral = true
					}
				}
			}

			if isLiteral {
				// Emit '=3D'. Convert = to =3D.
				toWrite := []byte("=3D")
				n := copy(p[i:], toWrite)
				i += n
				r.wrapped.Discard(1) // Consume '='

				if n < 3 {
					// Pending remainder
					r.pending = toWrite[n:]
					return i, nil
				}
				continue
			} else if isSoftBreak {
				// Malformed soft break: emit '=' but skip spaces
				p[i] = '='
				i++
				r.wrapped.Discard(1 + spacesCount) // Consume '=' and spaces
				continue
			}
			// Default: Valid soft break or Valid Hex. Copy '='.
			p[i] = '='
			i++
			r.wrapped.Discard(1)
		} else if b == ' ' || b == '\t' {
			// Check for trailing whitespace
			isTrailing := false
			peekSize := 1
			for {
				peaked, _ := r.wrapped.Peek(peekSize)
				if len(peaked) < peekSize {
					// EOF after spaces -> trailing
					isTrailing = true
					break
				}
				c := peaked[peekSize-1]
				if c == ' ' || c == '\t' {
					peekSize++
					continue
				}
				if c == '=' {
					// Check if '=' is a soft break
					// We need to look ahead from 'peekSize' (position of '=')
					// We can't reuse r.wrapped.Peek easily with large offset unless we loop?
					// actually we can.

					// We need to check if '=' is followed by [spaces]\r\n
					// We start scanning from peekSize.
					subPeekSize := peekSize + 1
					for {
						subPeaked, _ := r.wrapped.Peek(subPeekSize)
						if len(subPeaked) < subPeekSize {
							// EOF
							break
						}
						sc := subPeaked[subPeekSize-1]
						if sc == ' ' || sc == '\t' {
							subPeekSize++
							continue
						}
						if sc == '\r' || sc == '\n' {
							isTrailing = true
						}
						break
					}
				}
				break
			}

			if isTrailing {
				// Encode to =20 or =09
				toWrite := []byte("=20")
				if b == '\t' {
					toWrite = []byte("=09")
				}
				n := copy(p[i:], toWrite)
				i += n
				r.wrapped.Discard(1)
				if n < 3 {
					r.pending = toWrite[n:]
					return i, nil
				}
				continue
			}
			p[i] = b
			i++
			r.wrapped.Discard(1)
		} else {
			p[i] = b
			i++
			r.wrapped.Discard(1)
		}
	}
	return i, nil
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

type lineWrapper struct {
	w          io.Writer
	maxLineLen int

	curLineLen int
	cr         bool
}

func (w *lineWrapper) Write(b []byte) (int, error) {
	var written int
	for len(b) > 0 {
		var l []byte
		l, b = cutLine(b, w.maxLineLen-w.curLineLen)

		lf := bytes.HasSuffix(l, []byte("\n"))
		l = bytes.TrimSuffix(l, []byte("\n"))

		n, err := w.w.Write(l)
		if err != nil {
			return written, err
		}
		written += n

		cr := bytes.HasSuffix(l, []byte("\r"))
		if len(l) == 0 {
			cr = w.cr
		}

		if !lf && len(b) == 0 {
			w.curLineLen += len(l)
			w.cr = cr
			break
		}
		w.curLineLen = 0

		ending := []byte("\r\n")
		if cr {
			ending = []byte("\n")
		}
		_, err = w.w.Write(ending)
		if err != nil {
			return written, err
		}
		// If the written `\n` was part of the input bytes slice, then account for it.
		if lf {
			written++
		}
		w.cr = false
	}

	return written, nil
}

func cutLine(b []byte, max int) ([]byte, []byte) {
	for i := 0; i < len(b); i++ {
		if b[i] == '\r' && i == max {
			continue
		}
		if b[i] == '\n' {
			return b[:i+1], b[i+1:]
		}
		if i >= max {
			return b[:i], b[i:]
		}
	}
	return b, nil
}
