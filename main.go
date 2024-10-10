package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"willofdaedalus/superluminal/client"
	"willofdaedalus/superluminal/config"
	"willofdaedalus/superluminal/server"
)

var (
	reader *bufio.Reader = bufio.NewReader(os.Stdin)
	startServer bool
	numOfTries  int = 3
)

func init() {
	flag.StringVar(&config.DefaultConnection, "c", "", "the host and port to connect to")
	flag.BoolVar(&startServer, "start", false, "start a superluminal session server")
	flag.Parse()
}

func getInitialInput(prompt string) (string, error) {
	fmt.Printf("Enter your %s: ", prompt)
	pass, err := reader.ReadString('\n') // Read until newline
	if err != nil {
		return "", fmt.Errorf("Error reading input: %v", err)
	}

	return pass, nil
}

func getNumberOfClients() (int, error) {
	users := 0
	for numOfTries > 0 {
		fmt.Print("Number of clients (1-32): ")
		count, err := reader.ReadString('\n')
		if err != nil {
			return -1, fmt.Errorf("Error reading input: %v", err)
		}

		users, err = strconv.Atoi(strings.TrimSpace(count))
		if err != nil || (users <= 0 || users > 32) {
			fmt.Println("enter a valid number between 1-32")
			numOfTries -= 1
			continue
		}
		return users, nil
	}

	return -1, fmt.Errorf("enter a valid number between 1-32")
}

func main() {
	if startServer {
		name, err := getInitialInput("name")
		if err != nil {
			log.Fatal(err)
		}

		maxConns, err := getNumberOfClients()
		if err != nil {
			return
		}

		// start the server if the -start flag is provided
		s, err := server.CreateServer(name, maxConns)
		if err != nil {
			log.Fatal(err)
		}
		s.StartServer()
	} else if config.DefaultConnection != "" {
		pass, err := getInitialInput("passphrase")
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		err = client.ConnectToServer(ctx, pass)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("You must either provide '-start' to run the server or '-c' to connect to one.")
	}
}
