package utils

import "errors"

// server related errors
var (
	ErrCtxTimeOut       = errors.New("sprlmnl: server context timed out")
	ErrFailedServerAuth = errors.New("sprlmnl: couldn't send server auth to client")
	ErrClientExchFailed = errors.New("sprlmnl: couldn't reach client after retries")
	ErrWrongServer      = errors.New("sprlmnl: connected to non-superluminal server; exiting")
	ErrDecodeFailed     = errors.New("sprlmnl: couldn't decode data")
)

var (
	ErrInvalidHeader        = errors.New("sprlmnl: invalid server header")
	ErrDeadlineUnsuccessful = errors.New("sprlmnl: couldn't set deadline for operation")
	ErrUnknownHeader        = errors.New("sprlmnl: unknown server header")
)
