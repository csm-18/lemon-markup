package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println(LEMON_VERSION)
		fmt.Println(LEMON_ABOUT)
	} else if len(args) == 1 {
		if args[0] == "init" {

		} else if args[0] == "build" {

		} else if args[0] == "version" {

		} else if args[0] == "help" {

		} else {
			fmt.Println(LEMON_CLI_ERROR)
			os.Exit(0)
		}
	} else {
		fmt.Println(LEMON_CLI_ERROR)
		os.Exit(0)
	}

}

const LEMON_VERSION = "lemon 0.2.0"
const LEMON_ABOUT = `Lemon Markup is a simple, strict, compile-time-only static HTML extension language
For help:
  lemon help`
const LEMON_CLI_ERROR = "Error: Unknown command or malformed args passed to lemon!"
