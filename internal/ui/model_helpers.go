package ui

import (
	"time"

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
	currentView     int
	startCurField   int
	hostSide        bool
	showErrMsg      bool
	viewport        viewport.Model
	sessClientCount uint8
	appState        *state
	errMsg          string
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
	tabCount := m.appState.tabCount - 1
	if m.currentTab == tabCount {
		m.currentTab = 0
		return
	}
	m.currentTab += 1
}

func (m *model) setErrorMessage(msg string) {
	m.errMsg = msg
}

func (m *model) transitionView(view int) {
	m.view = view
}
