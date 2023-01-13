package epistles

import "errors"

var (
	ErrIDRequired   = errors.New("cannot execute query without an id stored on the model")
	ErrLinkRequired = errors.New("cannot fetch epistle information without a link")
)
