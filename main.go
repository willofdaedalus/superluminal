package main

import (
	"flag"
	"fmt"
	"strconv"
)

var (
	startServer       bool
	defaultConnection string
)

func init() {
	flag.StringVar(&defaultConnection, "c", "", "the host and port to connect to")
	flag.BoolVar(&startServer, "s", false, "start a superluminal session server")
	flag.Parse()
}

func validateClientNum(in string) error {
	if in == "" {
		return nil
	}

	v, err := strconv.Atoi(in)
	if err != nil || (v <= 0 || v > 32) {
		return fmt.Errorf("please enter a number between 1 and 32")
	}
	return err
}

// TODO; remember to disable signal processing for bubbletea
func main() {
	fmt.Println("hello world")
}
