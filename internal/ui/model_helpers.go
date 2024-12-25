package ui

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func (m model) mainRender() string {
    scrWidth := m.scrWidth - 2
    if scrWidth < 0 {
        scrWidth = 0
    }

    headerRender := m.HeaderView() 
    terminalView := lipgloss.NewStyle().
        Border(lipgloss.NormalBorder()).
        BorderTop(false).
        Width(scrWidth).
        Height(m.scrHeight - lipgloss.Height(headerRender) - 1).
        Render("no terminal content") // NOTE; pass the content

	// without the following changes there's an ugly gap between the headers
	// and the terminalView
    headerRenderModified := strings.ReplaceAll(headerRender, "└", "├")
    headerRenderModified = strings.ReplaceAll(headerRenderModified, "┘", "┤")

    cFinalRender := lipgloss.JoinVertical(
        lipgloss.Bottom,
        headerRenderModified,
        terminalView,
    )
    return cFinalRender
}

func NewModel() model {
	clients := make([]clientEntry, 32)

	return model{
		clientList: clients,
		hostSide:   true,
	}
}
