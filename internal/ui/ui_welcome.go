package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	hostname  string
	clientNum string
)

func (m *model) startInputsLogic() {
	if m.startCurField == 2 {
		m.startCurField = 1
		m.startInputs[1].Blur()
		m.startInputs[0].Focus()
	} else {
		m.startCurField = 2
		m.startInputs[0].Blur()
		m.startInputs[1].Focus()
	}
}

func readyStartInputs() []textinput.Model {
	inputs := make([]textinput.Model, 0, 2)

	// name input box
	nameInput := textinput.New()
	nameInput.Focus()
	nameInput.CharLimit = 15
	nameInput.Placeholder = "name of session"

	// client number input box
	numInput := textinput.New()
	numInput.Placeholder = "number of clients"
	numInput.CharLimit = 2

	inputs = append(inputs, nameInput, numInput)
	return inputs
}

func (m model) drawInputBox(label string, boxIdx, width int, selected bool) string {
	if boxIdx > len(m.startInputs) {
		return ""
	}

	labelText := lipgloss.NewStyle().
		MarginLeft(1).
		MarginTop(m.scrHeight / 15).
		Bold(selected).
		Render(label)

	textBox := m.startInputs[boxIdx]
	if selected {
		textBox.Focus() // panic happens here
	} else {
		textBox.Blur()
	}

	input := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(width).
		MarginLeft(1).
		Render(textBox.View())

	finalBox := lipgloss.JoinVertical(
		lipgloss.Left,
		labelText,
		input,
	)

	return finalBox
}

func (m model) StartScreenView() string {
	scrWidth := m.scrWidth / 4

	nameInput := m.drawInputBox(
		"name of session (clients see this)",
		0, (scrWidth-5)+1, m.startCurField == 1,
	)

	countInput := m.drawInputBox(
		"max number of clients (1 - 32)",
		1, (scrWidth-5)+1, m.startCurField == 2,
	)

	terminalView := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(scrWidth).
		Height(m.scrHeight/3).
		Render(nameInput, countInput)

	scr := lipgloss.Place(
		m.scrWidth, m.scrHeight,
		lipgloss.Center, lipgloss.Center,
		terminalView,
	)

	return scr
}
