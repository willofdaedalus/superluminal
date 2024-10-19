package term

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"
)

func getUserShell() string {
	shellPath := os.Getenv("SHELL")
	return filepath.Base(shellPath)
}

// create and return a pty as an *os.File
func CreatePtySession() (*os.File, error) {
	sh := exec.Command(getUserShell())
	sh.Env = append(sh.Env, "SPRLMNL=1")

	// pts, err := pty.StartWithSize(sh, &pty.Winsize{})
	pts, err := pty.Start(sh)
	if err != nil {
		return nil, err
	}

	if err := pty.InheritSize(os.Stdin, pts); err != nil {
		return nil, err
	}
	return pts, nil
}
