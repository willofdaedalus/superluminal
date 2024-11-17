package utils

import "errors"

// server related errors
var (
	ErrCtxTimeOut       = errors.New("sprlmnl: server context timed out")
	ErrFailedServerAuth = errors.New("sprlmnl: couldn't send server auth to client")
	ErrClientExchFailed = errors.New("sprlmnl: couldn't reach client after retries")
	ErrServerClosed     = errors.New("sprlmnl: server is closed")
	ErrWrongServer      = errors.New("sprlmnl: connected to non-superluminal server; exiting")
	ErrDecodeFailed     = errors.New("sprlmnl: couldn't decode data")
	ErrWrongPass        = errors.New("sprlmnl: client submitted the wrong passphrase")
	ErrServerFull       = errors.New("sprlmnl: server is full")
)

var (
	ErrInvalidHeader        = errors.New("sprlmnl: invalid server header")
	ErrDeadlineUnsuccessful = errors.New("sprlmnl: couldn't set deadline for operation")
	ErrFailedAfterRetries   = errors.New("sprlmnl: couldn't send message after retries")
	ErrUnknownHeader        = errors.New("sprlmnl: unknown server header")
)

var (
	ErrUnspecifiedPayload = errors.New("sprlmnl: payload is unspecified")
)
