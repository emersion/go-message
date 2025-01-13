package message

import (
	"io"
)

type Entity = Reader

func New(header Header, body io.Reader) (*Entity, error) {
	return NewReader(header, body)
}

func NewMultipart(header Header, parts []*Entity) (*Entity, error) {
	return NewMultipartReader(header, parts)
}
