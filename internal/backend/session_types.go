package backend

import (
	"net"
	"os"
	"sync"
	"time"
	"willofdaedalus/superluminal/internal/pipeline"
	"willofdaedalus/superluminal/internal/utils"
)

type errMessage [2]string
type clientUniqID string

type sessionClient struct {
	name    string
	pass    string
	uuid    string
	conn    net.Conn
	joined  time.Time
	isOwner bool
}

type Session struct {
	Owner         string
	maxConns      uint8
	pass          string
	hash          string
	clients       map[string]*sessionClient
	pipeline      *pipeline.Pipeline
	listener      net.Listener
	signals       []os.Signal
	mu            sync.Mutex
	tracker       *utils.SyncTracker
	passRegenTime time.Duration
	heartbeatTime time.Duration
}
