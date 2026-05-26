package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func Init() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	
	markupDir := filepath.Join(cwd, "markup")
	
	// Create markup directory
	if err := os.MkdirAll(markupDir, 0755); err != nil {
		return fmt.Errorf("failed to create markup directory: %w", err)
	}
	
	// Create hello page example
	helloFilePath := filepath.Join(markupDir, "hello.lm")
	helloContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hello Lemon Markup</title>
</head>
<body>
    <Header Title="Welcome to Lemon Markup"></Header>
    <Main Content="This is a simple hello page built with Lemon Markup."></Main>
</body>
</html>

<template name="Header">
    <header style="background-color: #333; color: white; padding: 20px; text-align: center;">
        <h1>{{ Title }}</h1>
    </header>
</template>

<template name="Main">
    <main style="padding: 40px; max-width: 800px; margin: 0 auto;">
        <p>{{ Content }}</p>
        <p>Lemon Markup is a static HTML compiler with reusable components.</p>
    </main>
</template>
`
	
	if err := os.WriteFile(helloFilePath, []byte(helloContent), 0644); err != nil {
		return fmt.Errorf("failed to create hello.lm: %w", err)
	}
	
	// Create a README in markup folder
	readmePath := filepath.Join(markupDir, "README.md")
	readmeContent := `# Markup Files

This directory contains Lemon Markup files (.lm files).

## Compiling

Run 'lm compile' from the project root to compile all markup files.

The compiled HTML files will be generated in the 'dist/' directory.

## Example

- hello.lm - A simple hello page with Header and Main components
`
	
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	
	fmt.Printf("✓ Created markup folder at: %s\n", markupDir)
	fmt.Printf("✓ Created example: hello.lm\n")
	fmt.Println("\nTo compile your markup files, run: lm compile")
	
	return nil
}
