package messages

import (
	"mime/multipart"
)

// Reader is an iterator over parts in a MIME multipart body.
type Reader struct {
	r *multipart.Reader
}

// NextPart returns the next part in the multipart or an error. When there are
// no more parts, the error io.EOF is returned.
func (r *Reader) NextPart() (*Entity, error) {
	p, err := r.r.NextPart()
	if err != nil {
		return nil, err
	}
	return NewEntity(p.Header, p), nil
}
