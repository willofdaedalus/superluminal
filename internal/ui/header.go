package ui

import (
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/lipgloss"
)

const (
	height = 1
)

func (m model) sessionHeaderLogic() string {
	text := "session"
	if m.currentTab == 1 {
		text = "[session]"
	}

	t := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(text)

	return t
}

func (m model) chatHeaderLogic() string {
	text := "chat"
	if m.currentTab == 2 {
		text = "[chat]"
	}

	t := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(text)

	return t
}

func (m model) terminalHeaderLogic() string {
	text := "terminal"
	if m.currentTab == 0 {
		text = "[terminal]"
	}

	t := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(text)

	return t
}

func (m model) HeaderView() string {
	headers := []string{
		m.terminalHeaderLogic(),
		m.sessionHeaderLogic(),
		m.chatHeaderLogic(),
	}

	if !m.hostSide {
		headers = []string {
			m.terminalHeaderLogic(),
			m.chatHeaderLogic(),
		}
	}

    return table.New().
        Border(lipgloss.NormalBorder()).
        Width(m.scrWidth).
        StyleFunc(table.StyleFunc(func(row, col int) lipgloss.Style {
            return lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
        })).
        Row(headers...).
		Render()
}

