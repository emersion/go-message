package message

import (
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
type Header2 struct {
	// Fields are in reverse order so that inserting a new field at the top is
	// cheap.
	l []headerField
	m map[string][]*headerField
}

func newHeader2(fs []headerField) Header2 {
	// Reverse order
	for i := len(fs)/2-1; i >= 0; i-- {
		opp := len(fs)-1-i
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
	f := &h.l[len(h.l) - 1]
	h.m[k] = append(h.m[k], f)
}

// Get gets the first value associated with the given key. If there are no
// values associated with the key, Get returns "".
func (h *Header2) Get(k string) string {
	fields := h.m[textproto.CanonicalMIMEHeaderKey(k)]
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields) - 1].v
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
	h *Header2
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
	return &fs.h.l[len(fs.h.l) - fs.cur - 1]
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
	h *Header2
	k string
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
	return fs.h.m[fs.k][len(fs.h.m[fs.k]) - fs.cur - 1]
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
