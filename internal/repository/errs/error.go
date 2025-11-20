package errs

import "errors"

var (
	ErrNotFound   = errors.New("entity not found")
	ErrDuplicate  = errors.New("entity already exists")
	ErrConstraint = errors.New("constraint violation")
)
