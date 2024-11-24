package backend

import (
	"bytes"
	"net"
	"sync"
	"time"
)

type sessionClient struct {
	name    string
	pass    string
	uuid    string
	conn    net.Conn
	joined  time.Time
	isOwner bool
}

type session struct {
	Owner         string
	maxConns      uint8
	pass          string
	hash          string
	clients       map[string]*sessionClient
	listener      net.Listener
	reader        bytes.Reader
	mu            sync.Mutex
	passRegenTime time.Duration
	heartbeatTime time.Duration
}
