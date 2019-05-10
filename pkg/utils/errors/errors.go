package errors

import (
	e "errors"
)

// Error is indagate struct of error
type Error struct {
	Code string
	Msg  string
	Op   string
	Err  error
}

// Define common errors here, once this set of constants changes,
// the swagger for Error.properties.conde.enum alse be updated.
const (
	NotFound         = "not found"
	Internal         = "internal error"
	Invalid          = "invalid"
	MethodNotAllowed = "method not allowed"
	Unauthorized     = "unauthorized"
)

var (
	ErrKeyNotFound   = e.New("key not found")
	ErrTxNotWritable = e.New("transaction is not writable")
)

func (e *Error) Error() string {
	return ""
}

// ErrorCode returns the code of the root error
func ErrorCode(err error) string {
	if err == nil {
		return ""
	}

	e, ok := err.(*Error)
	if !ok {
		return Internal
	}

	if err == nil {
		return ""
	}

	if e.Code != "" {
		return e.Code
	}
	if e.Err != nil {
		return ErrorCode(e.Err)
	}
	return Internal
}

// IsNotFound returns a boolean indicating whether the error is known to report that a key or was not found.
func IsNotFound(err error) bool {
	return err == ErrKeyNotFound
}
