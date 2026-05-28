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
			CreateLemonProject()
		} else if args[0] == "build" {

		} else if args[0] == "version" {
			fmt.Println(LEMON_VERSION)
		} else if args[0] == "help" {
			fmt.Println(LEMON_HELP)

		} else {
			PrintError(LEMON_CLI_ERROR)
		}
	} else {
		PrintError(LEMON_CLI_ERROR)
	}

}

const LEMON_VERSION = "lemon 0.2.0"
const LEMON_ABOUT = `Lemon Markup is a simple, strict, compile-time-only static HTML extension language
For help:
  lemon help`

const LEMON_HELP = `lemon Commands:
  1. lemon <no args>          -> print about message
  2. lemon init               -> initialize new lemon project
  3. lemon build              -> build the project
  4. lemon version            -> print lemon version
  5. lemon help               -> print lemon commands list`
const LEMON_CLI_ERROR = "Unknown command or malformed args passed to lemon!"
