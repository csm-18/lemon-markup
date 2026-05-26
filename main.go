package main

import (
	"fmt"
	"os"

	"lemon-markup/cmd"
)

func main() {
	args := os.Args[1:]

	if err := cmd.Execute(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
