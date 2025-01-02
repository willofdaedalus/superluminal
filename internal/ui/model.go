package ui

import (
	"willofdaedalus/superluminal/internal/backend"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	startView = iota + 1
)

const (
	headerSwitch = "alt+esc"
)

func NewModel(session *backend.Session) model {
	return model{
		view:           startView,
		startCurField:  1,
		clientList:     make([]clientEntry, 32),
		hostSide:       true,
		currentView:    termView,
		currentSession: session,
		startInputs:    readyStartInputs(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// handle input updates
	if m.view == startView && m.startCurField <= len(m.startInputs) {
		m.startInputs[m.startCurField-1], cmd = m.startInputs[m.startCurField-1].Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.view == startView {
				if err := m.validateStartInputs(); err != nil {
					m.showErrMsg = true
					return m, nil
				}
			}

		case "tab":
			if m.view == startView {
				m.startInputsLogic()
			}

		case headerSwitch:
			m.switchTab()
		}
	case tea.WindowSizeMsg:
		m.scrWidth = msg.Width
		m.scrHeight = msg.Height
	}
	return m, cmd
}

func (m model) View() string {
	return m.StartScreenView()
	// return m.startScreen()
	// return m.mainRender()
	// return m.HeaderView()
}
