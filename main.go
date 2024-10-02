package main

import (
	"log"
	"willofdaedalus/superluminal/internal/server"
)

func main() {
	s, err := server.CreateServer()
	if err != nil {
		log.Fatal(err)
	}

	s.StartServer()
}
