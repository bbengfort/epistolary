package epistles

import "errors"

var (
	ErrIDRequired      = errors.New("cannot execute query without an id stored on the model")
	ErrLinkRequired    = errors.New("cannot fetch epistle information without a link")
	ErrMissingPageSize = errors.New("missing page size in paginated query")
	ErrAlreadyExists   = errors.New("reading already exists")
)
