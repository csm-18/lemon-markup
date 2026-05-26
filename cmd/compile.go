package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lemon-markup/expander"
	"lemon-markup/generator"
	"lemon-markup/lexer"
	"lemon-markup/parser"
	"lemon-markup/registry"
)

func Compile() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	
	// Create dist directory
	distDir := filepath.Join(cwd, "dist")
	if err := os.RemoveAll(distDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean dist directory: %w", err)
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("failed to create dist directory: %w", err)
	}
	
	// Find all .lm files
	var lmFiles []string
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip dist directory
		if strings.Contains(path, filepath.Join(cwd, "dist")) {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".lm") {
			lmFiles = append(lmFiles, path)
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}
	
	if len(lmFiles) == 0 {
		fmt.Println("No .lm files found to compile")
		return nil
	}
	
	// Parse all files first to build global registry
	globalRegistry := registry.New()
	documents := make(map[string]*parser.Document)
	
	for _, filePath := range lmFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filePath, err)
		}
		
		// Lex
		lex := lexer.New(string(content))
		tokens := lex.Tokenize()
		
		// Parse
		p := parser.New(tokens)
		doc, err := p.Parse()
		if err != nil {
			return fmt.Errorf("parse error in %s: %w", filePath, err)
		}
		
		documents[filePath] = doc
		
		// Register templates
		if err := globalRegistry.Register(doc); err != nil {
			return fmt.Errorf("registration error in %s: %w", filePath, err)
		}
	}
	
	// Expand and generate output
	outputCount := 0
	for filePath, doc := range documents {
		// Expand all nodes
		exp := expander.New(globalRegistry)
		expandedDoc, err := exp.Expand(doc)
		if err != nil {
			return fmt.Errorf("expansion error in %s: %w", filePath, err)
		}
		
		// Only generate output for files with DOCTYPE
		if expandedDoc.HasDoctype {
			// Generate HTML
			gen := generator.New()
			html := gen.Generate(expandedDoc)
			
			// Determine output filename
			relPath, _ := filepath.Rel(cwd, filePath)
			relPath = strings.TrimSuffix(relPath, ".lm")
			outputPath := filepath.Join(distDir, relPath+".html")
			
			// Create subdirectories if needed
			outputFileDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputFileDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory for output: %w", err)
			}
			
			// Write output
			if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
				return fmt.Errorf("failed to write output %s: %w", outputPath, err)
			}
			
			fmt.Printf("✓ Compiled: %s -> %s\n", relPath+".lm", relPath+".html")
			outputCount++
		}
	}
	
	fmt.Printf("\n✓ Compilation complete: %d file(s) generated in dist/\n", outputCount)
	return nil
}
