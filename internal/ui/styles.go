package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func bold(text string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Render(text)
}

func normal(text string) string {
	return lipgloss.NewStyle().
		Render(text)
}
