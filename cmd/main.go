package cmd

import (
	"fmt"
	"os"
)

const Version = "0.1.1"

func Execute(args []string) error {
	if len(args) == 0 {
		return Help()
	}

	command := args[0]

	switch command {
	case "init":
		return Init()
	case "compile":
		return Compile()
	case "version":
		return Version_()
	case "help":
		return Help()
	case "--version", "-v":
		return Version_()
	case "--help", "-h":
		return Help()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Use 'lm help' for available commands\n")
		return fmt.Errorf("unknown command: %s", command)
	}
}

func Version_() error {
	fmt.Printf("lm version %s\n", Version)
	return nil
}

func Help() error {
	helpText := `Lemon Markup Compiler v` + Version + `

Usage: lm <command> [options]

Commands:
  init       Create a new markup directory with a hello page example
  compile    Compile all .lm files in the current directory to dist/
  version    Print the version number
  help       Show this help message

Examples:
  lm init           Create markup folder with example
  lm compile        Compile markup files to dist/
  lm version        Show version
  lm help           Show this message

Flags:
  -v, --version     Show version
  -h, --help        Show help
`
	fmt.Print(helpText)
	return nil
}
