package ui

import (
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/lipgloss"
)

const (
	height = 1
)

func terminalHeaderLogic() string {
	t := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render("terminal")

	return t
}

func (m model) HeaderView() string {
	var tabs []string

	terminal := terminalHeaderLogic()

	session := lipgloss.NewStyle().
		Bold(true).
		Render("session")

	chat := lipgloss.NewStyle().
		Bold(true).
		Render("chat")

	if m.hostSide {
		tabs = []string{
			terminal,
			session,
			chat,
		}
	} else {
		tabs = []string{
			terminal,
			chat,
		}
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		Width(m.scrWidth).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
		}).
		Row(tabs...).
		Render()

	return t
}
