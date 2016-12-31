package messages

import (
	"bufio"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

// Entity is either a message or a one of the parts in the body of a multipart
// entity.
type Entity struct {
	io.Reader // The entity's body.

	Header textproto.MIMEHeader // The entity's header.

	mediaType string
	mediaParams map[string]string
}

// NewEntity makes a new Entity with the provided header and body.
func NewEntity(header textproto.MIMEHeader, r io.Reader) *Entity {
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

	return &Entity{
		Reader: r,
		Header: header,
		mediaType: mediaType,
		mediaParams: mediaParams,
	}
}

// Read reads a message from r.
func Read(r io.Reader) (*Entity, error) {
	br := bufio.NewReader(r)
	h, err := textproto.NewReader(br).ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	return NewEntity(h, br), nil
}

// PartsReader returns a Reader that reads parts from this entity's body. If
// this entity is not multipart, it returns nil.
func (e *Entity) PartsReader() *Reader {
	if !strings.HasPrefix(e.mediaType, "multipart/") {
		return nil
	}

	return &Reader{multipart.NewReader(e, e.mediaParams["boundary"])}
}