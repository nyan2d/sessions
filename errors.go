package sessions

import "errors"

var (
	ErrNotFound   = errors.New("session not found")
	ErrInvalidTTL = errors.New("ttl <= 0")
)
