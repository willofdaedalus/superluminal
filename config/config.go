package config

import (
	"errors"
	"time"
)

// header related errors
var (
	ErrUnknownId           = errors.New("sprlmnl: unknown identifer for header")
	ErrHeaderShort         = errors.New("sprlmnl: server header length invalid")
	ErrTimeDifference      = errors.New("sprlmnl: header time difference too large")
	ErrDeadlineTimeout     = errors.New("sprlmnl: authentication key read timeout")
	ErrInvalidTimestamp    = errors.New("sprlml: failed to parse header timestamp")
	ErrInvalidHeaderFormat = errors.New("sprlmnl: server header is invalid")
)

// struct encoding related error
var (
	ErrEncodingClient = errors.New("sprlmnl: couldn't encode client data for server")
	ErrSendingClient  = errors.New("sprlmnl: couldn't send client across the network")
	ErrUnknownMessage = errors.New("sprlmnl: unknown message from server")
)

// server related errors
var (
	ErrServerFull       = errors.New("sprlmnl: server is at capacity. contact session owner")
	ErrServerReset      = errors.New("sprlmnl: connection reset at server side")
	ErrServerClosed     = errors.New("sprlmnl: server is not receiving connections")
	ErrServerTimeout    = errors.New("sprlmnl: server timed out due to i/o operation")
	ErrServerShutdown   = errors.New("sprlmnl: server has shutdown")
	ErrWrongServerPass  = errors.New("sprlmnl: server kicked you for wrong passphrase")
	ErrServerFailStart  = errors.New("sprlmnl: server couldn't start")
	ErrServerCtxTimeout = errors.New("sprlmnl: server context timed out on operation")
	ErrFailedToWriteMsg = errors.New("sprlmnl: server couldn't write a message to the connection")
	ErrClientAuthFailed = errors.New("sprlmnl: server kicked you for wrong passphrase")

	// session handler
	ErrPTYFailed = errors.New("sprlmnl: server couldn't create a pty session")
)

const (
	PTYFailed        = "server couldn't create a new pty session"
	ServerIO         = "server i/o timeout"
	NoSuchHost       = "no such host"
	ServerFull       = "server is at capacity"
	ShutdownMsg      = "SHUTDOWN"
	RejectedPass     = "server rejected your passphrase"
	ServerClosed     = "use of closed network connection"
	ConnectionReset  = "connection reset by peer"
	ClientAuthFailed = "server kicked you for wrong passphrase"

	MaxConnectionTime      = time.Second * 5
	MaxServerHandshakeTime = time.Second * 10
)

const (
	PreShutdown = "SHT:"
	PreInfo     = "INF:"
	PreHeader   = "HDR:"
	PreErr      = "ERR:"
)

var (
	DefaultConnection string
)
