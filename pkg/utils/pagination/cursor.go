package pagination

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/jxskiss/base62"
)

type Cursor struct {
	Start int64
	End   int64
	Size  uint32
	Exp   int64
}

func (c Cursor) GobEncode() ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(c.Start); err != nil {
		return nil, err
	}

	if err := encoder.Encode(c.End); err != nil {
		return nil, err
	}

	if err := encoder.Encode(c.Size); err != nil {
		return nil, err
	}

	if err := encoder.Encode(c.Exp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Cursor) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&c.Start); err != nil {
		return err
	}

	if err := decoder.Decode(&c.End); err != nil {
		return err
	}

	if err := decoder.Decode(&c.Size); err != nil {
		return err
	}

	if err := decoder.Decode(&c.Exp); err != nil {
		return err
	}

	return nil
}

func (c Cursor) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(c); err != nil {
		return nil, err
	}
	return base62.Encode(buf.Bytes()), nil
}

func (c *Cursor) UnmarshalText(data []byte) (err error) {
	var decoded []byte
	if decoded, err = base62.Decode(data); err != nil {
		return err
	}

	buf := bytes.NewReader(decoded)
	return gob.NewDecoder(buf).Decode(c)
}

func (c Cursor) Expires() time.Time {
	return time.UnixMilli(c.Exp)
}
