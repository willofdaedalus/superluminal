package ui

import (
	// "github.com/charmbracelet/bubbles/viewport"
)

const (
	hostMaxTabs = 3
	clientMaxTabs = 2
)

func NewModel() model {
	clients := make([]clientEntry, 32)

	return model{
		clientList: clients,
		hostSide:   true,
	}
}

func (m *model) switchTab() {
	tabCount := hostMaxTabs - 1

	if !m.hostSide {
		tabCount = clientMaxTabs - 1
	} 

	if m.currentTab == tabCount {
		m.currentTab = 0
		return
	}

	m.currentTab += 1
}
