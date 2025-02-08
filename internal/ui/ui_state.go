package ui

import (
	"willofdaedalus/superluminal/internal/backend"
	"willofdaedalus/superluminal/internal/client"
)

// state is an object that contains configuration and options
// that are necessary for the ui components to render depending
// on whether it's a client or the host
// I noticed I was doing a lot checks for hostSide and want to
// keep things clean and nice
type state struct {
	// general app wide options
	tabCount int

	// initial screen options
	charLimit              int
	firstInputPlaceholder  string
	secondInputPlaceholder string
	firstInputLabel        string
	secondInputLabel       string
	firstInputErrMsg       string
	secondInputErrMsg      string
	initErrMessage         string
	nameBlankErrMsg        string
	clientObj              *client.Client
	session                *backend.Session
}
