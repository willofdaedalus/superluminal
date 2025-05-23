package utils

import "errors"

// server related errors
var (
	ErrCtxTimeOut       = errors.New("sprlmnl: context timed out")
	ErrFailedServerAuth = errors.New("sprlmnl: client failed server auth")
	ErrClientExchFailed = errors.New("sprlmnl: couldn't reach client after retries")
	ErrConnectionClosed = errors.New("sprlmnl: connection closed on other side")
	ErrWrongServer      = errors.New("sprlmnl: connected to non-superluminal server; exiting")
	ErrDecodeFailed     = errors.New("sprlmnl: couldn't decode data")
	ErrWrongPass        = errors.New("sprlmnl: client submitted the wrong passphrase")
	ErrServerFull       = errors.New("sprlmnl: server is full")
)

var (
	ErrInvalidHeader           = errors.New("sprlmnl: invalid server header")
	ErrSetDeadlineUnsuccessful = errors.New("sprlmnl: couldn't set deadline for operation")
	ErrClientFailedAuth        = errors.New("sprlmnl: entered wrong passphrase")
	ErrClientEarlyExit         = errors.New("sprlmnl: client wants out")
	ErrFailedAfterRetries      = errors.New("sprlmnl: couldn't send message after retries")
	ErrUnknownHeader           = errors.New("sprlmnl: unknown server header")
	ErrLongWait                = errors.New("sprlmnl: waited too long for input")
)

// payload related errors
var (
	ErrUnspecifiedPayload    = errors.New("sprlmnl: payload is unspecified")
	ErrPayloadHeaderMismatch = errors.New("header and payload type passed do not match")
)
