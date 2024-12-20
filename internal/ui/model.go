package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	stateSwitch = "ctrl+esc"
	offset      = 2 // this offset is not fixed
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width - offset
		m.height = msg.Height - offset
	}

	return m, nil
}

func (m model) View() string {
	return m.renderMainBorder()
}
