package utils

import "errors"

// server related errors
var (
	ErrCtxTimeOut       = errors.New("sprlmnl: server context timed out")
	ErrFailedServerAuth = errors.New("sprlmnl: couldn't send server auth to client")
	ErrClientExchFailed = errors.New("sprlmnl: couldn't reach client after retries")
)
