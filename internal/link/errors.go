package link

import "errors"

var (
	ErrAliasTaken = errors.New("alias taken")
	// Validation errors
	ErrURLMalformed         = errors.New("url malformed")
	ErrURLMissingHost       = errors.New("url missing host")
	ErrTTLNonPositive       = errors.New("ttl is not positive")
	ErrURLRequired          = errors.New("url required")
	ErrURLTooLong           = errors.New("url too long")
	ErrURLUnsupportedScheme = errors.New("url unsupported scheme")
)
