package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildLemonProject() {
	// 1. Discover all active source files using your specialized crawler
	markupFiles, err := CollectAllMarkupFiles(".")
	if err != nil {
		PrintError("Unable to read markup files!")
	}

	if len(markupFiles) == 0 {
		fmt.Println("No .lm files found to compile")
		return
	}

	// 2. Ensure the base target build directory exists without wiping it
	distDir := "dist"
	CreateFolder(distDir)

	// 3. Clean up leftover HTML files from previous runs
	cleanupLeftoverHTML(distDir, markupFiles)

	// 4. Step 1 Pass: Parse all discovered source streams to populate the Global Registry
	globalRegistry := NewRegistry()
	documents := make(map[string]*Document)

	for _, filePath := range markupFiles {
		contentBytes, err := os.ReadFile(filePath)
		if err != nil {
			PrintError("Unable to read file: " + filePath)
		}

		// Execute internal single-package Lexer pass
		srcFile := File{
			name: filePath,
			text: string(contentBytes),
		}
		tokens := Lexer(srcFile)

		// Execute internal single-package Parser pass
		p := NewParser(tokens, filePath)
		doc := p.Parse()

		documents[filePath] = doc

		// Register extracted templates into the global pool
		globalRegistry.Register(doc, filePath)
	}

	// 5. Step 2 Pass: Unfold custom component layers and emit flat vanilla HTML structures
	outputCount := 0
	cwd, _ := os.Getwd()
	absCwd, _ := filepath.Abs(cwd)

	for filePath, doc := range documents {
		// Run recursive components macro expansion algorithm pass
		exp := NewExpander(globalRegistry, filePath)
		expandedDoc := exp.Expand(doc)

		// Only build physical output files for files declared with an explicit <!doctype html> entry
		if expandedDoc.HasDoctype {
			gen := NewGenerator()
			html := gen.Generate(expandedDoc)

			// Safely convert to absolute path before checking relationship to avoid mixing rel/abs errors
			absFilePath, _ := filepath.Abs(filePath)
			relPath, err := filepath.Rel(absCwd, absFilePath)
			if err != nil {
				relPath = filePath
			}
			relPath = strings.TrimSuffix(relPath, ".lm")

			// Strip the "markup/" directory prefix if present, so output maps cleanly into dist/ root
			if strings.HasPrefix(relPath, "markup"+string(filepath.Separator)) {
				relPath = strings.TrimPrefix(relPath, "markup"+string(filepath.Separator))
			} else if strings.HasPrefix(relPath, "markup/") || strings.HasPrefix(relPath, "markup\\") {
				relPath = relPath[7:]
			}

			outputPath := filepath.Join(distDir, relPath+".html")

			// Create target layout directories dynamically if deeply nested
			outputFileDir := filepath.Dir(outputPath)
			if outputFileDir != "." && outputFileDir != distDir {
				CreateFolder(outputFileDir)
			}

			// Save fully compiled template out to disk using your file builder helper
			CreateFile(outputPath, html)

			fmt.Printf("✓ Compiled: %s -> %s\n", filePath, outputPath)
			outputCount++
		}
	}

	fmt.Printf("\n✓ Compilation complete: %d file(s) generated in %s/\n", outputCount, distDir)
}

// Helper function to scan dist/ and safely remove HTML files whose source .lm file no longer exists
func cleanupLeftoverHTML(distDir string, activeSources []string) {
	validHTMLOutputs := make(map[string]bool)
	cwd, _ := os.Getwd()
	absCwd, _ := filepath.Abs(cwd)

	for _, srcPath := range activeSources {
		absFilePath, _ := filepath.Abs(srcPath)
		relPath, err := filepath.Rel(absCwd, absFilePath)
		if err != nil {
			relPath = srcPath
		}
		relPath = strings.TrimSuffix(relPath, ".lm")

		// Apply the same structural prefix adjustment during cleanup matching passes
		if strings.HasPrefix(relPath, "markup"+string(filepath.Separator)) {
			relPath = strings.TrimPrefix(relPath, "markup"+string(filepath.Separator))
		} else if strings.HasPrefix(relPath, "markup/") || strings.HasPrefix(relPath, "markup\\") {
			relPath = relPath[7:]
		}

		expectedHTMLPath := filepath.Join(distDir, relPath+".html")
		validHTMLOutputs[expectedHTMLPath] = true
	}

	// Walk through the dist directory to find and eliminate orphaned HTML files
	err := filepath.Walk(distDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// We only care about checking individual files ending in .html
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			if !validHTMLOutputs[path] {
				err := os.Remove(path)
				if err == nil {
					fmt.Printf("🧹 Removed leftover output: %s\n", path)
				}
			}
		}
		return nil
	})

	if err != nil {
		PrintError("Unable to complete leftover cleanup scan inside dist/")
	}
}

type File struct {
	name string
	text string
}
