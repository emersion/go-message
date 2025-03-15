package bufmsg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
)

type Entity struct {
	Header   message.Header
	Children []*Entity

	originalHeader message.Header
	rawBody        []byte
	body           []byte
	bodyUpdated    bool
}

func New(header message.Header) *Entity {
	return &Entity{
		Header:         header,
		originalHeader: header.Copy(),
	}
}

func Read(r io.Reader) (*Entity, error) {
	br := bufio.NewReader(r)
	th, err := textproto.ReadHeader(br)
	if err != nil {
		return nil, err
	}
	return readWithHeader(th, br)
}

func readWithHeader(th textproto.Header, body io.Reader) (*Entity, error) {
	entity := &Entity{
		Header:         message.Header{th},
		originalHeader: message.Header{th.Copy()},
	}

	if boundary, ok := entity.multipartBoundary(); ok {
		mr := textproto.NewMultipartReader(body, boundary)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			child, err := readWithHeader(part.Header, part)
			if err != nil {
				return nil, err
			}

			entity.Children = append(entity.Children, child)
		}
	} else {
		b, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
		entity.rawBody = b
	}

	return &Entity{
		Header: entity.Header,
	}, nil
}

func (e *Entity) multipartBoundary() (string, bool) {
	mediaType, mediaParams, _ := e.Header.ContentType()
	return mediaParams["boundary"], strings.HasPrefix(mediaType, "multipart/")
}

func (e *Entity) IsMultipart() bool {
	_, ok := e.multipartBoundary()
	return ok
}

func (e *Entity) Body() ([]byte, error) {
	if e.body != nil {
		return e.body, nil
	}

	if e.rawBody == nil {
		return nil, nil // entity is empty
	}

	me, err := message.New(e.originalHeader, bytes.NewReader(e.rawBody))
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(me.Body)
	if err != nil {
		return nil, err
	}
	e.body = b
	return b, nil
}

func (e *Entity) SetBody(b []byte) {
	e.body = b
	e.rawBody = nil
	e.bodyUpdated = true
}

func (e *Entity) RawBody() ([]byte, error) {
	if e.IsMultipart() || e.bodyUpdated {
		var buf bytes.Buffer
		if err := e.writeBodyTo(&buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	} else {
		return e.rawBody, nil
	}
}

func (e *Entity) WriteTo(w io.Writer) error {
	if err := textproto.WriteHeader(w, e.Header.Header); err != nil {
		return err
	}
	return e.writeBodyTo(w)
}

func (e *Entity) writeBodyTo(w io.Writer) error {
	if e.IsMultipart() {
		if e.bodyUpdated {
			return fmt.Errorf("bufmsg: SetBody was called on a multipart message")
		}

		mw := textproto.NewMultipartWriter(w)

		// TODO: grab boundary from header, if any -- otherwise set it

		for _, child := range e.Children {
			pw, err := mw.CreatePart(child.Header.Header)
			if err != nil {
				return err
			}

			if err := child.writeBodyTo(pw); err != nil {
				return err
			}
		}

		return mw.Close()
	} else if e.bodyUpdated {
		return nil // TODO: encode body to w
	} else {
		_, err := w.Write(e.rawBody)
		return err
	}
}

type WalkFunc func(path []int, entity *Entity) error

type walkQueueItem struct {
	path   []int
	entity *Entity
}

func (e *Entity) Walk(walkFunc WalkFunc) error {
	stack := []walkQueueItem{
		{nil, e},
	}
	seen := make(map[*Entity]bool)
	for len(stack) > 0 {
		// Pop an item out of the stack
		it := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if seen[it.entity] {
			return fmt.Errorf("bufmsg: cyclic multipart entity")
		}
		seen[it.entity] = true

		if err := walkFunc(it.path, it.entity); err != nil {
			return err
		}

		// Insert children into the stack in reverse order
		for i := len(it.entity.Children) - 1; i >= 0; i-- {
			child := it.entity.Children[i]

			childPath := make([]int, len(it.path))
			copy(childPath, it.path)
			childPath = append(childPath, i)

			stack = append(stack, walkQueueItem{childPath, child})
		}
	}
	return nil
}

func (e *Entity) Child(path []int) *Entity {
	cur := e
	for _, index := range path {
		if index >= len(cur.Children) {
			return nil
		}
		cur = cur.Children[index]
	}
	return cur
}
