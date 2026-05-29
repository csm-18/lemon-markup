package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

// recursively find all ".lm" files starting from the given path, and ignore folders that are not necessary
func CollectAllMarkupFiles(rootPath string) ([]string, error) {
	var files []string

	// Define directories to completely ignore
	ignoredDirs := map[string]bool{
		"dist":   true,
		"assets": true,
		"styles": true,
		"logic":  true,
	}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// If it's a directory and its name is in our ignore list, skip it entirely
		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the file has the ".lm" extension
		if filepath.Ext(d.Name()) == ".lm" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
