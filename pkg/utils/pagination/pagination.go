package pagination

import (
	"errors"
	"time"
)

const (
	DefaultPageSize uint32 = 20
	MaximumPageSize uint32 = 1000
	CursorDuration         = 12 * time.Hour
)

var (
	ErrMissingExpiration  = errors.New("cursor does not have an expires timestamp")
	ErrCursorExpired      = errors.New("cursor has expired and is no longer useable")
	ErrUnparsableToken    = errors.New("could not parse the next page token")
	ErrTokenQueryMismatch = errors.New("cannot change query parameters during pagination")
	ErrPageSizeTooLarge   = errors.New("page size is greater than the maximum allowed page size")
)

func New(startIndex, endIndex int64, pageSize uint32) *Cursor {
	if pageSize == 0 || pageSize > MaximumPageSize {
		pageSize = DefaultPageSize
	}

	return &Cursor{
		Start: startIndex,
		End:   endIndex,
		Size:  pageSize,
		Exp:   time.Now().Add(CursorDuration).UnixMilli(),
	}
}

func Parse(token string) (cursor *Cursor, err error) {
	cursor = &Cursor{}
	if err = cursor.UnmarshalText([]byte(token)); err != nil {
		return nil, err
	}

	if cursor.Size > MaximumPageSize {
		return nil, ErrPageSizeTooLarge
	}

	var expired bool
	if expired, err = cursor.HasExpired(); err != nil {
		return nil, err
	}

	if expired {
		return nil, ErrCursorExpired
	}

	return cursor, nil
}

func (c Cursor) PrevPage() *Cursor {
	return New(max(int64(0), c.Start-int64(c.Size)), c.Start, c.Size)
}

func (c Cursor) PageToken() (_ string, err error) {
	if c.Size > MaximumPageSize {
		return "", ErrPageSizeTooLarge
	}

	var expired bool
	if expired, err = c.HasExpired(); err != nil {
		return "", err
	}

	if expired {
		return "", ErrCursorExpired
	}

	var token []byte
	if token, err = c.MarshalText(); err != nil {
		return "", err
	}
	return string(token), nil
}

func (c Cursor) HasExpired() (bool, error) {
	if c.Exp == 0 {
		return false, ErrMissingExpiration
	}
	return time.Now().After(c.Expires()), nil
}

func (c Cursor) IsZero() bool {
	return c.Start == 0 && c.End == 0 && c.Size == 0 && c.Exp == 0
}
