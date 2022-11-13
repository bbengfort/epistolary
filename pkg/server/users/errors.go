package users

import "errors"

// Standard errors for database operations and checking.
var (
	ErrNoUserID      = errors.New("this operation requires a user id")
	ErrNotDerivedKey = errors.New("passwords must be stored as a derived key")
)
