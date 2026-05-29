package main

import (
	"fmt"
	"os"
	"unicode"
)

type Expander struct {
	registry *Registry
	fileName string
}

type visitedSet map[string]bool

func NewExpander(reg *Registry, fileName string) *Expander {
	return &Expander{
		registry: reg,
		fileName: fileName,
	}
}

// Global Panic-Exit Helper to enforce unified error messaging format
func (e *Expander) fatalError(msg string, line, col int) {
	fmt.Printf("Syntax Error: %s\n", msg)
	fmt.Printf("  --> %s:%d:%d\n", e.fileName, line, col)
	os.Exit(0)
}

func (e *Expander) Expand(doc *Document) *Document {
	// 1. Proactively catch circular graph structures before running any node mutations
	e.checkCircularReferences()

	expanded := &Document{
		Nodes:      make([]*Node, 0),
		Templates:  doc.Templates,
		HasDoctype: doc.HasDoctype,
	}

	for _, node := range doc.Nodes {
		expandedNode := e.expandNode(node, make(visitedSet))
		if expandedNode != nil {
			expanded.Nodes = append(expanded.Nodes, expandedNode)
		}
	}

	return expanded
}

func (e *Expander) expandNode(node *Node, visited visitedSet) *Node {
	if node == nil {
		return nil
	}

	switch node.Type {
	case NodeText:
		return node

	case NodeVariable:
		return node

	case NodeRaw:
		// Raw blocks must skip template or property expansion cycles entirely
		return node

	case NodeElement:
		// Custom components start with an Uppercase character choice
		if len(node.Tag) > 0 && unicode.IsUpper(rune(node.Tag[0])) {
			return e.expandComponent(node, visited)
		}

		// Process native HTML elements recursively down through their children tree branches
		newNode := &Node{
			Type:     node.Type,
			Tag:      node.Tag,
			Attrs:    copyAttrs(node.Attrs),
			Children: make([]*Node, 0),
			Line:     node.Line,
			Col:      node.Col,
		}

		for _, child := range node.Children {
			expanded := e.expandNode(child, visited)
			if expanded != nil {
				newNode.Children = append(newNode.Children, expanded)
			}
		}

		return newNode

	default:
		return node
	}
}

func (e *Expander) expandComponent(node *Node, visited visitedSet) *Node {
	componentName := node.Tag

	// Branch processing check for recursive runtime loops
	if visited[componentName] {
		e.fatalError(fmt.Sprintf("Circular reference branch component loop caught via layout instance invocation: '%s'", componentName), node.Line, node.Col)
	}

	template, exists := e.registry.Get(componentName)
	if !exists {
		e.fatalError(fmt.Sprintf("Undefined custom component layout tag token instance: '%s'. Ensure name registration maps exactly", componentName), node.Line, node.Col)
	}

	// Validate prop structural compatibility signatures
	e.validateProps(node, template)

	// Clone layout path history tracking array maps
	newVisited := make(visitedSet)
	for k, v := range visited {
		newVisited[k] = v
	}
	newVisited[componentName] = true

	// Build and unfold the component instance blueprint layout tree nodes structures safely
	result := e.expandTemplateContent(template.Content, node.Attrs, node.Children, newVisited, node.Line, node.Col)
	return result
}

func (e *Expander) expandTemplateContent(templateContent *Node, props map[string]string, children []*Node, visited visitedSet, line, col int) *Node {
	if len(templateContent.Children) == 0 {
		return &Node{
			Type:     NodeElement,
			Tag:      "div",
			Children: make([]*Node, 0),
			Line:     line,
			Col:      col,
		}
	}

	result := &Node{
		Type:     NodeElement,
		Tag:      "template-result",
		Children: make([]*Node, 0),
		Line:     line,
		Col:      col,
	}

	for _, child := range templateContent.Children {
		expanded := e.expandTemplateNode(child, props, children, visited)
		if expanded != nil {
			result.Children = append(result.Children, expanded)
		}
	}

	// Structural optimization: if a custom component has exactly one root element, don't bundle it under an extra placeholder tag wrapper
	if len(result.Children) == 1 {
		return result.Children[0]
	}

	return result
}

func (e *Expander) expandTemplateNode(node *Node, props map[string]string, children []*Node, visited visitedSet) *Node {
	if node == nil {
		return nil
	}

	switch node.Type {
	case NodeText:
		return node

	case NodeVariable:
		varName := node.VarName

		// Special dynamic target capture placeholder word for element children nodes injection streams
		if varName == "children" {
			if len(children) == 0 {
				return nil
			}
			if len(children) == 1 {
				return e.expandNode(children[0], visited)
			}

			fragment := &Node{
				Type:     NodeElement,
				Tag:      "fragment",
				Children: make([]*Node, 0),
				Line:     node.Line,
				Col:      node.Col,
			}
			for _, child := range children {
				expanded := e.expandNode(child, visited)
				if expanded != nil {
					fragment.Children = append(fragment.Children, expanded)
				}
			}
			return fragment
		}

		// Inject active variable values in place of their template keys
		value, ok := props[varName]
		if !ok {
			e.fatalError(fmt.Sprintf("Missing required property variable string value configuration reference: '%s'", varName), node.Line, node.Col)
		}

		return &Node{
			Type:    NodeText,
			Content: value,
			Line:    node.Line,
			Col:     node.Col,
		}

	case NodeRaw:
		return node

	case NodeElement:
		if len(node.Tag) > 0 && unicode.IsUpper(rune(node.Tag[0])) {
			return e.expandComponent(node, visited)
		}

		newNode := &Node{
			Type:     node.Type,
			Tag:      node.Tag,
			Attrs:    copyAttrs(node.Attrs),
			Children: make([]*Node, 0),
			Line:     node.Line,
			Col:      node.Col,
		}

		for _, child := range node.Children {
			expanded := e.expandTemplateNode(child, props, children, visited)
			if expanded != nil {
				newNode.Children = append(newNode.Children, expanded)
			}
		}

		return newNode

	default:
		return node
	}
}

func (e *Expander) validateProps(componentNode *Node, template *TemplateNode) {
	// Guard 1: Verify missing expected property attributes
	for propName := range template.UsedVars {
		if propName == "children" {
			continue
		}
		if _, ok := componentNode.Attrs[propName]; !ok {
			e.fatalError(fmt.Sprintf("Missing explicitly required prop variable assignment '%s' inside invocation for custom component '%s'", propName, componentNode.Tag), componentNode.Line, componentNode.Col)
		}
	}

	// Guard 2: Catch unrecognized property assignments
	for propName := range componentNode.Attrs {
		if _, ok := template.UsedVars[propName]; !ok {
			e.fatalError(fmt.Sprintf("Unrecognized, unused prop variable property assignment key '%s' passed into custom component element scope '%s'", propName, componentNode.Tag), componentNode.Line, componentNode.Col)
		}
	}
}

func (e *Expander) checkCircularReferences() {
	for name := range e.registry.All() {
		visited := make(visitedSet)
		e.checkCircular(name, visited)
	}
}

func (e *Expander) checkCircular(name string, visited visitedSet) {
	if visited[name] {
		// If caught globally during static analysis before running expand, fallback to structural origin tracking coordinates 0:0
		fmt.Printf("Syntax Error: Deep structural circular validation graph loop caught at component component identity: '%s'\n", name)
		fmt.Printf("  --> Global Graph Cross-Link Error\n")
		os.Exit(0)
	}

	visited[name] = true
	template, _ := e.registry.Get(name)

	components := e.findComponents(template.Content)
	for _, compName := range components {
		e.checkCircular(compName, visited)
	}

	delete(visited, name)
}

func (e *Expander) findComponents(node *Node) []string {
	var components []string
	if node == nil {
		return components
	}

	if node.Type == NodeElement && len(node.Tag) > 0 && unicode.IsUpper(rune(node.Tag[0])) {
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
