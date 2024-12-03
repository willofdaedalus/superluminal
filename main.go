package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"willofdaedalus/superluminal/internal/backend"
	"willofdaedalus/superluminal/internal/client"

	"golang.org/x/term"
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

// TODO; remember to disable signal processing for bubbletea
func main() {
	if startServer {
		session, err := backend.NewSession("hello", 5)
		if err != nil {
			log.Fatal(err.Error())
		}

		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
		session.Start()

	} else {
		errChan := make(chan error, 1)
		ctx := context.Background()
		client := client.New("hello")
		// err := client.ConnectToSession(ctx, "localhost", "42024")
		err := client.ConnectToSession(ctx, os.Args[1])

		if err != nil {
			log.Fatal(err.Error())
		}

		go func() {
			client.ListenForMessages(errChan)
		}()

		// Handle errors and shutdown
		for err := range errChan {
			// if errors.Is(err, utils.ErrServerFull) {
			// 	log.Fatal(err.Error())
			// }
			log.Println(err)
			break
		}
	}
}
