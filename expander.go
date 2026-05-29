package main

import (
	"fmt"
	"os"
	"strings"
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

func (e *Expander) fatalError(msg string, line, col int) {
	fmt.Printf("Syntax Error: %s\n", msg)
	fmt.Printf("  --> %s:%d:%d\n", e.fileName, line, col)
	os.Exit(0)
}

func (e *Expander) Expand(doc *Document) *Document {
	e.checkCircularReferences()

	expanded := &Document{
		Nodes:      make([]*Node, 0),
		Templates:  doc.Templates,
		HasDoctype: doc.HasDoctype,
	}

	for _, node := range doc.Nodes {
		expandedNode := e.expandNode(node, make(visitedSet), nil)
		if expandedNode != nil {
			expanded.Nodes = append(expanded.Nodes, expandedNode)
		}
	}

	return expanded
}

func (e *Expander) expandNode(node *Node, visited visitedSet, props map[string]string) *Node {
	if node == nil {
		return nil
	}

	switch node.Type {
	case NodeText:
		return node

	case NodeVariable:
		if val, ok := props[node.VarName]; ok {
			return &Node{
				Type:    NodeText,
				Content: val,
				Line:    node.Line,
				Col:     node.Col,
			}
		}
		return node

	case NodeElement:
		return e.expandTemplateNode(node, visited, props)

	case NodeRaw:
		return node

	default:
		return node
	}
}

func (e *Expander) expandTemplateNode(node *Node, visited visitedSet, props map[string]string) *Node {
	if len(node.Tag) > 0 && unicode.IsUpper(rune(node.Tag[0])) {
		return e.expandComponent(node, visited, props)
	}

	resolvedAttrs := make(map[string]string)
	for k, v := range node.Attrs {
		if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
			varName := strings.TrimSpace(v[2 : len(v)-2])
			if val, ok := props[varName]; ok {
				resolvedAttrs[k] = val
				continue
			}
		}
		resolvedAttrs[k] = v
	}

	newNode := &Node{
		Type:     node.Type,
		Tag:      node.Tag,
		Attrs:    resolvedAttrs,
		Children: make([]*Node, 0),
		Line:     node.Line,
		Col:      node.Col,
	}

	for _, child := range node.Children {
		expandedChild := e.expandNode(child, visited, props)
		if expandedChild != nil {
			newNode.Children = append(newNode.Children, expandedChild)
		}
	}

	return newNode
}

func (e *Expander) expandComponent(componentNode *Node, visited visitedSet, parentProps map[string]string) *Node {
	template, exists := e.registry.Get(componentNode.Tag)
	if !exists {
		e.fatalError(fmt.Sprintf("Reference Error: Custom component symbol identifier definition '%s' could not be found anywhere within global template sheets tracking scopes", componentNode.Tag), componentNode.Line, componentNode.Col)
	}

	componentProps := make(map[string]string)

	for propName, propValue := range componentNode.Attrs {
		if strings.HasPrefix(propValue, "{{") && strings.HasSuffix(propValue, "}}") {
			varName := strings.TrimSpace(propValue[2 : len(propValue)-2])
			if val, ok := parentProps[varName]; ok {
				componentProps[propName] = val
				continue
			}
		}
		componentProps[propName] = propValue
	}

	e.verifyPassedProps(template, componentNode)

	visited[componentNode.Tag] = true
	defer delete(visited, componentNode.Tag)

	wrapperNode := &Node{
		Type:     NodeElement,
		Tag:      "template-result",
		Children: make([]*Node, 0),
		Line:     componentNode.Line,
		Col:      componentNode.Col,
	}

	e.unfoldTemplateContent(template.Content, wrapperNode, componentProps, componentNode.Children, visited, parentProps)

	return wrapperNode
}

func (e *Expander) unfoldTemplateContent(currentNode, targetParent *Node, componentProps map[string]string, callerChildren []*Node, visited visitedSet, parentProps map[string]string) {
	if currentNode == nil {
		return
	}

	switch currentNode.Type {
	case NodeText:
		targetParent.Children = append(targetParent.Children, currentNode)

	case NodeVariable:
		if currentNode.VarName == "children" {
			for _, child := range callerChildren {
				expandedChild := e.expandNode(child, visited, parentProps)
				if expandedChild != nil {
					targetParent.Children = append(targetParent.Children, expandedChild)
				}
			}
		} else {
			if val, ok := componentProps[currentNode.VarName]; ok {
				targetParent.Children = append(targetParent.Children, &Node{
					Type:    NodeText,
					Content: val,
					Line:    currentNode.Line,
					Col:     currentNode.Col,
				})
			} else {
				targetParent.Children = append(targetParent.Children, currentNode)
			}
		}

	case NodeElement:
		var nodeToAppend *Node

		if len(currentNode.Tag) > 0 && unicode.IsUpper(rune(currentNode.Tag[0])) {
			nodeToAppend = e.expandComponent(currentNode, visited, componentProps)
		} else {
			resolvedAttrs := make(map[string]string)
			for k, v := range currentNode.Attrs {
				if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
					varName := strings.TrimSpace(v[2 : len(v)-2])
					if val, ok := componentProps[varName]; ok {
						resolvedAttrs[k] = val
						continue
					}
				}
				resolvedAttrs[k] = v
			}

			nodeToAppend = &Node{
				Type:     currentNode.Type,
				Tag:      currentNode.Tag,
				Attrs:    resolvedAttrs,
				Children: make([]*Node, 0),
				Line:     currentNode.Line,
				Col:      currentNode.Col,
			}

			for _, child := range currentNode.Children {
				e.unfoldTemplateContent(child, nodeToAppend, componentProps, callerChildren, visited, parentProps)
			}
		}

		if nodeToAppend != nil {
			targetParent.Children = append(targetParent.Children, nodeToAppend)
		}

	case NodeRaw:
		targetParent.Children = append(targetParent.Children, currentNode)
	}
}

func (e *Expander) verifyPassedProps(template *TemplateNode, componentNode *Node) {
	for propName := range componentNode.Attrs {
		if _, ok := template.UsedVars[propName]; !ok && propName != "children" {
			e.fatalError(fmt.Sprintf("Specification Error: Unused parameter or property assignment key '%s' passed into custom component element scope '%s'", propName, componentNode.Tag), componentNode.Line, componentNode.Col)
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
		fmt.Printf("Syntax Error: Deep structural circular validation graph loop caught at component identifier: '%s'\n", name)
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
