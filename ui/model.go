package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	stateSwitch = "ctrl+esc"
)

type model struct {
	PtyContent chan []byte
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (m model) View() string {
	return ""
}
