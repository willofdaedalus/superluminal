package backend

import (
	"bytes"
	"net"
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

type Session struct {
	Owner    string
	maxConns uint8
	pass     string
	hash     string
	clients  map[string]*sessionClient
	listener net.Listener
	reader   bytes.Reader
}
