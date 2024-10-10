package config

import "time"

const (
	MaxConnectionTime = time.Second * 5
	ServerClosed = "use of closed network connection"
	ConnectionReset = "connection reset by peer"
	NoSuchHost = "no such host"
	RejectedPass = "server rejected your passphrase"
)

var (
	DefaultConnection string
)

