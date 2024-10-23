package main

import (
	"log"
	"sync"
	"willofdaedalus/superluminal/client"
	"willofdaedalus/superluminal/server"
)

func main() {
	var wg sync.WaitGroup

	s, err := server.CreateServer()
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go s.Start()
	c := client.CreateClient("tony")
	go c.ConnectToServer("localhost", "42024")
	wg.Wait()
}
