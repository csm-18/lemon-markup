package main

import (
	"strings"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(doc *Document) string {
	var output strings.Builder

	for _, node := range doc.Nodes {
		g.generateNode(node, &output)
	}

	return output.String()
}

func (g *Generator) generateNode(node *Node, output *strings.Builder) {
	if node == nil {
		return
	}

	switch node.Type {
	case NodeText:
		output.WriteString(node.Content)

	case NodeVariable:
		// Variables should have been expanded during the expansion pass,
		// but if any slip through unmapped, we preserve them gracefully.
		output.WriteString("{{ ")
		output.WriteString(node.VarName)
		output.WriteString(" }}")

	case NodeElement:
		// Intercept internal structural wrapper placeholders so they don't pollute the final HTML output
		if node.Tag == "template-result" || node.Tag == "fragment" || node.Tag == "template-content" {
			for _, child := range node.Children {
				g.generateNode(child, output)
			}
		} else {
			g.generateElement(node, output)
		}

	case NodeRaw:
		// Raw blocks simply unpack and stringify their underlying text layers completely untouched
		for _, child := range node.Children {
			g.generateNode(child, output)
		}

	default:
		for _, child := range node.Children {
			g.generateNode(child, output)
		}
	}
}

func (g *Generator) generateElement(node *Node, output *strings.Builder) {
	// Opening structural bracket tag definition marker
	output.WriteString("<")
	output.WriteString(node.Tag)

	// Stringify serialized attribute/property key-value combinations
	for key, value := range node.Attrs {
		output.WriteString(" ")
		output.WriteString(key)
		if value != "" {
			output.WriteString("=\"")
			output.WriteString(value)
			output.WriteString("\"")
		}
	}

	// LEVERAGING PACKAGE SCOPE: reuses the 'voidElements' map directly from parser.go
	if voidElements[node.Tag] {
		output.WriteString(">")
		return
	}

	output.WriteString(">")

	// Render structural children descendants recursively
	for _, child := range node.Children {
		g.generateNode(child, output)
	}

	// Explicit closing bracket element closure string injection
	output.WriteString("</")
	output.WriteString(node.Tag)
	output.WriteString(">")
}
