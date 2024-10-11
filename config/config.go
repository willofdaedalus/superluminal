package config

import (
	"errors"
	"time"
)

var (
	// header related errors
	ErrUnknownId           = errors.New("sprlmnl: unknown identifer for header")
	ErrEmptyHeader         = errors.New("sprlmnl: server header is empty")
	ErrHeaderShort         = errors.New("sprlmnl: server header too short")
	ErrTimeDifference      = errors.New("sprlmnl: header time difference too large")
	ErrDeadlineTimeout     = errors.New("sprlmnl: authentication key read timeout")
	ErrInvalidTimestamp    = errors.New("sprlml: failed to parse header timestamp")
	ErrInvalidHeaderFormat = errors.New("sprlmnl: server header is invalid")

	// struct encoding related error
	ErrEncodingClient = errors.New("sprlmnl: couldn't encode client data for server")
	ErrSendingClient  = errors.New("sprlmnl: couldn't send client across the network")

	// server related errors
	ErrServerFull     = errors.New("sprlmnl: server is at capacity. contact session owner")
	ErrServerReset    = errors.New("sprlmnl: connection reset at server side")
	ErrServerClosed   = errors.New("sprlmnl: server is not receiving connections")
	ErrServerTimeout  = errors.New("sprlmnl: server timed out due to i/o operation")
	ErrServerShutdown = errors.New("sprlmnl: server has shutdown")
)

const (
	MaxConnectionTime = time.Second * 5
	ShutdownMsg       = "SHUTDOWN"
	ServerClosed      = "use of closed network connection"
	ConnectionReset   = "connection reset by peer"
	NoSuchHost        = "no such host"
	RejectedPass      = "server rejected your passphrase"
	ServerIO          = "server i/o timeout"
	ServerFull        = "server is at capacity"
)

var (
	DefaultConnection string
)
