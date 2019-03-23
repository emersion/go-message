package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
)

type headerField struct {
	k string
	v string
}

func newHeaderField(k, v string) headerField {
	return headerField{k: textproto.CanonicalMIMEHeaderKey(k), v: v}
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
type Header2 struct {
	// Fields are in reverse order so that inserting a new field at the top is
	// cheap.
	l []headerField
	m map[string][]*headerField
}

func newHeader2(fs []headerField) Header2 {
	// Reverse order
	for i := len(fs)/2 - 1; i >= 0; i-- {
		opp := len(fs) - 1 - i
		fs[i], fs[opp] = fs[opp], fs[i]
	}

	// Populate map
	var m map[string][]*headerField
	if len(fs) > 0 {
		m = make(map[string][]*headerField)
		for i, f := range fs {
			m[f.k] = append(m[f.k], &fs[i])
		}
	}

	return Header2{l: fs, m: m}
}

// Add adds the key, value pair to the header. It prepends to any existing
// fields associated with key.
func (h *Header2) Add(k, v string) {
	k = textproto.CanonicalMIMEHeaderKey(k)

	if h.m == nil {
		h.m = make(map[string][]*headerField)
	}

	h.l = append(h.l, newHeaderField(k, v))
	f := &h.l[len(h.l)-1]
	h.m[k] = append(h.m[k], f)
}

// Get gets the first value associated with the given key. If there are no
// values associated with the key, Get returns "".
func (h *Header2) Get(k string) string {
	fields := h.m[textproto.CanonicalMIMEHeaderKey(k)]
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1].v
}

// Set sets the header fields associated with key to the single field value.
// It replaces any existing values associated with key.
func (h *Header2) Set(k, v string) {
	h.Del(k)
	h.Add(k, v)
}

// Del deletes the values associated with key.
func (h *Header2) Del(k string) {
	k = textproto.CanonicalMIMEHeaderKey(k)

	// Delete existing keys
	for i := len(h.l) - 1; i >= 0; i-- {
		if h.l[i].k == k {
			h.l = append(h.l[:i], h.l[i+1:]...)
		}
	}

	delete(h.m, k)
}

// Has checks whether the header has a field with the specified key.
func (h *Header2) Has(k string) bool {
	_, ok := h.m[textproto.CanonicalMIMEHeaderKey(k)]
	return ok
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
	h   *Header2
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
func (h Header2) Fields() HeaderFields {
	return &headerFields{&h, -1}
}

type headerFieldsByKey struct {
	h   *Header2
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
func (h Header2) FieldsByKey(k string) HeaderFields {
	return &headerFieldsByKey{&h, textproto.CanonicalMIMEHeaderKey(k), -1}
}

func readLineSlice(r *bufio.Reader) ([]byte, error) {
	var line []byte
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
		if c != ' ' && c != '\t' {
			r.UnreadByte()
			break
		}
		n++
	}
	return n
}

func readContinuedLineSlice(r *bufio.Reader) ([]byte, error) {
	// Read the first line.
	line, err := readLineSlice(r)
	if err != nil {
		return nil, err
	}

	if len(line) == 0 { // blank line - no continuation
		return line, nil
	}

	buf := trim(line)

	// Read continuation lines.
	for skipSpace(r) > 0 {
		line, err := readLineSlice(r)
		if err != nil {
			break
		}

		buf = append(buf, ' ')
		buf = append(buf, trim(line)...)
	}
	return buf, nil
}

func readHeader(r *bufio.Reader) (Header2, error) {
	var fs []headerField

	// The first line cannot start with a leading space.
	if buf, err := r.Peek(1); err == nil && isSpace(buf[0]) {
		line, err := readLineSlice(r)
		if err != nil {
			return newHeader2(fs), err
		}

		return newHeader2(fs), fmt.Errorf("message: malformed MIME header initial line: %v", string(line))
	}

	for {
		kv, err := readContinuedLineSlice(r)
		if len(kv) == 0 {
			return newHeader2(fs), err
		}

		// Key ends at first colon; should not have trailing spaces
		// but they appear in the wild, violating specs, so we remove
		// them if present.
		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			return newHeader2(fs), fmt.Errorf("message: malformed MIME header line: %v", string(kv))
		}

		endKey := i
		for endKey > 0 && isSpace(kv[endKey-1]) {
			endKey--
		}

		key := textproto.CanonicalMIMEHeaderKey(string(kv[:endKey]))

		// As per RFC 7230 field-name is a token, tokens consist of one or more chars.
		// We could return a ProtocolError here, but better to be liberal in what we
		// accept, so if we get an empty key, skip it.
		if key == "" {
			continue
		}

		// Skip initial spaces in value.
		i++ // skip colon
		for i < len(kv) && isSpace(kv[i]) {
			i++
		}

		value := string(kv[i:])

		fs = append(fs, newHeaderField(key, value))

		if err != nil {
			return newHeader2(fs), err
		}
	}
}

func writeHeader2(w io.Writer, h Header2) error {
	// TODO: wrap lines
	for i := len(h.l) - 1; i >= 0; i-- {
		f := h.l[i]
		if _, err := io.WriteString(w, f.k+": "+f.v+"\r\n"); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\r\n")
	return err
}
