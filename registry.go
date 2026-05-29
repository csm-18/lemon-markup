package main

import (
	"fmt"
	"os"
)

type Registry struct {
	templates map[string]*TemplateNode
}

func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[string]*TemplateNode),
	}
}

// Register adds templates from a Document to the global registry.
// If a template name is already claimed by another file, it throws a fatal error and halts.
func (r *Registry) Register(doc *Document, fileName string) {
	for name, template := range doc.Templates {
		if _, exists := r.templates[name]; exists {
			fmt.Printf("Syntax Error: Global template name collision! '%s' is already defined in another file.\n", name)
			fmt.Printf("  --> %s (Tried to re-declare here)\n", fileName)
			os.Exit(0)
		}
		r.templates[name] = template
	}
}

func (r *Registry) Get(name string) (*TemplateNode, bool) {
	template, ok := r.templates[name]
	return template, ok
}

func (r *Registry) Exists(name string) bool {
	_, ok := r.templates[name]
	return ok
}

func (r *Registry) All() map[string]*TemplateNode {
	return r.templates
}
