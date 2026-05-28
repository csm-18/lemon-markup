package main

import (
	"fmt"
	"os"
)

func PrintError(message string) {
	fmt.Println("Error:", message)
	os.Exit(0)
}

func ProjectFolderIsEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		println("Error reading directory:", err.Error())
		os.Exit(0)
	}

	for _, entry := range entries {
		name := entry.Name()

		//ignore these
		if name == ".git" ||
			name == "README.md" ||
			name == "LICENSE.txt" {
			continue
		}

		//any other file/folder means not empty
		return false
	}

	return true
}

func CreateFile(name string, content string) {
	err := os.WriteFile(name, []byte(content), 0644)
	if err != nil {
		PrintError("Unable to create file: " + name)
	}
}

func CreateFolder(name string) {
	err := os.MkdirAll(name, 0755)
	if err != nil {
		PrintError("Unable to create folder: " + name)
	}
}
