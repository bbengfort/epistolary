package epistles

import "errors"

var (
	ErrIDRequired = errors.New("cannot execute query without an id stored on the model")
)
