package ui

import (
	"time"
	"willofdaedalus/superluminal/internal/backend"

	"github.com/charmbracelet/bubbles/viewport"
)

type clientEntry struct {
	clientName string
	timeJoined time.Time
}

type model struct {
	msg            string
	clientList     []clientEntry
	scrWidth       int
	scrHeight      int
	currentTab     int
	hostSide       bool
	currentView    int
	viewport       viewport.Model
	currentSession *backend.Session
}

const (
	termView = iota + 1
)

const (
	hostMaxTabs   = 3
	clientMaxTabs = 2
)

func NewModel(session *backend.Session) model {
	clients := make([]clientEntry, 32)

	return model{
		clientList:     clients,
		hostSide:       true,
		currentView:    termView,
		currentSession: session,
	}
}

func (m *model) switchTab() {
	tabCount := hostMaxTabs - 1

	if !m.hostSide {
		tabCount = clientMaxTabs - 1
	}

	if m.currentTab == tabCount {
		m.currentTab = 0
		return
	}
	m.currentTab += 1
}
