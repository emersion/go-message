package messages

import (
	"bufio"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

type Part struct {
	io.Reader

	Header textproto.MIMEHeader

	mediaType string
	mediaParams map[string]string
}

func NewPart(header textproto.MIMEHeader, r io.Reader) *Part {
	r = decode(header.Get("Content-Transfer-Encoding"), r)
	header.Del("Content-Transfer-Encoding")

	mediaType, mediaParams, _ := mime.ParseMediaType(header.Get("Content-Type"))
	if charset, ok := mediaParams["charset"]; ok {
		if converted, err := charsetReader(charset, r); err == nil {
			r = converted
		}

		mediaParams["charset"] = "utf-8"
		header.Set("Content-Type", mime.FormatMediaType(mediaType, mediaParams))
	}

	return &Part{
		Reader: r,
		Header: header,
		mediaType: mediaType,
		mediaParams: mediaParams,
	}
}

func ReadPart(r io.Reader) (*Part, error) {
	br := bufio.NewReader(r)
	h, err := textproto.NewReader(br).ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	return NewPart(h, br), nil
}

func (p *Part) ChildrenReader() *Reader {
	if !strings.HasPrefix(p.mediaType, "multipart/") {
		return nil
	}

	return &Reader{multipart.NewReader(p, p.mediaParams["boundary"])}
}
