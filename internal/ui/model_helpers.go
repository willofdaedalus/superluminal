package ui

import (
	"fmt"
)

import "github.com/charmbracelet/lipgloss"

func (m model) renderMainBorder() string {
	outer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(m.scrWidth - 2).
		Height(m.scrHeight - 2).
		Align(lipgloss.Left).
		Render(fmt.Sprintf("%dx%d", m.scrWidth, m.scrHeight))

	return outer
}

func NewModel() model {
	clients := make([]clientEntry, 32)

	return model{
		clientList: clients,
		hostSide:   true,
	}
}
