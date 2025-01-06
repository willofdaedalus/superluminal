package ui

import (
	"time"
	"willofdaedalus/superluminal/internal/backend"
	"willofdaedalus/superluminal/internal/client"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type clientEntry struct {
	clientName string
	timeJoined time.Time
}

type model struct {
	clientList      []clientEntry
	startInputs     []textinput.Model
	scrWidth        int
	scrHeight       int
	currentTab      int
	view            int
	hostSide        bool
	currentView     int
	viewport        viewport.Model
	currentSession  *backend.Session
	client          *client.Client
	startCurField   int
	sessClientCount uint8
	showErrMsg      bool
	clientErrChan   chan error
}

const (
	termView = iota + 1
)

const (
	hostMaxTabs   = 3
	clientMaxTabs = 2
)

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
