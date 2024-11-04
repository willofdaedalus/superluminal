package ui

import (
	"time"
)

type clientEntry struct {
	clientName string
	timeJoined time.Time
}

type model struct {
	msg           string
	clientList    []clientEntry
	width, height int
}
