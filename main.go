package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"willofdaedalus/superluminal/client"
	"willofdaedalus/superluminal/config"
	"willofdaedalus/superluminal/server"
)

var startServer bool

func init() {
	flag.StringVar(&config.DefaultConnection, "c", "", "the host and port to connect to")
	flag.BoolVar(&startServer, "start", false, "start the server")
	flag.Parse()
}

func main() {
	if startServer {
		// start the server if the -start flag is provided
		s, err := server.CreateServer()
		if err != nil {
			log.Fatal(err)
		}
		s.StartServer()
	} else if config.DefaultConnection != "" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter your passphrase: ")
		pass, err := reader.ReadString('\n') // Read until newline
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
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
