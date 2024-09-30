package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"willofdaedalus/superluminal/internal/utils"
)

func main() {
	var read string
	pass, err := utils.GeneratePassphrase()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("your passphrase is %s\n", pass)

	hashed, err := utils.HashPassphrase(pass)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	chances := 3

	for chances > 0 {
		fmt.Printf("enter passphrase to enter session: ")
		if scanner.Scan() {
			read = scanner.Text() // Get the whole line including spaces
		}

		if !utils.CheckPassphrase(hashed, read) {
			fmt.Println("wrong passphrase")
			chances--
		} else {
			fmt.Println("welcome")
			return
		}
	}

}
