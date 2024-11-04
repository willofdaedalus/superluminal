package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"willofdaedalus/superluminal/internal/backend"
	"willofdaedalus/superluminal/internal/client"

	"github.com/charmbracelet/huh"
)

var (
	startServer       bool
	defaultConnection string
)

func init() {
	flag.StringVar(&defaultConnection, "c", "", "the host and port to connect to")
	flag.BoolVar(&startServer, "s", false, "start a superluminal session server")
	flag.Parse()
}

func validateClientNum(in string) error {
	if in == "" {
		return nil
	}

	v, err := strconv.Atoi(in)
	if err != nil || (v <= 0 || v > 32) {
		return fmt.Errorf("please enter a number between 1 and 32")
	}
	return err
}

func main() {
	if startServer {
		var name string
		var maxConns string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("what's your name?").
					Prompt("> ").
					CharLimit(25).
					Value(&name),
			).WithTheme(huh.ThemeBase16()),
			huh.NewGroup(
				huh.NewInput().
					Title("how many maximum clients? (not adjustable once session starts)").
					Prompt("> ").
					Validate(validateClientNum).
					Placeholder("1-32").
					Value(&maxConns),
			).WithTheme(huh.ThemeBase16()),
		)

		err := form.Run()
		if err != nil {
			log.Fatal(err)
		}

		v, _ := strconv.Atoi(maxConns)

		s, err := backend.CreateServer(name, v)
		if err != nil {
			log.Fatal(err)
		}

		s.Run()
	} else {
		var name, pass string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("what's your name?").
					Prompt("> ").
					CharLimit(25).
					Value(&name),
			).WithTheme(huh.ThemeBase16()),
			huh.NewGroup(
				huh.NewInput().
					Title("enter the passphrase for validation").
					Prompt("> ").
					Value(&pass),
			).WithTheme(huh.ThemeBase16()),
		)
		err := form.Run()
		if err != nil {
			log.Fatal(err)
		}

		c := client.CreateClient(name, pass)
		c.ConnectToServer("localhost", "42024")
	}
}
