package message

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
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
	// Normalize encoding type
	cleanEncodingType := strings.ToLower(enc)
	// Some encodings have a trailing dot
	cleanEncodingType = strings.Trim(cleanEncodingType, ".")
	// Some encodings have a leading and trailing single quote
	cleanEncodingType = strings.Trim(cleanEncodingType, "'")
	// Some encodings have a charset= prefix
	cleanEncodingType = strings.TrimPrefix(cleanEncodingType, "charset=")
	switch cleanEncodingType {
	case "quoted-printable":
		dec = quotedprintable.NewReader(r)
	case "base64":
		wrapped := &whitespaceReplacingReader{wrapped: r}
		dec = base64.NewDecoder(base64.StdEncoding, wrapped)
	case "7bit", "7-bit", "8bit", "8-bit", "binary", "", "ascii", "us-ascii", "utf8", "utf-8", "ansi_x3.4-1968", "text/plain", "text/html":
		dec = r
	case "iso-8859-1", "it-ascii":
		return decodeWithDecoder(charmap.ISO8859_1.NewDecoder(), r)
	case "windows-1252", "cp1252":
		return decodeWithDecoder(charmap.Windows1252.NewDecoder(), r)
	case "iso-2022-jp":
		return decodeWithDecoder(japanese.ISO2022JP.NewDecoder(), r)
	case "iso-8859-14":
		return decodeWithDecoder(charmap.ISO8859_14.NewDecoder(), r)
	case "iso-8859-2":
		return decodeWithDecoder(charmap.ISO8859_2.NewDecoder(), r)
	case "windows-1251":
		return decodeWithDecoder(charmap.Windows1251.NewDecoder(), r)
	case "iso-8859-15":
		return decodeWithDecoder(charmap.ISO8859_15.NewDecoder(), r)
	case "windows-1256":
		return decodeWithDecoder(charmap.Windows1256.NewDecoder(), r)
	case "koi8-u":
		return decodeWithDecoder(charmap.KOI8U.NewDecoder(), r)
	case "ks_c_5601-1987":
		return decodeWithDecoder(korean.EUCKR.NewDecoder(), r)
	case "gbk":
		return decodeWithDecoder(simplifiedchinese.GBK.NewDecoder(), r)
	case "iso-8859-6":
		return decodeWithDecoder(charmap.ISO8859_6.NewDecoder(), r)
	case "windows-1257":
		return decodeWithDecoder(charmap.Windows1257.NewDecoder(), r)
	case "windows-1250":
		return decodeWithDecoder(charmap.Windows1250.NewDecoder(), r)
	case "gb2312":
		return decodeWithDecoder(simplifiedchinese.GB18030.NewDecoder(), r)
	case "iso-8859-8-i":
		return decodeWithDecoder(charmap.ISO8859_8I.NewDecoder(), r)
	case "windows-1258":
		return decodeWithDecoder(charmap.Windows1258.NewDecoder(), r)
	case "big5":
		return decodeWithDecoder(traditionalchinese.Big5.NewDecoder(), r)
	case "windows-1255":
		return decodeWithDecoder(charmap.Windows1255.NewDecoder(), r)
	case "windows-1253":
		return decodeWithDecoder(charmap.Windows1253.NewDecoder(), r)
	case "iso-8859-9":
		return decodeWithDecoder(charmap.ISO8859_9.NewDecoder(), r)
	case "windows-1254":
		return decodeWithDecoder(charmap.Windows1254.NewDecoder(), r)
	case "shift-jis":
		return decodeWithDecoder(japanese.ShiftJIS.NewDecoder(), r)
	case "utf-16le":
		return decodeWithDecoder(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), r)
	case "iso-8859-5":
		return decodeWithDecoder(charmap.ISO8859_5.NewDecoder(), r)
	case "iso-8859-7":
		return decodeWithDecoder(charmap.ISO8859_7.NewDecoder(), r)
	case "iso_8859-1":
		return decodeWithDecoder(charmap.ISO8859_1.NewDecoder(), r)
	default:
		return nil, fmt.Errorf("unhandled encoding %q", enc)
	}
	return dec, nil
}

func decodeWithDecoder(decoder *encoding.Decoder, input io.Reader) (io.Reader, error) {
	var decodedBuffer bytes.Buffer
	transformedReader := transform.NewReader(input, decoder)
	if _, err := io.Copy(&decodedBuffer, transformedReader); err != nil {
		return nil, UnknownEncodingError{e: err}
	}
	return &decodedBuffer, nil
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

func encodingWriter(enc string, w io.Writer) (io.WriteCloser, error) {
	var wc io.WriteCloser
	// Normalize encoding type
	cleanEncodingType := strings.ToLower(enc)
	// Some encodings have a trailing dot
	cleanEncodingType = strings.Trim(cleanEncodingType, ".")
	// Some encodings have a leading and trailing single quote
	cleanEncodingType = strings.Trim(cleanEncodingType, "'")
	// Some encodings have a charset= prefix
	cleanEncodingType = strings.TrimPrefix(cleanEncodingType, "charset=")
	switch cleanEncodingType {
	case "quoted-printable":
		wc = quotedprintable.NewWriter(w)
	case "base64":
		wc = base64.NewEncoder(base64.StdEncoding, &lineWrapper{w: w, maxLineLen: 76})
	case "7bit", "7-bit", "8bit", "8-bit", "utf8", "utf-8":
		wc = nopCloser{&lineWrapper{w: w, maxLineLen: 998}}
	case "binary", "", "ascii", "us-ascii", "ansi_x3.4-1968", "text/plain", "text/html":
		wc = nopCloser{w}
	case "iso-8859-1", "it-ascii":
		wc = transform.NewWriter(w, charmap.ISO8859_1.NewEncoder())
	case "windows-1252", "cp1252":
		wc = transform.NewWriter(w, charmap.Windows1252.NewEncoder())
	case "iso-2022-jp":
		wc = transform.NewWriter(w, japanese.ISO2022JP.NewEncoder())
	case "iso-8859-14":
		wc = transform.NewWriter(w, charmap.ISO8859_14.NewEncoder())
	case "iso-8859-2":
		wc = transform.NewWriter(w, charmap.ISO8859_2.NewEncoder())
	case "windows-1251":
		wc = transform.NewWriter(w, charmap.Windows1251.NewEncoder())
	case "windows-1256":
		wc = transform.NewWriter(w, charmap.Windows1256.NewEncoder())
	case "koi8-u":
		wc = transform.NewWriter(w, charmap.KOI8U.NewEncoder())
	case "ks_c_5601-1987":
		wc = transform.NewWriter(w, korean.EUCKR.NewEncoder())
	case "gbk":
		wc = transform.NewWriter(w, simplifiedchinese.GBK.NewEncoder())
	case "iso-8859-6":
		wc = transform.NewWriter(w, charmap.ISO8859_6.NewEncoder())
	case "windows-1257":
		wc = transform.NewWriter(w, charmap.Windows1257.NewEncoder())
	case "windows-1250":
		wc = transform.NewWriter(w, charmap.Windows1250.NewEncoder())
	case "gb2312":
		wc = transform.NewWriter(w, simplifiedchinese.GB18030.NewEncoder())
	case "iso-8859-8-i":
		wc = transform.NewWriter(w, charmap.ISO8859_8I.NewEncoder())
	case "windows-1258":
		wc = transform.NewWriter(w, charmap.Windows1258.NewEncoder())
	case "big5":
		wc = transform.NewWriter(w, traditionalchinese.Big5.NewEncoder())
	case "windows-1255":
		wc = transform.NewWriter(w, charmap.Windows1255.NewEncoder())
	case "windows-1253":
		wc = transform.NewWriter(w, charmap.Windows1253.NewEncoder())
	case "iso-8859-9":
		wc = transform.NewWriter(w, charmap.ISO8859_9.NewEncoder())
	case "windows-1254":
		wc = transform.NewWriter(w, charmap.Windows1254.NewEncoder())
	case "shift-jis":
		wc = transform.NewWriter(w, japanese.ShiftJIS.NewEncoder())
	case "utf-16le":
		wc = transform.NewWriter(w, unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder())
	case "iso-8859-5":
		wc = transform.NewWriter(w, charmap.ISO8859_5.NewEncoder())
	case "iso-8859-7":
		wc = transform.NewWriter(w, charmap.ISO8859_7.NewEncoder())
	case "iso_8859-1":
		wc = transform.NewWriter(w, charmap.ISO8859_1.NewEncoder())
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
