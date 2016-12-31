package messages

import (
	"mime/multipart"
)

// MultipartReader is an iterator over parts in a MIME multipart body.
type MultipartReader struct {
	r *multipart.Reader
}

// NextPart returns the next part in the multipart or an error. When there are
// no more parts, the error io.EOF is returned.
func (r *MultipartReader) NextPart() (*Entity, error) {
	p, err := r.r.NextPart()
	if err != nil {
		return nil, err
	}
	return NewEntity(p.Header, p), nil
}
