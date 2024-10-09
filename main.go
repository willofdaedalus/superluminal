package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
		// connect to the server if the -c flag is provided
		ctx := context.Background()

		err := client.ConnectToServer(ctx)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("You must either provide '-start' to run the server or '-c' to connect to one.")
	}
}
