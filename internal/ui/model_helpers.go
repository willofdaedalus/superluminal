package ui

import "github.com/charmbracelet/lipgloss"

func (m model) renderMainBorder() string {
	outer := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Bottom).
		Render("hello world")

	return outer
}

func NewModel() model {
	clients := make([]clientEntry, 32)

	return model{
		clientList: clients,
	}
}
