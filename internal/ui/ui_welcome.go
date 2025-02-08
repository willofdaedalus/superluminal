package ui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	hostname  string
	clientNum string
)

func setupHostSide(m *model, num int) {
	m.sessClientCount = uint8(num)
	m.appState.session.SetMaxConns(uint8(num))
	// set the session name here based on the user's input
	// we update this everytime in the event the user changes their mind
	// about what to call the session; once we begin though it should be
	// impossible to change the sesion name
	m.appState.session.Owner = m.startInputs[0].Value()
}

func setupClientSide(m *model, toValidate string) {
	// if the user sends the wrong pass, the server reacts by resending
	// the auth prompt which sets the client's sentpass value; using that
	// we can print the error message to the user without explicitly
	// checking whether the function that handles auth on the client returns
	// good or bad
	m.showErrMsg = m.appState.clientObj.SentPass
	// pass the text field text to the client via a channel
	m.appState.clientObj.SendPassphrase(toValidate)
	m.appState.clientObj.SetName(m.startInputs[0].Value())
}

// validates the user input by checking for the expected values and such
// while also filling the right fields in the model with the user submitted
// information for a seamless and smooth transition
func (m *model) validateStartInputs() error {
	m.showErrMsg = false
	firstInput := m.startInputs[0].Value()
	if len(firstInput) == 0 {
		m.setErrorMessage(m.appState.firstInputErrMsg)
		return fmt.Errorf("wrong input")
	}

	toValidate := m.startInputs[1].Value()
	if len(toValidate) == 0 {
		m.setErrorMessage(m.appState.secondInputErrMsg)
		return fmt.Errorf("wrong input")
	}

	if m.hostSide {
		num, err := strconv.Atoi(toValidate)
		if num < 1 || num > 32 {
			m.setErrorMessage(m.appState.secondInputErrMsg)
			return fmt.Errorf("wrong input")
		}

		if err != nil {
			return err
		}

		setupHostSide(m, num)
		return nil
	}

	setupClientSide(m, toValidate)
	return nil
}

func (m *model) switchStartInput() {
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

func readyStartInputs(hostSide bool) []textinput.Model {
	inputs := make([]textinput.Model, 0, 2)
	limit := 2
	namePlaceholder := "name of session"
	numberPlaceholder := "number of clients"

	if !hostSide {
		namePlaceholder = "your name"
		numberPlaceholder = "passphrase"
		// this is assuming diceware chooses to use 5 * 12 char words
		// +1 for sane divisions
		limit = 65
	}

	// name input box
	nameInput := textinput.New()
	nameInput.Focus()
	nameInput.CharLimit = 15
	nameInput.Placeholder = namePlaceholder

	// client number input box
	numInput := textinput.New()
	numInput.Placeholder = numberPlaceholder
	numInput.CharLimit = limit

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
		textBox.Focus()
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

	errText := ""
	// show the errmsg on start
	if m.showErrMsg {
		errText = m.errMsg
	}

	nameInput := m.drawInputBox(
		m.appState.firstInputLabel,
		0, (scrWidth-5)+1, m.startCurField == 1,
	)

	countInput := m.drawInputBox(
		m.appState.secondInputLabel,
		1, (scrWidth-5)+1, m.startCurField == 2,
	)

	errBox := lipgloss.NewStyle().
		Border(lipgloss.HiddenBorder()).
		Width(scrWidth - 5).
		Foreground(lipgloss.Color("#ff0000")).
		Render(errText)

	terminalView := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		MarginTop(2).
		Width(scrWidth).
		Height(m.scrHeight/3).
		Render(fmt.Sprintf("\033[31m%s\033[0m", nameInput), countInput, errBox)

	scr := lipgloss.Place(
		m.scrWidth, m.scrHeight,
		lipgloss.Center, lipgloss.Center,
		terminalView,
	)

	return scr
}
