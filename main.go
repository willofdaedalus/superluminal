package main

import (
	"log"
	"time"
	"willofdaedalus/superluminal/client"
	"willofdaedalus/superluminal/server"
)

func main() {
	s, err := server.CreateServer()
	if err != nil {
		log.Fatal(err)
	}

	go s.Start()
	c := client.CreateClient("tony")
	go c.ConnectToServer("localhost", "42024")
	<-time.After(time.Second * 7)
}
