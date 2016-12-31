package messages

import (
	"mime/multipart"
)

type Reader struct {
	r *multipart.Reader
}

func (r *Reader) NextPart() (*Entity, error) {
	if p, err := r.r.NextPart(); err != nil {
		return nil, err
	} else {
		return NewEntity(p.Header, p), nil
	}
}
