package main

import (
	"bytes"
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

func CopyFolder(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyFolder(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func FolderSync(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return syncFolder(src, dst)
}

func syncFolder(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := syncFolder(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := syncFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func syncFile(src, dst string) error {
	if _, err := os.Stat(dst); err != nil {
		if os.IsNotExist(err) {
			return copyFile(src, dst)
		}
		return err
	}

	equal, err := filesAreEqual(src, dst)
	if err != nil {
		return err
	}

	if !equal {
		return copyFile(src, dst)
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

func filesAreEqual(a, b string) (bool, error) {
	aData, err := os.ReadFile(a)
	if err != nil {
		return false, err
	}

	bData, err := os.ReadFile(b)
	if err != nil {
		return false, err
	}

	return bytes.Equal(aData, bData), nil
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
