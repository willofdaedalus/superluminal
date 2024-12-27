package ui

import (
	"time"
	"github.com/charmbracelet/bubbles/viewport"
)

type clientEntry struct {
	clientName string
	timeJoined time.Time
}

type model struct {
	msg           string
	clientList    []clientEntry
	scrWidth int
	scrHeight int
	currentTab int
	hostSide bool
	termView viewport.Model
}
