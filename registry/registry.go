package registry

import (
	"fmt"

	"lemon-markup/parser"
)

type Registry struct {
	templates map[string]*parser.TemplateNode
}

func New() *Registry {
	return &Registry{
		templates: make(map[string]*parser.TemplateNode),
	}
}

func (r *Registry) Register(doc *parser.Document) error {
	for name, template := range doc.Templates {
		if _, exists := r.templates[name]; exists {
			return fmt.Errorf("template name collision: '%s' is already defined", name)
		}
		r.templates[name] = template
	}
	return nil
}

func (r *Registry) Get(name string) (*parser.TemplateNode, bool) {
	template, ok := r.templates[name]
	return template, ok
}

func (r *Registry) Exists(name string) bool {
	_, ok := r.templates[name]
	return ok
}

func (r *Registry) All() map[string]*parser.TemplateNode {
	return r.templates
}
