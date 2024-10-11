package config

import "time"

const (
	MaxConnectionTime = time.Second * 5
	ShutdownMsg = "SHUTDOWN"
	ServerClosed = "use of closed network connection"
	ConnectionReset = "connection reset by peer"
	NoSuchHost = "no such host"
	RejectedPass = "server rejected your passphrase"
	ServerIO = "server i/o timeout"
	ServerFull = "server is at capacity"
)

var (
	DefaultConnection string
)

