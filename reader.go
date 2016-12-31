package messages

import (
	"mime/multipart"
)

type Reader struct {
	r *multipart.Reader
}

func (r *Reader) NextPart() (*Part, error) {
	if p, err := r.r.NextPart(); err != nil {
		return nil, err
	} else {
		return NewPart(p.Header, p), nil
	}
}
