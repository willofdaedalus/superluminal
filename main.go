package main

import (
	"log"
	"willofdaedalus/superluminal/internal/server"
)

func main() {
	// if len(os.Args) < 2 {
	// 	fmt.Println("incorrect number of args. see help for more information")
	// 	return
	// }
	//
	// cli.ParseArgs(os.Args[1], os.Args[2])

	s, err := server.CreateServer()
	if err != nil {
		log.Fatal(err)
	}

	s.StartServer()
}
