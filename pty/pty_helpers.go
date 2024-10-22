package pty

import (
	"fmt"
	"github.com/creack/pty"
	"os"
	"os/exec"
)

// gets the user's default shell; if not successful, falls back to bash
func getUserShell() string {
	sh := os.Getenv("SHELL")
	if sh == "" {
		// if the shell variable couldn't be returned just
		// use the default bash which is probably definitely installed
		sh = "/bin/bash"
	}

	return sh
}

// creates and returns a new pty session that reads from os.Stdin
func createSession() (*os.File, error) {
	sh := exec.Command(getUserShell())

	ptmx, err := pty.Start(sh)
	if err != nil {
		return nil, err
	}

	if err = pty.InheritSize(os.Stdin, ptmx); err != nil {
		fmt.Println("couldn't resize pty:", err)
	}

	return ptmx, nil
}
