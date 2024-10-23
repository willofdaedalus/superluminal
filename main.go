package main

import (
	"log"
	"willofdaedalus/superluminal/server"
)

func main() {
	s, err := server.CreateServer()
	if err != nil {
		log.Fatal(err)
	}

	s.StartServer()
}
