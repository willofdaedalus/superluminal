package ui

import (
	"willofdaedalus/superluminal/internal/backend"
	"willofdaedalus/superluminal/internal/client"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	startView = iota + 1
	mainView
)

const (
	headerSwitch = "alt+esc"
)

func NewModel(hostSide bool) (*model, error) {
	var c *client.Client
	var session *backend.Session
	var err error
	var appState *state

	if hostSide {
		// use a temporary name here which we'll later change when the user
		// submits their own name with the passphrase to connect to the server
		// we're using a maxconns of 2 which we'll change later by resizing based
		// on user input in the form fields
		session, err = backend.NewSession("hello", 2)
		if err != nil {
			return nil, err
		}

		appState = &state{
			charLimit:              2,
			firstInputPlaceholder:  "session name",
			secondInputPlaceholder: "number of clients",
			firstInputLabel:        "name of your session (clients see this)",
			secondInputLabel:       "number of clients (1-32)",
			firstInputErrMsg:       "name cannot be blank",
			secondInputErrMsg:      "enter a valid number between 1 and 32",
			session:                session,
			tabCount:               hostMaxTabs,
		}
	} else {
		// use a temporary name here which we'll later change when the user
		// submits their own name with the passphrase to connect to the server
		c = client.New("temp-name")

		// addr := "localhost:42024"
		// if len(os.Args) > 1 {
		// 	addr = os.Args[1]
		// }

		// err = c.ConnectToSession(addr)
		// if err != nil {
		// 	return nil, err
		// }

		appState = &state{
			charLimit:              65,
			firstInputPlaceholder:  "your name",
			secondInputPlaceholder: "passphrase",
			firstInputLabel:        "your name (hosts see this)",
			secondInputLabel:       "passphrase for session (ask the host)",
			firstInputErrMsg:       "name cannot be blank",
			secondInputErrMsg:      "password cannot be blank",
			clientObj:              c,
			tabCount:               clientMaxTabs,
		}
	}

	return &model{
		view:          startView,
		startCurField: 1,
		clientList:    make([]clientEntry, 32),
		hostSide:      hostSide,
		currentView:   termView,
		startInputs:   readyStartInputs(hostSide),
		appState:      appState,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// handle input updates
	if m.view == startView && m.startCurField <= len(m.startInputs) {
		m.startInputs[m.startCurField-1], cmd = m.startInputs[m.startCurField-1].Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.view == startView {
				if err := m.validateStartInputs(); err != nil {
					m.showErrMsg = true
					return m, nil
				}

				m.transitionView(mainView)

				// if m.hostSide {
				// 	m.appState.session.Start()
				// 	return m, nil
				// } else {
				// 	m.clientErrChan = make(chan error, 1)
				// 	go func() {
				// 		m.appState.clientObj.ListenForMessages(m.clientErrChan)
				// 	}()

				// 	// Handle errors and shutdown
				// 	for {
				// 		select {
				// 		case err, ok := <-m.clientErrChan:
				// 			if !ok {
				// 				// this is actually an application exit
				// 				return m, nil
				// 			}
				// 			if err != nil {
				// 				fmt.Println("got something")
				// 				log.Println(err)
				// 			}
				// 		}
				// 	}
				// }
			}

		case "tab":
			if m.view == startView {
				m.switchStartInput()
			}

		case headerSwitch:
			m.switchTab()
		}
	case tea.WindowSizeMsg:
		m.scrWidth = msg.Width
		m.scrHeight = msg.Height
	}
	return m, cmd
}

func (m model) View() string {
	switch m.view {
	case startView:
		return m.StartScreenView()
	case mainView:
		return m.mainRender()
	default:
		return ""
	}
	// return m.startScreen()
	// return m.mainRender()
	// return m.HeaderView()
}
