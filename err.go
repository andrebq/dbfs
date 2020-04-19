package dbfs

import (
	"errors"
)

var (
	errNotFound = errors.New("not found")
)

// IsErrNotFound return e, if and only if, e == errNotFound
func IsErrNotFound(e error) error {
	if e == errNotFound {
		return e
	}
	return nil
}
