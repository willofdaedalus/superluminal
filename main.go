package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"willofdaedalus/superluminal/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
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
	model, err := ui.NewModel(startServer)
	if err != nil {
		log.Fatalf("couldn't connect to server because of new client")
	}

	p := tea.NewProgram(*model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// if startServer {
	// 	session, err := backend.NewSession("hello", 5)
	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}

	// 	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer term.Restore(int(os.Stdin.Fd()), oldState)
	// 	session.Start()

	// } else {
	// 	errChan := make(chan error, 1)
	// 	ctx := context.Background()
	// 	client := client.New("hello")
	// 	addr := "localhost:42024"
	// 	if len(os.Args) > 1 {
	// 		addr = os.Args[1]
	// 	}

	// 	// err := client.ConnectToSession("localhost", "42024")
	// 	err := client.ConnectToSession(addr)
	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}

	// 	fmt.Println("connected to server...")

	// 	go func() {
	// 		client.ListenForMessages(errChan)
	// 	}()

	// 	// Handle errors and shutdown
	// 	for {
	// 		select {
	// 		case err, ok := <-errChan:
	// 			if !ok {
	// 				fmt.Println("error channel closed")
	// 				return
	// 			}
	// 			if err != nil {
	// 				fmt.Println("got something")
	// 				log.Println(err)
	// 			}
	// 		}
	// 	}
	// }
}
