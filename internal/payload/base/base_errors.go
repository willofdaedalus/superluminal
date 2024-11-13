package base

import "errors"

var (
	ErrHeaderPayloadMismatch = errors.New("header and payload type passed do not match")
)
