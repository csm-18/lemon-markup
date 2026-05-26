package expander

import (
	"fmt"
	"unicode"

	"lemon/parser"
	"lemon/registry"
)

type Expander struct {
	registry *registry.Registry
}

type visitedSet map[string]bool

func New(reg *registry.Registry) *Expander {
	return &Expander{
		registry: reg,
	}
}

func (e *Expander) Expand(doc *parser.Document) (*parser.Document, error) {
	// Check for circular references
	if err := e.checkCircularReferences(); err != nil {
		return nil, err
	}
	
	// Expand all nodes
	expanded := &parser.Document{
		Nodes:      make([]*parser.Node, 0),
		Templates:  doc.Templates,
		HasDoctype: doc.HasDoctype,
	}
	
	for _, node := range doc.Nodes {
		expandedNode, err := e.expandNode(node, make(visitedSet))
		if err != nil {
			return nil, err
		}
		if expandedNode != nil {
			expanded.Nodes = append(expanded.Nodes, expandedNode)
		}
	}
	
	return expanded, nil
}

func (e *Expander) expandNode(node *parser.Node, visited visitedSet) (*parser.Node, error) {
	if node == nil {
		return nil, nil
	}
	
	switch node.Type {
	case parser.NodeText:
		return node, nil
	
	case parser.NodeVariable:
		return node, nil
	
	case parser.NodeRaw:
		// Don't expand inside raw blocks
		return node, nil
	
	case parser.NodeElement:
		// Check if it's a custom component (uppercase)
		if len(node.Tag) > 0 && unicode.IsUpper(rune(node.Tag[0])) {
			return e.expandComponent(node, visited)
		}
		
		// Expand children of native elements
		newNode := &parser.Node{
			Type:     node.Type,
			Tag:      node.Tag,
			Attrs:    copyAttrs(node.Attrs),
			Children: make([]*parser.Node, 0),
			Line:     node.Line,
			Col:      node.Col,
		}
		
		for _, child := range node.Children {
			expanded, err := e.expandNode(child, visited)
			if err != nil {
				return nil, err
			}
			if expanded != nil {
				newNode.Children = append(newNode.Children, expanded)
			}
		}
		
		return newNode, nil
	
	default:
		return node, nil
	}
}

func (e *Expander) expandComponent(node *parser.Node, visited visitedSet) (*parser.Node, error) {
	componentName := node.Tag
	
	// Check for circular reference
	if visited[componentName] {
		return nil, fmt.Errorf("circular reference detected: %s at %d:%d", componentName, node.Line, node.Col)
	}
	
	template, exists := e.registry.Get(componentName)
	if !exists {
		return nil, fmt.Errorf("undefined component: %s at %d:%d", componentName, node.Line, node.Col)
	}
	
	// Validate props
	if err := e.validateProps(node, template); err != nil {
		return nil, fmt.Errorf("%w at %d:%d", err, node.Line, node.Col)
	}
	
	// Create visited set for this branch
	newVisited := make(visitedSet)
	for k, v := range visited {
		newVisited[k] = v
	}
	newVisited[componentName] = true
	
	// Expand template content with props
	result, err := e.expandTemplateContent(template.Content, node.Attrs, node.Children, newVisited)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

func (e *Expander) expandTemplateContent(templateContent *parser.Node, props map[string]string, children []*parser.Node, visited visitedSet) (*parser.Node, error) {
	if len(templateContent.Children) == 0 {
		return &parser.Node{
			Type:     parser.NodeElement,
			Tag:      "div",
			Children: make([]*parser.Node, 0),
		}, nil
	}
	
	// For single root template, return it; for multiple, wrap in fragment
	result := &parser.Node{
		Type:     parser.NodeElement,
		Tag:      "template-result",
		Children: make([]*parser.Node, 0),
	}
	
	for _, child := range templateContent.Children {
		expanded, err := e.expandTemplateNode(child, props, children, visited)
		if err != nil {
			return nil, err
		}
		if expanded != nil {
			result.Children = append(result.Children, expanded)
		}
	}
	
	// If single child, return it directly; otherwise return wrapper
	if len(result.Children) == 1 {
		return result.Children[0], nil
	}
	
	return result, nil
}

func (e *Expander) expandTemplateNode(node *parser.Node, props map[string]string, children []*parser.Node, visited visitedSet) (*parser.Node, error) {
	if node == nil {
		return nil, nil
	}
	
	switch node.Type {
	case parser.NodeText:
		return node, nil
	
	case parser.NodeVariable:
		varName := node.VarName
		
		if varName == "children" {
			// Inject children
			if len(children) == 0 {
				return nil, nil
			}
			// Return children as a fragment
			if len(children) == 1 {
				return e.expandNode(children[0], visited)
			}
			// Multiple children - wrap in fragment
			fragment := &parser.Node{
				Type:     parser.NodeElement,
				Tag:      "fragment",
				Children: make([]*parser.Node, 0),
			}
			for _, child := range children {
				expanded, err := e.expandNode(child, visited)
				if err != nil {
					return nil, err
				}
				if expanded != nil {
					fragment.Children = append(fragment.Children, expanded)
				}
			}
			return fragment, nil
		}
		
		// Regular prop interpolation
		value, ok := props[varName]
		if !ok {
			return nil, fmt.Errorf("missing required prop: %s", varName)
		}
		
		return &parser.Node{
			Type:    parser.NodeText,
			Content: value,
		}, nil
	
	case parser.NodeRaw:
		return node, nil
	
	case parser.NodeElement:
		// Check if uppercase (custom component)
		if len(node.Tag) > 0 && unicode.IsUpper(rune(node.Tag[0])) {
			return e.expandComponent(node, visited)
		}
		
		// Native element - expand children
		newNode := &parser.Node{
			Type:     node.Type,
			Tag:      node.Tag,
			Attrs:    copyAttrs(node.Attrs),
			Children: make([]*parser.Node, 0),
			Line:     node.Line,
			Col:      node.Col,
		}
		
		for _, child := range node.Children {
			expanded, err := e.expandTemplateNode(child, props, children, visited)
			if err != nil {
				return nil, err
			}
			if expanded != nil {
				newNode.Children = append(newNode.Children, expanded)
			}
		}
		
		return newNode, nil
	
	default:
		return node, nil
	}
}

func (e *Expander) validateProps(componentNode *parser.Node, template *parser.TemplateNode) error {
	// Check for missing required props
	for propName := range template.UsedVars {
		if propName == "children" {
			continue // children is always available
		}
		if _, ok := componentNode.Attrs[propName]; !ok {
			return fmt.Errorf("missing required prop '%s' for component '%s'", propName, componentNode.Tag)
		}
	}
	
	// Check for unused props
	for propName := range componentNode.Attrs {
		if _, ok := template.UsedVars[propName]; !ok {
			return fmt.Errorf("unused prop '%s' passed to component '%s'", propName, componentNode.Tag)
		}
	}
	
	return nil
}

func (e *Expander) checkCircularReferences() error {
	for name := range e.registry.All() {
		visited := make(visitedSet)
		if err := e.checkCircular(name, visited); err != nil {
			return err
		}
	}
	return nil
}

func (e *Expander) checkCircular(name string, visited visitedSet) error {
	if visited[name] {
		return fmt.Errorf("circular reference detected: %s", name)
	}
	
	visited[name] = true
	template, _ := e.registry.Get(name)
	
	// Check all components referenced in template
	components := e.findComponents(template.Content)
	for _, compName := range components {
		if err := e.checkCircular(compName, visited); err != nil {
			return err
		}
	}
	
	delete(visited, name)
	return nil
}

func (e *Expander) findComponents(node *parser.Node) []string {
	var components []string
	if node == nil {
		return components
	}
	
	if node.Type == parser.NodeElement && len(node.Tag) > 0 && unicode.IsUpper(rune(node.Tag[0])) {
		components = append(components, node.Tag)
	}
	
	for _, child := range node.Children {
		components = append(components, e.findComponents(child)...)
	}
	
	return components
}

func copyAttrs(attrs map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range attrs {
		result[k] = v
	}
	return result
}
