package textproto

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"regexp"
	"strings"
)

type headerField struct {
	b []byte // Raw header field, including whitespace
	k string
	v string
}

func newHeaderField(k, v string, b []byte) headerField {
	return headerField{k: textproto.CanonicalMIMEHeaderKey(k), v: v, b: b}
}

// A Header represents the key-value pairs in a message header.
//
// The header representation is idempotent: if the header can be read and
// written, the result will be exactly the same as the original (including
// whitespace). This is required for e.g. DKIM.
//
// Mutating the header is restricted: the only two allowed operations are
// inserting a new header field at the top and deleting a header field. This is
// again necessary for DKIM.
type Header struct {
	// Fields are in reverse order so that inserting a new field at the top is
	// cheap.
	l []headerField
	m map[string][]*headerField
}

func makeHeaderMap(fs []headerField) map[string][]*headerField {
	if len(fs) == 0 {
		return nil
	}

	m := make(map[string][]*headerField)
	for i, f := range fs {
		m[f.k] = append(m[f.k], &fs[i])
	}
	return m
}

func newHeader(fs []headerField) Header {
	// Reverse order
	for i := len(fs)/2 - 1; i >= 0; i-- {
		opp := len(fs) - 1 - i
		fs[i], fs[opp] = fs[opp], fs[i]
	}

	// Populate map
	m := makeHeaderMap(fs)

	return Header{l: fs, m: m}
}

// Add adds the key, value pair to the header. It prepends to any existing
// fields associated with key.
func (h *Header) Add(k, v string) {
	k = textproto.CanonicalMIMEHeaderKey(k)

	if h.m == nil {
		h.m = make(map[string][]*headerField)
	}

	h.l = append(h.l, newHeaderField(k, v, nil))
	f := &h.l[len(h.l)-1]
	h.m[k] = append(h.m[k], f)
}

// Get gets the first value associated with the given key. If there are no
// values associated with the key, Get returns "".
func (h *Header) Get(k string) string {
	fields := h.m[textproto.CanonicalMIMEHeaderKey(k)]
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1].v
}

// Set sets the header fields associated with key to the single field value.
// It replaces any existing values associated with key.
func (h *Header) Set(k, v string) {
	h.Del(k)
	h.Add(k, v)
}

// Del deletes the values associated with key.
func (h *Header) Del(k string) {
	k = textproto.CanonicalMIMEHeaderKey(k)

	delete(h.m, k)

	// Delete existing keys
	for i := len(h.l) - 1; i >= 0; i-- {
		if h.l[i].k == k {
			h.l = append(h.l[:i], h.l[i+1:]...)
		}
	}
}

// Has checks whether the header has a field with the specified key.
func (h *Header) Has(k string) bool {
	_, ok := h.m[textproto.CanonicalMIMEHeaderKey(k)]
	return ok
}

// Copy creates an independent copy of the header.
func (h *Header) Copy() Header {
	l := make([]headerField, len(h.l))
	copy(l, h.l)
	m := makeHeaderMap(l)
	return Header{l: l, m: m}
}

// HeaderFields iterates over header fields. Its cursor starts before the first
// field of the header. Use Next to advance from field to field.
type HeaderFields interface {
	// Next advances to the next header field. It returns true on success, or
	// false if there is no next field.
	Next() (more bool)
	// Key returns the key of the current field.
	Key() string
	// Value returns the value of the current field.
	Value() string
	// Del deletes the current field.
	Del()
}

type headerFields struct {
	h   *Header
	cur int
}

func (fs *headerFields) Next() bool {
	fs.cur++
	return fs.cur < len(fs.h.l)
}

func (fs *headerFields) field() *headerField {
	if fs.cur < 0 {
		panic("message: HeaderFields method called before Next")
	}
	if fs.cur >= len(fs.h.l) {
		panic("message: HeaderFields method called after Next returned false")
	}
	return &fs.h.l[len(fs.h.l)-fs.cur-1]
}

func (fs *headerFields) Key() string {
	return fs.field().k
}

func (fs *headerFields) Value() string {
	return fs.field().v
}

func (fs *headerFields) Del() {
	f := fs.field()

	ok := false
	for i, ff := range fs.h.m[f.k] {
		if ff == f {
			ok = true
			fs.h.m[f.k] = append(fs.h.m[f.k][:i], fs.h.m[f.k][i+1:]...)
			if len(fs.h.m[f.k]) == 0 {
				delete(fs.h.m, f.k)
			}
			break
		}
	}
	if !ok {
		panic("message: field not found in Header.m")
	}

	fs.h.l = append(fs.h.l[:fs.cur], fs.h.l[fs.cur+1:]...)
	fs.cur--
}

// Fields iterates over all the header fields.
//
// The header may not be mutated while iterating, except using HeaderFields.Del.
func (h *Header) Fields() HeaderFields {
	return &headerFields{h, -1}
}

type headerFieldsByKey struct {
	h   *Header
	k   string
	cur int
}

func (fs *headerFieldsByKey) Next() bool {
	fs.cur++
	return fs.cur < len(fs.h.m[fs.k])
}

func (fs *headerFieldsByKey) field() *headerField {
	if fs.cur < 0 {
		panic("message: HeaderFields.Key or Value called before Next")
	}
	if fs.cur >= len(fs.h.m[fs.k]) {
		panic("message: HeaderFields.Key or Value called after Next returned false")
	}
	return fs.h.m[fs.k][len(fs.h.m[fs.k])-fs.cur-1]
}

func (fs *headerFieldsByKey) Key() string {
	return fs.field().k
}

func (fs *headerFieldsByKey) Value() string {
	return fs.field().v
}

func (fs *headerFieldsByKey) Del() {
	f := fs.field()

	ok := false
	for i := range fs.h.l {
		if f == &fs.h.l[i] {
			ok = true
			fs.h.l = append(fs.h.l[:i], fs.h.l[i+1:]...)
			break
		}
	}
	if !ok {
		panic("message: field not found in Header.l")
	}

	fs.h.m[fs.k] = append(fs.h.m[fs.k][:fs.cur], fs.h.m[fs.k][fs.cur+1:]...)
	if len(fs.h.m[fs.k]) == 0 {
		delete(fs.h.m, fs.k)
	}
	fs.cur--
}

// FieldsByKey iterates over all fields having the specified key.
//
// The header may not be mutated while iterating, except using HeaderFields.Del.
func (h *Header) FieldsByKey(k string) HeaderFields {
	return &headerFieldsByKey{h, textproto.CanonicalMIMEHeaderKey(k), -1}
}

func readLineSlice(r *bufio.Reader, line []byte) ([]byte, error) {
	for {
		l, more, err := r.ReadLine()
		if err != nil {
			return nil, err
		}

		// Avoid the copy if the first call produced a full line.
		if line == nil && !more {
			return l, nil
		}

		line = append(line, l...)
		if !more {
			break
		}
	}

	return line, nil
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// trim returns s with leading and trailing spaces and tabs removed.
// It does not assume Unicode or UTF-8.
func trim(s []byte) []byte {
	i := 0
	for i < len(s) && isSpace(s[i]) {
		i++
	}
	n := len(s)
	for n > i && isSpace(s[n-1]) {
		n--
	}
	return s[i:n]
}

// skipSpace skips R over all spaces and returns the number of bytes skipped.
func skipSpace(r *bufio.Reader) int {
	n := 0
	for {
		c, err := r.ReadByte()
		if err != nil {
			// bufio will keep err until next read.
			break
		}
		if !isSpace(c) {
			r.UnreadByte()
			break
		}
		n++
	}
	return n
}

func hasContinuationLine(r *bufio.Reader) bool {
	c, err := r.ReadByte()
	if err != nil {
		return false // bufio will keep err until next read.
	}
	r.UnreadByte()
	return isSpace(c)
}

func readContinuedLineSlice(r *bufio.Reader) ([]byte, error) {
	// Read the first line.
	line, err := readLineSlice(r, nil)
	if err != nil {
		return nil, err
	}

	if len(line) == 0 { // blank line - no continuation
		return line, nil
	}

	line = append(line, '\r', '\n')

	// Read continuation lines.
	for hasContinuationLine(r) {
		line, err = readLineSlice(r, line)
		if err != nil {
			break // bufio will keep err until next read.
		}

		line = append(line, '\r', '\n')
	}

	return line, nil
}

func writeContinued(b *strings.Builder, l []byte) {
	// Strip trailing \r, if any
	if len(l) > 0 && l[len(l)-1] == '\r' {
		l = l[:len(l)-1]
	}
	l = trim(l)
	if len(l) == 0 {
		return
	}
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.Write(l)
}

// Strip newlines and spaces around newlines.
func trimAroundNewlines(v []byte) string {
	var b strings.Builder
	for {
		i := bytes.IndexByte(v, '\n')
		if i < 0 {
			writeContinued(&b, v)
			break
		}
		writeContinued(&b, v[:i])
		v = v[i+1:]
	}

	return b.String()
}

// ReadHeader reads a MIME header from r. The header is a sequence of possibly
// continued Key: Value lines ending in a blank line.
func ReadHeader(r *bufio.Reader) (Header, error) {
	var fs []headerField

	// The first line cannot start with a leading space.
	if buf, err := r.Peek(1); err == nil && isSpace(buf[0]) {
		line, err := readLineSlice(r, nil)
		if err != nil {
			return newHeader(fs), err
		}

		return newHeader(fs), fmt.Errorf("message: malformed MIME header initial line: %v", string(line))
	}

	for {
		kv, err := readContinuedLineSlice(r)
		if len(kv) == 0 {
			return newHeader(fs), err
		}

		// Key ends at first colon; should not have trailing spaces but they
		// appear in the wild, violating specs, so we remove them if present.
		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			return newHeader(fs), fmt.Errorf("message: malformed MIME header line: %v", string(kv))
		}

		key := textproto.CanonicalMIMEHeaderKey(string(trim(kv[:i])))

		// As per RFC 7230 field-name is a token, tokens consist of one or more
		// chars. We could return a an error here, but better to be liberal in
		// what we accept, so if we get an empty key, skip it.
		if key == "" {
			continue
		}

		i++ // skip colon
		v := kv[i:]

		value := trimAroundNewlines(v)
		fs = append(fs, newHeaderField(key, value, kv))

		if err != nil {
			return newHeader(fs), err
		}
	}
}

const maxHeaderLen = 76

// Regexp that detects Quoted Printable (QP) characters
var qpReg = regexp.MustCompile("(=[0-9A-Z]{2,2})+")

// formatHeaderField formats a header field, ensuring each line is no longer
// than 76 characters. It tries to fold lines at whitespace characters if
// possible. If the header contains a word longer than this limit, it will be
// split.
func formatHeaderField(k, v string) string {
	s := k + ": "

	if v == "" {
		return s + "\r\n"
	}

	first := true
	for len(v) > 0 {
		maxlen := maxHeaderLen
		if first {
			maxlen -= len(s)
		}

		// We'll need to fold before i
		foldBefore := maxlen + 1
		foldAt := len(v)

		var folding string
		if foldBefore > len(v) {
			// We reached the end of the string
			if v[len(v)-1] != '\n' {
				// If there isn't already a trailing CRLF, insert one
				folding = "\r\n"
			}
		} else {
			// Find the last QP character before limit
			foldAtQP := qpReg.FindAllStringIndex(v[:foldBefore], -1)
			// Find the closest whitespace before i
			foldAtEOL := strings.LastIndexAny(v[:foldBefore], " \t\n")

			// Fold at the latest whitespace by default
			foldAt = foldAtEOL

			// if there are QP characters in the string
			if len(foldAtQP) > 0 {
				// Get the start index of the last QP character
				foldAtQPLastIndex := foldAtQP[len(foldAtQP)-1][0]
				if foldAtQPLastIndex > foldAt {
					// Fold at the latest QP character if there are no whitespaces after it and before line hard limit
					foldAt = foldAtQPLastIndex
				}
			}

			if foldAt == 0 {
				// The whitespace we found was the previous folding WSP
				foldAt = foldBefore - 1
			} else if foldAt < 0 {
				// We didn't find any whitespace, we have to insert one
				foldAt = foldBefore - 2
			}

			switch v[foldAt] {
			case ' ', '\t':
				if v[foldAt-1] != '\n' {
					folding = "\r\n" // The next char will be a WSP, don't need to insert one
				}
			case '\n':
				folding = "" // There is already a CRLF, nothing to do
			default:
				folding = "\r\n " // Another char, we need to insert CRLF + WSP
			}
		}

		s += v[:foldAt] + folding
		v = v[foldAt:]
		first = false
	}

	return s
}

// WriteHeader writes a MIME header to w.
func WriteHeader(w io.Writer, h Header) error {
	// TODO: wrap lines when necessary

	for i := len(h.l) - 1; i >= 0; i-- {
		f := h.l[i]

		if f.b == nil {
			f.b = []byte(formatHeaderField(f.k, f.v))
		}

		if _, err := w.Write(f.b); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte{'\r', '\n'})
	return err
}
