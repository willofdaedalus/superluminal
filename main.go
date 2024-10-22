package main

import (
	"fmt"
	"log"
	"willofdaedalus/superluminal/pty"
)

func main() {
	t, err := pty.NewTether()
	if err != nil {
		log.Fatal(err)
	}

	t.WriteTo([]byte("ls\n"))
	fmt.Print(string(t.ReadFrom()))
}
