package ui

import (
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/lipgloss"
)

func (m model) sessionHeaderLogic() string {
	text := normal("session")
	if m.currentTab == 2 {
		text = bold("[session]")
	}

	return text
}

func (m model) chatHeaderLogic() string {
	text := "chat"
	if m.currentTab == 1 {
		text = "[chat]"
	}

	return text
}

func (m model) terminalHeaderLogic() string {
	text := "terminal"
	if m.currentTab == 0 {
		text = "[terminal]"
	}

	return text
}

func (m model) HeaderView() string {
	headers := []string{
		m.terminalHeaderLogic(),
		m.chatHeaderLogic(),
		m.sessionHeaderLogic(),
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
			// align all the header text to the center of their respective boxes
            return lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
        })).
        Row(headers...).
		Render()
}

